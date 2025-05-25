package mongo

import "go.mongodb.org/mongo-driver/bson/primitive"

// StreamingNS ...
type StreamingNS struct {
	NS NS `bson:"ns"`
}

// NS ...
type NS struct {
	Db   string `bson:"db"`
	Coll string `bson:"coll"`
}

// ===================================
// Mongo any collection streaming
// ===================================

// StreamType ...
type StreamType struct {
	OperationType string `json:"operationType" bson:"operationType"`
}

// StreamInsert ...
type StreamInsert[T any] struct {
	FullDocument T `json:"fullDocument" bson:"fullDocument"`
}

// StreamUpdate ...
type StreamUpdate[T any] struct {
	DocumentKey       DocumentKey          `json:"documentKey" bson:"documentKey"`
	UpdateDescription UpdateDescription[T] `json:"updateDescription" bson:"updateDescription"`
}

// StreamDelete ...
type StreamDelete struct {
	DocumentKey DocumentKey `json:"documentKey" bson:"documentKey"`
}

// DocumentKey ...
type DocumentKey struct {
	ID primitive.ObjectID `json:"_id" bson:"_id"`
}

// UpdateDescription ...
type UpdateDescription[T any] struct {
	UpdatedFields T        `json:"updatedFields" bson:"updatedFields"`
	RemovedFields []string `json:"removedFields" bson:"removedFields"`
}
