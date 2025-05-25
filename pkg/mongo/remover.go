package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// Remover ...
type Remover struct {
	client            *mongo.Client
	db                string
	collection        string
	connectionTimeout time.Duration
}

// Remove ...
func (s *Remover) Remove(ctx context.Context, doc interface{}) (deletedCount int, err error) {
	collection := s.client.Database(s.db).Collection(s.collection)
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	res, e := collection.DeleteOne(ctx, doc)
	if e == nil {
		deletedCount = int(res.DeletedCount)
		return
	}
	err = e
	return
}

// RemoveMany ...
func (s *Remover) RemoveMany(ctx context.Context, doc interface{}) (deletedCount int, err error) {
	collection := s.client.Database(s.db).Collection(s.collection)
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	res, e := collection.DeleteMany(ctx, doc)
	if e == nil {
		deletedCount = int(res.DeletedCount)
		return
	}
	err = e
	return
}

// NewRemover ...
func NewRemover(client *mongo.Client, db string, collection string, connectionTimeout time.Duration) *Remover {
	return &Remover{
		client:            client,
		db:                db,
		collection:        collection,
		connectionTimeout: connectionTimeout,
	}
}
