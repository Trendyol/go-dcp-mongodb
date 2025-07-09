package main

import (
	"time"

	dcpmongodb "github.com/Trendyol/go-dcp-mongodb"
	config "github.com/Trendyol/go-dcp-mongodb/configs"
	dcpConfig "github.com/Trendyol/go-dcp/config"
)

func main() {
	connector, err := dcpmongodb.NewConnectorBuilder(config.Config{
		MongoDB: config.MongoDB{
			Connection: config.Connection{
				URI:      "localhost:27017",
				Database: "exampleDB",
			},
			CollectionMapping: map[string]string{
				"_default": "exampleCollection",
			},
			Batch: config.BatchConfig{
				TickerDuration:    10 * time.Second,
				SizeLimit:         1000,
				ByteSizeLimit:     "10mb",
				ConcurrentRequest: 1,
			},
			ConnectionPool: config.ConnectionPool{
				MaxPoolSize:   100,
				MinPoolSize:   5,
				MaxIdleTimeMS: 300000, // 5 minutes
			},
			Timeouts: config.Timeouts{
				ConnectTimeoutMS:         10000, // 10 seconds
				ServerSelectionTimeoutMS: 30000, // 30 seconds
				SocketTimeoutMS:          30000, // 30 seconds
				BulkRequestTimeoutMS:     30000, // 30 seconds
			},
			ShardKeys: []string{"customer.id"},
		},
		Dcp: dcpConfig.Dcp{
			Username:   "user",
			Password:   "pass",
			BucketName: "bucketName",
			Hosts:      []string{"http://localhost:8091"},
			Dcp: dcpConfig.ExternalDcp{
				Group: dcpConfig.DCPGroup{
					Name: "groupName",
				},
			},
			Metadata: dcpConfig.Metadata{
				Config: map[string]string{
					"bucket":     "checkpoint-bucket-name",
					"scope":      "_default",
					"collection": "_default",
				},
				Type: "couchbase",
			},
		},
	}).Build()
	if err != nil {
		panic(err)
	}

	defer connector.Close()
	connector.Start()
}
