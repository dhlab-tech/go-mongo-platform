package mongo

import (
	"context"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Searcher provides query and search operations for typed entities in MongoDB.
// It converts MongoDB BSON documents to typed entities using JSON as an intermediate format.
type Searcher[T d] struct {
	client            *mongo.Client
	db                string
	collection        string
	connectionTimeout time.Duration
}

// All retrieves all documents from the collection and returns them as typed entities.
func (s *Searcher[T]) All(ctx context.Context) (items []T, err error) {
	collection := s.client.Database(s.db).Collection(s.collection)
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	cur, e := collection.Find(ctx, bson.M{})
	if e != nil {
		return nil, e
	}
	defer func() {
		_ = cur.Close(ctx)
	}()
	for cur.Next(ctx) {
		var bsonDocument bson.D
		if err = cur.Decode(&bsonDocument); err != nil {
			continue
		}
		var temporaryBytes []byte
		temporaryBytes, err = bson.MarshalExtJSON(bsonDocument, false, false)
		if err != nil {
			continue
		}
		var instance T
		if err = json.Unmarshal(temporaryBytes, &instance); err != nil {
			return
		}
		items = append(items, instance)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return
}

// Find retrieves documents matching the provided filter and returns them as typed entities.
// The filter is a map of field names to values.
func (s *Searcher[T]) Find(ctx context.Context, filter map[string]interface{}) (items []T, err error) {
	collection := s.client.Database(s.db).Collection(s.collection)
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	bsonFilter := bson.M(filter)
	cur, e := collection.Find(ctx, bsonFilter)
	if e != nil {
		return nil, e
	}
	defer func() {
		_ = cur.Close(ctx)
	}()
	for cur.Next(ctx) {
		var bsonDocument bson.D
		if err = cur.Decode(&bsonDocument); err != nil {
			continue
		}
		var temporaryBytes []byte
		temporaryBytes, err = bson.MarshalExtJSON(bsonDocument, false, false)
		if err != nil {
			continue
		}
		var instance T
		if err = json.Unmarshal(temporaryBytes, &instance); err != nil {
			return
		}
		items = append(items, instance)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return
}

// FindOne retrieves a single document matching the provided filter.
// Returns the entity, a boolean indicating if it was found, and any error.
// If no document is found, returns found=false and err=nil (not an error condition).
func (s *Searcher[T]) FindOne(ctx context.Context, filter map[string]interface{}) (item T, found bool, err error) {
	collection := s.client.Database(s.db).Collection(s.collection)
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	bsonFilter := bson.M(filter)
	var bsonDocument bson.D
	err = collection.FindOne(ctx, bsonFilter).Decode(&bsonDocument)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return item, false, nil
		}
		return item, false, err
	}
	var temporaryBytes []byte
	temporaryBytes, err = bson.MarshalExtJSON(bsonDocument, false, false)
	if err != nil {
		return item, false, err
	}
	var instance T
	if err = json.Unmarshal(temporaryBytes, &instance); err != nil {
		return item, false, err
	}
	return instance, true, nil
}

// FindWithFilter retrieves documents matching the provided BSON filter.
// This method provides more flexibility than Find as it accepts complex BSON queries.
func (s *Searcher[T]) FindWithFilter(ctx context.Context, filter bson.M) (items []T, err error) {
	collection := s.client.Database(s.db).Collection(s.collection)
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	cur, e := collection.Find(ctx, filter)
	if e != nil {
		return nil, e
	}
	defer func() {
		_ = cur.Close(ctx)
	}()
	for cur.Next(ctx) {
		var bsonDocument bson.D
		if err = cur.Decode(&bsonDocument); err != nil {
			continue
		}
		var temporaryBytes []byte
		temporaryBytes, err = bson.MarshalExtJSON(bsonDocument, false, false)
		if err != nil {
			continue
		}
		var instance T
		if err = json.Unmarshal(temporaryBytes, &instance); err != nil {
			return
		}
		items = append(items, instance)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return
}

// FindOneWithFilter retrieves a single document matching the provided BSON filter.
// Returns the entity, a boolean indicating if it was found, and any error.
// This method provides more flexibility than FindOne as it accepts complex BSON queries.
func (s *Searcher[T]) FindOneWithFilter(ctx context.Context, filter bson.M) (item T, found bool, err error) {
	collection := s.client.Database(s.db).Collection(s.collection)
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	var bsonDocument bson.D
	err = collection.FindOne(ctx, filter).Decode(&bsonDocument)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return item, false, nil
		}
		return item, false, err
	}
	var temporaryBytes []byte
	temporaryBytes, err = bson.MarshalExtJSON(bsonDocument, false, false)
	if err != nil {
		return item, false, err
	}
	var instance T
	if err = json.Unmarshal(temporaryBytes, &instance); err != nil {
		return item, false, err
	}
	return instance, true, nil
}

// NewSearcher creates a new Searcher instance for a typed entity.
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
