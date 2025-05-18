package db

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	Name            = "news"
	PostsCollection = "posts"
)

func CreateMongoClient(ctx context.Context, uri string) (*mongo.Client, error) {
	logger := log.Ctx(ctx)
	mongoClient, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// retry ping with exponential backoff
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 30 * time.Second
	err = backoff.Retry(func() error {
		pingCtx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		logger.Debug().Msg("trying to connect to mongodb")
		return mongoClient.Ping(pingCtx, nil)
	}, backoff.WithContext(bo, ctx))
	if err != nil {
		return nil, err
	}

	// create weighted full-text index
	coll := mongoClient.Database(Name).Collection(PostsCollection)
	if _, err = coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "title", Value: "text"},
			{Key: "content", Value: "text"},
		},
		Options: options.Index().SetWeights(bson.D{
			{Key: "title", Value: 10},
			{Key: "content", Value: 1},
		}),
	}); err != nil {
		return nil, err
	}

	return mongoClient, nil
}

func GetPostsCollection(mongoClient *mongo.Client) *mongo.Collection {
	return mongoClient.Database(Name).Collection(PostsCollection)
}
