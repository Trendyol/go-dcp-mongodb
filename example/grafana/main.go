package main

import (
	"fmt"
	"log"
	"time"

	dcpmongodb "github.com/Trendyol/go-dcp-mongodb"
	"github.com/couchbase/gocb/v2"
)

func main() {
	time.Sleep(25 * time.Second) //wait for couchbase container initialize

	go seedCouchbaseBucket()

	connector, err := dcpmongodb.NewConnectorBuilder("config.yml").
		Build()
	if err != nil {
		panic(err)
	}

	defer connector.Close()
	connector.Start()
}

func seedCouchbaseBucket() {
	cluster, err := gocb.Connect("couchbase://couchbase", gocb.ClusterOptions{
		Username: "user",
		Password: "password",
	})
	if err != nil {
		log.Fatal(err)
	}

	bucket := cluster.Bucket("dcp-test")
	err = bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection := bucket.DefaultCollection()

	counter := 0
	for {
		counter++
		documentID := fmt.Sprintf("doc-%d", counter)
		document := map[string]interface{}{
			"counter": counter,
			"message": "Hello Couchbase",
			"time":    time.Now().Format(time.RFC3339),
		}
		_, err := collection.Upsert(documentID, document, nil)
		if err != nil {
			log.Println("Error inserting document:", err)
		} else {
			log.Println("Inserted document:", documentID)
		}
		time.Sleep(1 * time.Second)
	}
}
