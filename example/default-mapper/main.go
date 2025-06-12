package main

import (
	dcpmongodb "github.com/Trendyol/go-dcp-mongodb"
)

func main() {
	connector, err := dcpmongodb.NewConnectorBuilder("config.yml").Build()
	if err != nil {
		panic(err)
	}

	defer connector.Close()
	connector.Start()
}
