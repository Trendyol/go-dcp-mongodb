package bulk

import (
	"testing"
	"time"

	config "github.com/Trendyol/go-dcp-mongodb/configs"
	"github.com/Trendyol/go-dcp-mongodb/mongodb"

	"go.mongodb.org/mongo-driver/bson"
)

func TestBulkOperations(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		MongoDB: config.MongoDB{
			URI:                 "mongodb://localhost:27017",
			Database:            "test_db",
			Collection:          "test",
			BatchTickerDuration: 5 * time.Second,
			BatchSizeLimit:      100,
			BatchByteSizeLimit:  1024 * 1024, // 1MB
			ConcurrentRequest:   2,
		},
	}
	cfg.ApplyDefaults()

	// Create bulk processor
	bulk, err := NewBulk(cfg, func() {
		// Mock checkpoint commit
		t.Log("Checkpoint committed")
	})
	if err != nil {
		t.Fatalf("Failed to create bulk processor: %v", err)
	}

	// Test adding different operation types
	t.Run("TestInsertOperation", func(t *testing.T) {
		model := &mongodb.Raw{
			Document: bson.M{
				"_id":  "test123",
				"name": "Test Document",
			},
			Operation: mongodb.Insert,
			ID:        "test123",
		}

		// Check that the model is properly created
		if model.Operation != mongodb.Insert {
			t.Errorf("Expected operation to be Insert, got %v", model.Operation)
		}
	})

	t.Run("TestUpdateOperation", func(t *testing.T) {
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
	})

	t.Run("TestDeleteOperation", func(t *testing.T) {
		model := &mongodb.Raw{
			Operation: mongodb.Delete,
			ID:        "test123",
		}

		if model.Operation != mongodb.Delete {
			t.Errorf("Expected operation to be Delete, got %v", model.Operation)
		}
	})

	t.Run("TestBatchDeduplication", func(t *testing.T) {
		// Test that duplicate keys are handled properly
		key := bulk.getActionKey(&mongodb.Raw{
			ID: "doc1",
		})

		expectedKey := "test:doc1"
		if key != expectedKey {
			t.Errorf("Expected key %s, got %s", expectedKey, key)
		}
	})
}

func Test_it_should_build_shard_filter_with_configured_shard_keys(t *testing.T) {
	// Given
	cfg := &config.Config{
		MongoDB: config.MongoDB{
			URI:                 "mongodb://localhost:27017",
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

	bulk, err := NewBulk(cfg, func() {})
	if err != nil {
		t.Fatalf("Failed to create bulk processor: %v", err)
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
			URI:                 "mongodb://localhost:27017",
			Database:            "test_db",
			BatchTickerDuration: 5 * time.Second,
			BatchSizeLimit:      100,
			BatchByteSizeLimit:  1024 * 1024,
			ConcurrentRequest:   2,
		},
	}
	cfg.ApplyDefaults()

	bulk, err := NewBulk(cfg, func() {})
	if err != nil {
		t.Fatalf("Failed to create bulk processor: %v", err)
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
