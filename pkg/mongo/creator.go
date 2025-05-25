package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Creator ...
type Creator struct {
	client            *mongo.Client
	db                string
	collection        string
	connectionTimeout time.Duration
}

// Create ...
func (s *Creator) Create(ctx context.Context, doc bson.D) (id primitive.ObjectID, err error) {
	var found bool
	for k, d := range doc {
		if d.Key == "version" {
			doc[k].Value = time.Now().UnixNano()
			found = true
			break
		}
	}
	if !found {
		doc = append(doc, bson.E{Key: "version", Value: time.Now().UnixNano()})
	}
	return s.C(ctx, doc)
}

// C ...
func (s *Creator) C(ctx context.Context, doc interface{}) (id primitive.ObjectID, err error) {
	var ok bool
	collection := s.client.Database(s.db).Collection(s.collection)
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	res, e := collection.InsertOne(ctx, doc)
	if e == nil {
		if id, ok = res.InsertedID.(primitive.ObjectID); ok {
			return
		}
		err = fmt.Errorf("cannot convert inserted id to string %s", res.InsertedID)
		return
	}
	err = e
	return
}

// NewCreator ...
func NewCreator(client *mongo.Client, db string, collection string, connectionTimeout time.Duration) *Creator {
	return &Creator{
		client:            client,
		db:                db,
		collection:        collection,
		connectionTimeout: connectionTimeout,
	}
}
