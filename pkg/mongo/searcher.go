package mongo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Searcher ...
type Searcher[T d] struct {
	client            *mongo.Client
	db                string
	collection        string
	connectionTimeout time.Duration
}

// All ...
func (s *Searcher[T]) All(ctx context.Context) (items []T, err error) {
	logger := zerolog.Ctx(ctx)
	collection := s.client.Database(s.db).Collection(s.collection)
	ctx, cancel := context.WithTimeout(context.Background(), s.connectionTimeout)
	defer cancel()
	cur, e := collection.Find(ctx, bson.M{})
	if e == nil {
		defer func() {
			_ = cur.Close(ctx)
		}()
		for cur.Next(ctx) {
			var bsonDocument bson.D
			var temporaryBytes []byte
			err = cur.Decode(&bsonDocument)
			temporaryBytes, err = bson.MarshalExtJSON(bsonDocument, false, false)
			if err != nil {
				continue
			}
			var instance T
			er := json.Unmarshal(temporaryBytes, &instance)
			if er != nil {
				err = er
				return
			}
			logger.Debug().Bytes("instance", temporaryBytes).Msg("searcher:All")
			items = append(items, instance)
		}
		return
	}
	err = e
	return
}

// NewSearcher ...
func NewSearcher[T d](
	client *mongo.Client,
	db string,
	collection string,
	connectionTimeout time.Duration,
) *Searcher[T] {
	return &Searcher[T]{
		client:            client,
		db:                db,
		collection:        collection,
		connectionTimeout: connectionTimeout,
	}
}
