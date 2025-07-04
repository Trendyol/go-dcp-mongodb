package client

import (
	"context"
	"time"

	config "github.com/Trendyol/go-dcp-mongodb/configs"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoClient(cfg config.MongoDB) (*mongo.Client, error) {
	ctx := context.Background()

	clientOpts := options.Client().ApplyURI("mongodb://" + cfg.Connection.URI)
	clientOpts.SetRetryWrites(true)
	clientOpts.SetRetryReads(true)

	if cfg.Connection.Username != "" && cfg.Connection.Password != "" {
		clientOpts.SetAuth(options.Credential{
			Username:   cfg.Connection.Username,
			Password:   cfg.Connection.Password,
			AuthSource: cfg.Connection.Database,
		})
	}

	clientOpts.SetMaxPoolSize(cfg.ConnectionPool.MaxPoolSize)
	clientOpts.SetMinPoolSize(cfg.ConnectionPool.MinPoolSize)
	clientOpts.SetMaxConnIdleTime(time.Duration(cfg.ConnectionPool.MaxIdleTimeMS) * time.Millisecond)

	clientOpts.SetConnectTimeout(time.Duration(cfg.Timeouts.ConnectTimeoutMS) * time.Millisecond)
	clientOpts.SetServerSelectionTimeout(time.Duration(cfg.Timeouts.ServerSelectionTimeoutMS) * time.Millisecond)
	clientOpts.SetSocketTimeout(time.Duration(cfg.Timeouts.SocketTimeoutMS) * time.Millisecond)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()

	if err = client.Ping(pingCtx, nil); err != nil {
		errDisc := client.Disconnect(ctx)
		if errDisc != nil {
			return nil, errDisc
		}

		return nil, err
	}

	return client, nil
}
