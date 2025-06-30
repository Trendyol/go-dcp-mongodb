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
			URI:                 "localhost:27017",
			Database:            "exampleDB",
			Collection:          "exampleCollection",
			BatchTickerDuration: 10 * time.Second,
			ConcurrentRequest:   1,
			ShardKeys:           []string{"customer.id"},
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
