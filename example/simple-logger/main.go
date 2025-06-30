package main

import (
	dcpmongodb "github.com/Trendyol/go-dcp-mongodb"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := createLogger()
	connector, err := dcpmongodb.NewConnectorBuilder("config.yml").
		SetLogger(logger).
		Build()
	if err != nil {
		panic(err)
	}

	defer connector.Close()
	connector.Start()
}

func createLogger() *logrus.Logger {
	logger := logrus.New()

	logger.SetLevel(logrus.ErrorLevel)
	formatter := &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyMsg:   "msg",
			logrus.FieldKeyLevel: "logLevel",
			logrus.FieldKeyTime:  "timestamp",
		},
	}

	logger.SetFormatter(formatter)
	return logger
}
