package bulk

import (
	"context"
	"fmt"
	"strings"

	"github.com/bytedance/sonic"

	"github.com/Trendyol/go-dcp-mongodb/metric"

	config "github.com/Trendyol/go-dcp-mongodb/configs"
	"github.com/Trendyol/go-dcp-mongodb/mongodb"
	"github.com/Trendyol/go-dcp-mongodb/mongodb/client"

	"sync"
	"time"

	"github.com/Trendyol/go-dcp/helpers"
	"github.com/Trendyol/go-dcp/logger"
	"github.com/Trendyol/go-dcp/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"
)

type Bulk struct {
	client              *mongo.Client
	database            *mongo.Database
	collectionMapping   map[string]string
	dcpCheckpointCommit func()
	batchTicker         *time.Ticker
	batchCommitTicker   *time.Ticker
	batchTickerDuration time.Duration
	batchSize           int
	batchSizeLimit      int
	batchByteSizeLimit  int
	batchByteSize       int
	concurrentRequest   int
	batch               []BatchItem
	batchKeys           map[string]int
	batchIndex          int
	flushLock           sync.Mutex
	isDcpRebalancing    bool
	metricsRecorder     mongodb.MetricsRecorder
	shardKeys           []string
}

type BatchItem struct {
	Model mongodb.Model
	Bytes []byte
	Size  int
}

func NewBulk(cfg *config.Config, dcpCheckpointCommit func()) (*Bulk, error) {
	client, err := client.NewMongoClient(cfg.MongoDB)
	if err != nil {
		return nil, err
	}

	var shardKeys []string
	if cfg.MongoDB.ShardKeys != nil {
		shardKeys = cfg.MongoDB.ShardKeys
	}

	batchTickerDuration := cfg.MongoDB.Batch.TickerDuration
	batchSizeLimit := cfg.MongoDB.Batch.SizeLimit
	batchByteSizeLimit := helpers.ResolveUnionIntOrStringValue(cfg.MongoDB.Batch.ByteSizeLimit)
	concurrentRequest := cfg.MongoDB.Batch.ConcurrentRequest

	b := &Bulk{
		client:              client,
		database:            client.Database(cfg.MongoDB.Connection.Database),
		collectionMapping:   cfg.MongoDB.CollectionMapping,
		dcpCheckpointCommit: dcpCheckpointCommit,
		batchTickerDuration: batchTickerDuration,
		batchTicker:         time.NewTicker(batchTickerDuration),
		batchSizeLimit:      batchSizeLimit,
		batchByteSizeLimit:  batchByteSizeLimit,
		concurrentRequest:   concurrentRequest,
		batch:               make([]BatchItem, 0, batchSizeLimit),
		batchKeys:           make(map[string]int, batchSizeLimit),
		shardKeys:           shardKeys,
		metricsRecorder:     metric.NewMetricsRecorder(),
	}

	if batchCommitTickerDuration := cfg.MongoDB.Batch.CommitTickerDuration; batchCommitTickerDuration != nil {
		b.batchCommitTicker = time.NewTicker(*batchCommitTickerDuration)
	}

	return b, nil
}

func (b *Bulk) StartBulk() {
	for range b.batchTicker.C {
		b.flushMessages()
	}
}

func (b *Bulk) Close() {
	b.batchTicker.Stop()
	if b.batchCommitTicker != nil {
		b.batchCommitTicker.Stop()
	}
	b.flushMessages()
}

func (b *Bulk) AddActions(
	ctx *models.ListenerContext,
	eventTime time.Time,
	actions []mongodb.Model,
	couchbaseCollectionName string,
) {
	b.flushLock.Lock()

	if b.isDcpRebalancing {
		logger.Log.Warn("could not add new message to batch while rebalancing")
		b.flushLock.Unlock()
		return
	}

	mongoDBCollectionName := b.getCollectionName(couchbaseCollectionName)

	for _, action := range actions {
		if rawModel, ok := action.(*mongodb.Raw); ok {
			rawModel.MongoCollection = mongoDBCollectionName
		}

		bytes, err := sonic.Marshal(action)
		if err != nil {
			logger.Log.Error("error marshaling action: %v", err)
			continue
		}
		size := len(bytes)

		key := b.getActionKey(action)

		if batchIndex, ok := b.batchKeys[key]; ok {
			b.batchByteSize += size - b.batch[batchIndex].Size
			b.batch[batchIndex] = BatchItem{
				Model: action,
				Bytes: bytes,
				Size:  size,
			}
		} else {
			b.batch = append(b.batch, BatchItem{
				Model: action,
				Bytes: bytes,
				Size:  size,
			})
			b.batchKeys[key] = b.batchIndex
			b.batchIndex++
			b.batchSize++
			b.batchByteSize += size
		}
	}

	ctx.Ack()
	b.flushLock.Unlock()

	b.metricsRecorder.RecordProcessLatency(time.Since(eventTime).Milliseconds())

	if b.batchSize >= b.batchSizeLimit || b.batchByteSize >= b.batchByteSizeLimit {
		b.flushMessages()
	}
}

func (b *Bulk) getCollectionName(couchbaseCollectionName string) string {
	if mongoCollectionName, exists := b.collectionMapping[couchbaseCollectionName]; exists {
		return mongoCollectionName
	}

	logger.Log.Error("there is no collection mapping for couchbase collection: %s", couchbaseCollectionName)
	panic(fmt.Errorf("there is no collection mapping for couchbase collection: %s", couchbaseCollectionName))
}

func (b *Bulk) getActionKey(model mongodb.Model) string {
	if rawModel, ok := model.(*mongodb.Raw); ok {
		mongoCollection := rawModel.MongoCollection

		if rawModel.ID != "" {
			return fmt.Sprintf("%s:%s", mongoCollection, rawModel.ID)
		}

		if id, ok := rawModel.Document["_id"]; ok {
			return fmt.Sprintf("%s:%v", mongoCollection, id)
		}
	}

	return fmt.Sprintf("batch:%d", b.batchIndex)
}

func (b *Bulk) flushMessages() {
	b.flushLock.Lock()
	defer b.flushLock.Unlock()

	if b.isDcpRebalancing {
		return
	}

	if len(b.batch) > 0 {
		err := b.bulkRequest()
		if err != nil {
			logger.Log.Error("error while bulk request: %v", err)
			panic(err)
		}

		b.batchTicker.Reset(b.batchTickerDuration)

		b.batch = b.batch[:0]
		b.batchKeys = make(map[string]int, b.batchSizeLimit)
		b.batchIndex = 0
		b.batchSize = 0
		b.batchByteSize = 0
	}

	b.checkAndCommit()
}

func (b *Bulk) bulkRequest() error {
	eg, _ := errgroup.WithContext(context.Background())

	chunks := helpers.ChunkSlice(b.batch, b.concurrentRequest)

	startedTime := time.Now()

	for i := range chunks {
		if len(chunks[i]) > 0 {
			eg.Go(b.processBatchChunk(chunks[i]))
		}
	}

	err := eg.Wait()

	b.metricsRecorder.RecordBulkRequestProcessLatency(time.Since(startedTime).Milliseconds())

	return err
}

func (b *Bulk) processBatchChunk(batchItems []BatchItem) func() error {
	return func() error {
		operations := make(map[string][]mongo.WriteModel)

		for _, item := range batchItems {
			model := item.Model
			if rawModel, ok := model.(*mongodb.Raw); ok {
				collection := rawModel.MongoCollection

				var writeModel mongo.WriteModel
				switch rawModel.Operation {
				case mongodb.Insert, mongodb.Update, mongodb.Upsert:
					writeModel = mongo.NewReplaceOneModel().
						SetFilter(b.buildFilter(rawModel.Document)).
						SetReplacement(rawModel.Document).
						SetUpsert(true)
				case mongodb.Delete:
					writeModel = mongo.NewDeleteOneModel().SetFilter(b.buildFilter(rawModel.Document))
				default:
					writeModel = mongo.NewInsertOneModel().SetDocument(rawModel.Document)
				}

				operations[collection] = append(operations[collection], writeModel)
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		for collectionName, writeModels := range operations {
			collection := b.database.Collection(collectionName)

			opts := options.BulkWrite().SetOrdered(false)
			result, err := collection.BulkWrite(ctx, writeModels, opts)

			if err != nil {
				if mongoErr, ok := err.(mongo.BulkWriteException); ok {
					for _, writeErr := range mongoErr.WriteErrors {
						if writeErr.Code == 11000 {
							logger.Log.Error("Duplicate key error: %v\n", err)
						}
					}
					continue
				}
				b.recordErrors(collectionName, writeModels)
				return fmt.Errorf("bulk write error for collection %s: %v", collectionName, err)
			}

			b.recordSuccess(collectionName, result)
		}

		return nil
	}
}

func (b *Bulk) buildFilter(document map[string]interface{}) bson.M {
	filter := bson.M{"_id": document["_id"]}

	for _, shardKey := range b.shardKeys {
		value := b.getNestedValue(document, shardKey)
		if value != nil {
			filter[shardKey] = value
		}
	}

	return filter
}

func (b *Bulk) getNestedValue(document map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := document

	for _, part := range parts {
		if current == nil {
			return nil
		}

		if value, ok := current[part]; ok {
			if nextMap, isMap := value.(map[string]interface{}); isMap {
				current = nextMap
			} else {
				return value
			}
		} else {
			return nil
		}
	}

	return current
}

func (b *Bulk) recordErrors(collection string, operations []mongo.WriteModel) {
	for _, op := range operations {
		switch op.(type) {
		case *mongo.UpdateOneModel, *mongo.UpdateManyModel:
			b.metricsRecorder.RecordUpdateError(collection, 1)
		case *mongo.DeleteOneModel, *mongo.DeleteManyModel:
			b.metricsRecorder.RecordDeleteError(collection, 1)
		}
	}
}

func (b *Bulk) recordSuccess(collection string, result *mongo.BulkWriteResult) {
	b.metricsRecorder.RecordUpdateSuccess(collection, result.ModifiedCount+result.UpsertedCount)
	b.metricsRecorder.RecordDeleteSuccess(collection, result.DeletedCount)
}

func (b *Bulk) checkAndCommit() {
	if b.batchCommitTicker == nil {
		b.dcpCheckpointCommit()
		return
	}

	select {
	case <-b.batchCommitTicker.C:
		b.dcpCheckpointCommit()
	default:
		return
	}
}
