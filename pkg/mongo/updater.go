package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	boolFalse = false
)

// Updater ...
type Updater struct {
	client            *mongo.Client
	db                string
	collection        string
	connectionTimeout time.Duration
}

// UpdateOne ...
// Алгоритм записи данных в монго
// 1. Взяли данные с версией, например 1
// 2. Подготовили изменение в данных
// 3. Вытаемся обновить данные с версией 1
// 4. Если не получилось (версия ушла дальше), то берем снова данные с новой версией, например 3, и пытаемся их записать
func (s *Updater) UpdateOne(ctx context.Context, id string, version *int64, set bson.D, unset bson.D) (found bool, err error) {
	if set == nil && unset == nil {
		return
	}
	collection := s.client.Database(s.db).Collection(s.collection)
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return
	}
	filter := bson.D{{Key: "_id", Value: _id}}
	for k, d := range set {
		if d.Key == "version" {
			set[k].Value = time.Now().UnixNano()
			found = true
			break
		}
	}
	if !found {
		set = append(set, bson.E{Key: "version", Value: time.Now().UnixNano()})
	}
	update := bson.D{bson.E{Key: "$set", Value: set}}
	if version != nil {
		filter = append(filter, bson.E{Key: "version", Value: *version})
	}
	if unset != nil {
		update = append(update, bson.E{Key: "$unset", Value: unset})
	}
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	opts := options.Update()
	opts.Upsert = &boolFalse
	result, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return
	}
	found = result.MatchedCount == 1
	return
}

// NewUpdater ...
func NewUpdater(client *mongo.Client, db string, collection string, connectionTimeout time.Duration) *Updater {
	return &Updater{
		client:            client,
		db:                db,
		collection:        collection,
		connectionTimeout: connectionTimeout,
	}
}
