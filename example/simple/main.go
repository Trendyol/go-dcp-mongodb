package main

import (
	dcpmongodb "github.com/Trendyol/go-dcp-mongodb"
	"github.com/Trendyol/go-dcp-mongodb/couchbase"
	"github.com/Trendyol/go-dcp-mongodb/mongodb"
	"github.com/Trendyol/go-dcp/logger"
	"go.mongodb.org/mongo-driver/bson"
)

func mapper(event couchbase.Event) []mongodb.Model {
	docID := string(event.Key)

	valueMap, err := parseEventValue(event.Value)
	if err != nil {
		logger.Log.Error("Failed to parse document - Key: %s, Value: %s, Error: %v", docID, string(event.Value), err)
	}

	valueMap["_id"] = docID
	valueMap["docVersion"] = 1

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

func main() {
	connector, err := dcpmongodb.NewConnectorBuilder("config.yml").
		SetMapper(func(e couchbase.Event) []mongodb.Model {
			return mapper(e)
		}).
		Build()
	if err != nil {
		panic(err)
	}

	defer connector.Close()
	connector.Start()
}
