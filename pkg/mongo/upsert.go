package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Upsert ...
type Upsert struct {
	client            *mongo.Client
	db                string
	collection        string
	connectionTimeout time.Duration
}

// UpsertOne ...
func (s *Upsert) UpsertOne(ctx context.Context, id string, doc interface{}) (err error) {
	collection := s.client.Database(s.db).Collection(s.collection)
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	var wc []mongo.WriteModel
	uom := mongo.NewUpdateOneModel()
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return
	}
	uom.SetFilter(bson.D{{Key: "_id", Value: _id}})
	uom.SetUpdate(bson.M{"$set": doc})
	uom.SetUpsert(true)
	wc = append(wc, uom)
	_, err = collection.BulkWrite(ctx, wc)
	return
}

// UpsertMany ...
func (s *Upsert) UpsertMany(ctx context.Context, ids []string, set []interface{}) (err error) {
	collection := s.client.Database(s.db).Collection(s.collection)
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	var wc []mongo.WriteModel
	for i := 0; i < len(set); i++ {
		uom := mongo.NewUpdateOneModel()
		_id, er := primitive.ObjectIDFromHex(ids[i])
		if er != nil {
			return er
		}
		uom.SetFilter(bson.D{{Key: "_id", Value: _id}})
		uom.SetUpdate(bson.M{"$set": set[i]})
		uom.SetUpsert(true)
		wc = append(wc, uom)
	}
	_, err = collection.BulkWrite(ctx, wc)
	return
}

// NewUpsert ...
func NewUpsert(client *mongo.Client, db string, collection string, connectionTimeout time.Duration) *Upsert {
	return &Upsert{
		client:            client,
		db:                db,
		collection:        collection,
		connectionTimeout: connectionTimeout,
	}
}
