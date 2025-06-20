package bulk

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	dbName              string
	collectionName      string
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
	metric              *Metric
	metricCounterMutex  sync.Mutex
	shardKeys           []string
}

type BatchItem struct {
	Model mongodb.Model
	Bytes []byte
	Size  int
}

type Metric struct {
	InsertErrorCounter          map[string]int64
	UpdateSuccessCounter        map[string]int64
	UpdateErrorCounter          map[string]int64
	DeleteSuccessCounter        map[string]int64
	DeleteErrorCounter          map[string]int64
	ProcessLatencyMs            int64
	BulkRequestProcessLatencyMs int64
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

	b := &Bulk{
		client:              client,
		dbName:              cfg.MongoDB.Database,
		collectionName:      cfg.MongoDB.Collection,
		dcpCheckpointCommit: dcpCheckpointCommit,
		batchTickerDuration: cfg.MongoDB.BatchTickerDuration,
		batchTicker:         time.NewTicker(cfg.MongoDB.BatchTickerDuration),
		batchSizeLimit:      cfg.MongoDB.BatchSizeLimit,
		batchByteSizeLimit:  helpers.ResolveUnionIntOrStringValue(cfg.MongoDB.BatchByteSizeLimit),
		concurrentRequest:   cfg.MongoDB.ConcurrentRequest,
		batch:               make([]BatchItem, 0, cfg.MongoDB.BatchSizeLimit),
		batchKeys:           make(map[string]int, cfg.MongoDB.BatchSizeLimit),
		shardKeys:           shardKeys,
		metric: &Metric{
			InsertErrorCounter:   make(map[string]int64),
			UpdateSuccessCounter: make(map[string]int64),
			UpdateErrorCounter:   make(map[string]int64),
			DeleteSuccessCounter: make(map[string]int64),
			DeleteErrorCounter:   make(map[string]int64),
		},
	}

	if cfg.MongoDB.BatchCommitTickerDuration != nil {
		b.batchCommitTicker = time.NewTicker(*cfg.MongoDB.BatchCommitTickerDuration)
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

func (b *Bulk) AddActions(ctx *models.ListenerContext, eventTime time.Time, actions []mongodb.Model) {
	b.flushLock.Lock()
	if b.isDcpRebalancing {
		logger.Log.Warn("could not add new message to batch while rebalancing")
		b.flushLock.Unlock()
		return
	}

	for _, action := range actions {
		bytes, err := json.Marshal(action)
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

	b.metric.ProcessLatencyMs = time.Since(eventTime).Milliseconds()

	if b.batchSize >= b.batchSizeLimit || b.batchByteSize >= b.batchByteSizeLimit {
		b.flushMessages()
	}
}

func (b *Bulk) getActionKey(model mongodb.Model) string {
	if rawModel, ok := model.(*mongodb.Raw); ok {
		if rawModel.ID != "" {
			return fmt.Sprintf("%s:%s", b.collectionName, rawModel.ID)
		}

		if id, ok := rawModel.Document["_id"]; ok {
			return fmt.Sprintf("%s:%v", b.collectionName, id)
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

	b.metric.BulkRequestProcessLatencyMs = time.Since(startedTime).Milliseconds()

	return err
}

func (b *Bulk) processBatchChunk(batchItems []BatchItem) func() error {
	return func() error {
		operations := make(map[string][]mongo.WriteModel)

		for _, item := range batchItems {
			model := item.Model
			if rawModel, ok := model.(*mongodb.Raw); ok {
				collection := b.collectionName

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
			collection := b.client.Database(b.dbName).Collection(collectionName)

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
				b.recordErrors(collectionName, writeModels, err)
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

func (b *Bulk) recordErrors(collection string, operations []mongo.WriteModel, err error) {
	b.metricCounterMutex.Lock()
	defer b.metricCounterMutex.Unlock()

	for _, op := range operations {
		switch op.(type) {
		case *mongo.InsertOneModel:
			b.metric.InsertErrorCounter[collection]++
		case *mongo.UpdateOneModel, *mongo.UpdateManyModel:
			b.metric.UpdateErrorCounter[collection]++
		case *mongo.DeleteOneModel, *mongo.DeleteManyModel:
			b.metric.DeleteErrorCounter[collection]++
		}
	}
}

func (b *Bulk) recordSuccess(collection string, result *mongo.BulkWriteResult) {
	b.metricCounterMutex.Lock()
	defer b.metricCounterMutex.Unlock()

	b.metric.UpdateSuccessCounter[collection] += result.ModifiedCount + result.UpsertedCount
	b.metric.DeleteSuccessCounter[collection] += result.DeletedCount
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

func (b *Bulk) GetMetric() *Metric {
	return b.metric
}

func (b *Bulk) LockMetrics() {
	b.metricCounterMutex.Lock()
}

func (b *Bulk) UnlockMetrics() {
	b.metricCounterMutex.Unlock()
}
