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

	clientOpts := options.Client().ApplyURI("mongodb://" + cfg.URI)
	clientOpts.SetRetryWrites(true)
	clientOpts.SetRetryReads(true)

	if cfg.Username != "" && cfg.Password != "" {
		clientOpts.SetAuth(options.Credential{
			Username:   cfg.Username,
			Password:   cfg.Password,
			AuthSource: cfg.Database,
		})
	}

	clientOpts.SetMaxPoolSize(cfg.MaxPoolSize)
	clientOpts.SetMinPoolSize(cfg.MinPoolSize)
	clientOpts.SetMaxConnIdleTime(time.Duration(cfg.MaxIdleTimeMS) * time.Millisecond)

	clientOpts.SetConnectTimeout(time.Duration(cfg.ConnectTimeoutMS) * time.Millisecond)
	clientOpts.SetServerSelectionTimeout(time.Duration(cfg.ServerSelectionTimeoutMS) * time.Millisecond)
	clientOpts.SetSocketTimeout(time.Duration(cfg.SocketTimeoutMS) * time.Millisecond)

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
