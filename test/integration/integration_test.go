package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	dcpmongodb "github.com/Trendyol/go-dcp-mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoDB(t *testing.T) {
	time.Sleep(time.Minute)

	connector, err := dcpmongodb.NewConnectorBuilder("config.yml").Build()
	if err != nil {
		t.Fatal(err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		connector.Start()
	}()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		clientOpts := options.Client().ApplyURI("mongodb://localhost:27017")
		client, err := mongo.Connect(ctx, clientOpts)
		if err != nil {
			t.Fatalf("Failed to connect mongodb: %s", err)
		}
		defer client.Disconnect(ctx)

		collection := client.Database("dcp-test").Collection("test")

		testCtx, testCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer testCancel()

	CountCheckLoop:
		for {
			select {
			case <-testCtx.Done():
				t.Fatalf("deadline exceed")
			default:
				count, err := collection.CountDocuments(context.Background(), bson.M{})
				if err != nil {
					t.Fatalf("could not get count from mongodb: %s", err)
				}

				t.Logf("Document count: %d", count)

				if count == 31591 {
					connector.Close()
					break CountCheckLoop
				}
				time.Sleep(2 * time.Second)
			}
		}
	}()

	wg.Wait()
}
