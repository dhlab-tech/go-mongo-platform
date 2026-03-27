package mongo

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Searcher provides query and search operations for typed entities in MongoDB.
// It converts MongoDB BSON documents to typed entities using JSON as an intermediate format.
type Searcher[T d] struct {
	client            *mongo.Client
	db                string
	collection        string
	connectionTimeout time.Duration
}

func findOpts() *options.FindOptions {
	return options.Find().SetBatchSize(2000)
}

// decodeCursorDoc decodes a cursor document into T. Pointer-to-struct entities (common in this codebase)
// use bson.Unmarshal directly; the legacy JSON path is kept as a fallback for unusual types.
func decodeCursorDoc[T d](cur *mongo.Cursor) (instance T, err error) {
	var zero T
	typ := reflect.TypeOf(zero)
	if typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Struct {
		ptr := reflect.New(typ.Elem())
		instance = ptr.Interface().(T)
		err = cur.Decode(instance)
		return instance, err
	}
	if typ.Kind() == reflect.Struct {
		err = cur.Decode(&instance)
		return instance, err
	}
	var bsonDocument bson.D
	if err = cur.Decode(&bsonDocument); err != nil {
		return instance, err
	}
	var temporaryBytes []byte
	temporaryBytes, err = bson.MarshalExtJSON(bsonDocument, false, false)
	if err != nil {
		return instance, err
	}
	err = json.Unmarshal(temporaryBytes, &instance)
	return instance, err
}

func decodeSingleDoc[T d](raw bson.Raw) (instance T, err error) {
	var zero T
	typ := reflect.TypeOf(zero)
	if typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Struct {
		ptr := reflect.New(typ.Elem())
		instance = ptr.Interface().(T)
		err = bson.Unmarshal(raw, instance)
		return instance, err
	}
	if typ.Kind() == reflect.Struct {
		err = bson.Unmarshal(raw, &instance)
		return instance, err
	}
	var bsonDocument bson.D
	if err = bson.Unmarshal(raw, &bsonDocument); err != nil {
		return instance, err
	}
	var temporaryBytes []byte
	temporaryBytes, err = bson.MarshalExtJSON(bsonDocument, false, false)
	if err != nil {
		return instance, err
	}
	err = json.Unmarshal(temporaryBytes, &instance)
	return instance, err
}

// All retrieves all documents from the collection and returns them as typed entities.
func (s *Searcher[T]) All(ctx context.Context) (items []T, err error) {
	collection := s.client.Database(s.db).Collection(s.collection)
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	cur, e := collection.Find(ctx, bson.M{}, findOpts())
	if e != nil {
		return nil, e
	}
	defer func() {
		_ = cur.Close(ctx)
	}()
	for cur.Next(ctx) {
		var instance T
		instance, err = decodeCursorDoc[T](cur)
		if err != nil {
			continue
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
	cur, e := collection.Find(ctx, bsonFilter, findOpts())
	if e != nil {
		return nil, e
	}
	defer func() {
		_ = cur.Close(ctx)
	}()
	for cur.Next(ctx) {
		var instance T
		instance, err = decodeCursorDoc[T](cur)
		if err != nil {
			continue
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
	var raw bson.Raw
	err = collection.FindOne(ctx, bsonFilter).Decode(&raw)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return item, false, nil
		}
		return item, false, err
	}
	instance, err := decodeSingleDoc[T](raw)
	if err != nil {
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
	cur, e := collection.Find(ctx, filter, findOpts())
	if e != nil {
		return nil, e
	}
	defer func() {
		_ = cur.Close(ctx)
	}()
	for cur.Next(ctx) {
		var instance T
		instance, err = decodeCursorDoc[T](cur)
		if err != nil {
			continue
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
	var raw bson.Raw
	err = collection.FindOne(ctx, filter).Decode(&raw)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return item, false, nil
		}
		return item, false, err
	}
	instance, err := decodeSingleDoc[T](raw)
	if err != nil {
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
