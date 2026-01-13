package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Creator handles document creation operations in MongoDB.
// It automatically sets the version field to the current timestamp (UnixNano)
// if not already present in the document.
type Creator struct {
	client            *mongo.Client
	db                string
	collection        string
	connectionTimeout time.Duration
}

// Create inserts a new document into MongoDB.
// It automatically sets the version field to the current timestamp (UnixNano) if not present.
// Returns the inserted document's ObjectID.
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

// C inserts a document into MongoDB without automatic version field handling.
// This is a lower-level method that directly inserts the provided document.
// Use Create for automatic version field management.
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

// NewCreator creates a new Creator instance for the specified database and collection.
// The connectionTimeout is used for all MongoDB operations performed by this Creator.
func NewCreator(client *mongo.Client, db string, collection string, connectionTimeout time.Duration) *Creator {
	return &Creator{
		client:            client,
		db:                db,
		collection:        collection,
		connectionTimeout: connectionTimeout,
	}
}
