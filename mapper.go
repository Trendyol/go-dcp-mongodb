package dcpmongodb

import (
	"github.com/Trendyol/go-dcp-mongodb/couchbase"
	"github.com/Trendyol/go-dcp-mongodb/mongodb"
	"github.com/Trendyol/go-dcp/logger"
	"go.mongodb.org/mongo-driver/bson"
)

type Mapper func(event couchbase.Event) []mongodb.Model

func DefaultMapper(event couchbase.Event) []mongodb.Model {
	docID := string(event.Key)

	valueMap, err := parseEventValue(event.Value)
	if err != nil {
		logger.Log.Error("Failed to parse document - Key: %s, Value: %s, Error: %v", docID, string(event.Value), err)
	}

	valueMap["_id"] = docID

	operation := determineOperation(event)

	model := &mongodb.Raw{
		Document:  valueMap,
		Operation: operation,
		ID:        docID,
	}

	if operation == mongodb.Delete {
		model.Document = map[string]interface{}{"_id": docID}
	}

	return []mongodb.Model{model}
}

func parseEventValue(eventValue []byte) (map[string]interface{}, error) {
	valueMap := make(map[string]interface{})

	if eventValue == nil {
		return valueMap, nil
	}

	err := bson.UnmarshalExtJSON(eventValue, false, &valueMap)
	return valueMap, err
}

func determineOperation(event couchbase.Event) mongodb.OperationType {
	switch {
	case event.IsDeleted || event.IsExpired:
		return mongodb.Delete
	case event.IsMutated:
		return mongodb.Upsert
	default:
		return mongodb.Insert
	}
}
