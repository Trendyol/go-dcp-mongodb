package bulk

import (
	"testing"
	"time"

	config "github.com/Trendyol/go-dcp-mongodb/configs"
	"github.com/Trendyol/go-dcp-mongodb/mongodb"

	"go.mongodb.org/mongo-driver/bson"
)

func Test_it_should_handle_bulk_operations(t *testing.T) {
	t.Run("it_should_handle_insert_operation", func(t *testing.T) {
		testItShouldHandleInsertOperation(t)
	})

	t.Run("it_should_handle_update_operation", func(t *testing.T) {
		testItShouldHandleUpdateOperation(t)
	})

	t.Run("it_should_handle_delete_operation", func(t *testing.T) {
		testItShouldHandleDeleteOperation(t)
	})

	t.Run("it_should_handle_batch_deduplication", func(t *testing.T) {
		testItShouldHandleBatchDeduplication(t)
	})
}

func createTestBulkWithoutConnection(t *testing.T) *Bulk {
	cfg := &config.Config{
		MongoDB: config.MongoDB{
			URI:                 "localhost:27017",
			Database:            "test_db",
			Collection:          "test",
			BatchTickerDuration: 5 * time.Second,
			BatchSizeLimit:      100,
			BatchByteSizeLimit:  1024 * 1024, // 1MB
			ConcurrentRequest:   2,
		},
	}
	cfg.ApplyDefaults()

	bulk := &Bulk{
		client:              nil,
		database:            nil,
		collectionName:      cfg.MongoDB.Collection,
		dcpCheckpointCommit: func() { t.Log("Checkpoint committed") },
		batchTickerDuration: cfg.MongoDB.BatchTickerDuration,
		batchTicker:         time.NewTicker(cfg.MongoDB.BatchTickerDuration),
		batchSizeLimit:      cfg.MongoDB.BatchSizeLimit,
		batchByteSizeLimit:  1024 * 1024,
		concurrentRequest:   cfg.MongoDB.ConcurrentRequest,
		batch:               make([]BatchItem, 0, cfg.MongoDB.BatchSizeLimit),
		batchKeys:           make(map[string]int, cfg.MongoDB.BatchSizeLimit),
		shardKeys:           cfg.MongoDB.ShardKeys,
		metric: &Metric{
			InsertErrorCounter:   make(map[string]int64),
			UpdateSuccessCounter: make(map[string]int64),
			UpdateErrorCounter:   make(map[string]int64),
			DeleteSuccessCounter: make(map[string]int64),
			DeleteErrorCounter:   make(map[string]int64),
		},
	}

	return bulk
}

func testItShouldHandleInsertOperation(t *testing.T) {
	model := &mongodb.Raw{
		Document: bson.M{
			"_id":  "test123",
			"name": "Test Document",
		},
		Operation: mongodb.Insert,
		ID:        "test123",
	}

	if model.Operation != mongodb.Insert {
		t.Errorf("Expected operation to be Insert, got %v", model.Operation)
	}
}

func testItShouldHandleUpdateOperation(t *testing.T) {
	model := &mongodb.Raw{
		Document: bson.M{
			"name": "Updated Document",
		},
		Operation: mongodb.Update,
		ID:        "test123",
	}

	if model.Operation != mongodb.Update {
		t.Errorf("Expected operation to be Update, got %v", model.Operation)
	}
}

func testItShouldHandleDeleteOperation(t *testing.T) {
	model := &mongodb.Raw{
		Operation: mongodb.Delete,
		ID:        "test123",
	}

	if model.Operation != mongodb.Delete {
		t.Errorf("Expected operation to be Delete, got %v", model.Operation)
	}
}

func testItShouldHandleBatchDeduplication(t *testing.T) {
	bulk := createTestBulkWithoutConnection(t)

	key := bulk.getActionKey(&mongodb.Raw{
		ID: "doc1",
	})

	expectedKey := "test:doc1"
	if key != expectedKey {
		t.Errorf("Expected key %s, got %s", expectedKey, key)
	}
}

func Test_it_should_build_shard_filter_with_configured_shard_keys(t *testing.T) {
	// Given
	cfg := &config.Config{
		MongoDB: config.MongoDB{
			URI:                 "localhost:27017",
			Database:            "test_db",
			BatchTickerDuration: 5 * time.Second,
			BatchSizeLimit:      100,
			BatchByteSizeLimit:  1024 * 1024,
			ConcurrentRequest:   2,
			ShardKeys: []string{
				"customer.id",
				"tenant.id",
			},
		},
	}
	cfg.ApplyDefaults()

	bulk := &Bulk{
		client:         nil,
		database:       nil,
		collectionName: "test",
		shardKeys:      cfg.MongoDB.ShardKeys,
	}

	document := map[string]interface{}{
		"_id": "test123",
		"customer": map[string]interface{}{
			"id": "customer123",
		},
		"tenant": map[string]interface{}{
			"id": "tenant456",
		},
	}

	// When
	filter := bulk.buildFilter(document)

	// Then
	expectedFilter := bson.M{
		"_id":         "test123",
		"customer.id": "customer123",
		"tenant.id":   "tenant456",
	}

	if len(filter) != len(expectedFilter) {
		t.Errorf("Expected filter length %d, got %d", len(expectedFilter), len(filter))
	}

	for key, expectedValue := range expectedFilter {
		if actualValue, exists := filter[key]; !exists || actualValue != expectedValue {
			t.Errorf("Expected filter[%s] = %v, got %v", key, expectedValue, actualValue)
		}
	}
}

func Test_it_should_build_filter_with_only_id_when_no_shard_keys_configured(t *testing.T) {
	// Given
	cfg := &config.Config{
		MongoDB: config.MongoDB{
			URI:                 "localhost:27017",
			Database:            "test_db",
			BatchTickerDuration: 5 * time.Second,
			BatchSizeLimit:      100,
			BatchByteSizeLimit:  1024 * 1024,
			ConcurrentRequest:   2,
		},
	}
	cfg.ApplyDefaults()

	bulk := &Bulk{
		client:         nil,
		database:       nil,
		collectionName: "test",
		shardKeys:      nil,
	}

	document := map[string]interface{}{
		"_id": "test123",
		"customer": map[string]interface{}{
			"id": "customer123",
		},
	}

	// When
	filter := bulk.buildFilter(document)

	// Then
	expectedFilter := bson.M{"_id": "test123"}

	if len(filter) != len(expectedFilter) {
		t.Errorf("Expected filter length %d, got %d", len(expectedFilter), len(filter))
	}

	if filter["_id"] != "test123" {
		t.Errorf("Expected filter['_id'] = 'test123', got %v", filter["_id"])
	}
}

func Test_getNestedValue_should_return_correct_nested_value(t *testing.T) {
	bulk := &Bulk{}

	document := map[string]interface{}{
		"customer": map[string]interface{}{
			"id": "customer123",
			"profile": map[string]interface{}{
				"name": "John Doe",
			},
		},
		"tenant": map[string]interface{}{
			"id": "tenant456",
		},
	}

	value := bulk.getNestedValue(document, "customer.id")
	if value != "customer123" {
		t.Errorf("Expected 'customer123', got %v", value)
	}

	value = bulk.getNestedValue(document, "customer.profile.name")
	if value != "John Doe" {
		t.Errorf("Expected 'John Doe', got %v", value)
	}

	value = bulk.getNestedValue(document, "nonexistent.path")
	if value != nil {
		t.Errorf("Expected nil, got %v", value)
	}

	value = bulk.getNestedValue(document, "tenant.id")
	if value != "tenant456" {
		t.Errorf("Expected 'tenant456', got %v", value)
	}
}

func Test_getActionKey_should_return_correct_key(t *testing.T) {
	bulk := &Bulk{
		collectionName: "test_collection",
		batchIndex:     5,
	}

	model := &mongodb.Raw{
		ID: "test123",
		Document: map[string]interface{}{
			"_id": "test123",
		},
	}

	key := bulk.getActionKey(model)
	expectedKey := "test_collection:test123"
	if key != expectedKey {
		t.Errorf("Expected key %s, got %s", expectedKey, key)
	}

	model = &mongodb.Raw{
		ID: "",
		Document: map[string]interface{}{
			"_id": "doc456",
		},
	}

	key = bulk.getActionKey(model)
	expectedKey = "test_collection:doc456"
	if key != expectedKey {
		t.Errorf("Expected key %s, got %s", expectedKey, key)
	}

	model = &mongodb.Raw{
		ID: "",
		Document: map[string]interface{}{
			"name": "test",
		},
	}

	key = bulk.getActionKey(model)
	expectedKey = "batch:5"
	if key != expectedKey {
		t.Errorf("Expected key %s, got %s", expectedKey, key)
	}
}
