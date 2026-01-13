package mongo

import "go.mongodb.org/mongo-driver/bson/primitive"

// StreamingNS represents the namespace information in a Change Stream event.
type StreamingNS struct {
	NS NS `bson:"ns"`
}

// NS represents a MongoDB namespace (database and collection).
type NS struct {
	Db   string `bson:"db"`
	Coll string `bson:"coll"`
}

// ===================================
// Mongo any collection streaming
// ===================================

// StreamType represents the operation type in a Change Stream event.
type StreamType struct {
	OperationType string `json:"operationType" bson:"operationType"`
}

// StreamInsert represents an insert operation in a Change Stream event.
// It contains the full document that was inserted.
type StreamInsert[T any] struct {
	FullDocument T `json:"fullDocument" bson:"fullDocument"`
}

// StreamUpdate represents an update operation in a Change Stream event.
// It contains the document key and the update description with changed fields.
type StreamUpdate[T any] struct {
	DocumentKey       DocumentKey          `json:"documentKey" bson:"documentKey"`
	UpdateDescription UpdateDescription[T] `json:"updateDescription" bson:"updateDescription"`
}

// StreamDelete represents a delete operation in a Change Stream event.
// It contains the document key of the deleted document.
type StreamDelete struct {
	DocumentKey DocumentKey `json:"documentKey" bson:"documentKey"`
}

// DocumentKey represents the _id field of a MongoDB document in Change Stream events.
type DocumentKey struct {
	ID primitive.ObjectID `json:"_id" bson:"_id"`
}

// UpdateDescription represents the changes in an update operation from a Change Stream event.
// It contains the fields that were updated and the fields that were removed.
type UpdateDescription[T any] struct {
	UpdatedFields T        `json:"updatedFields" bson:"updatedFields"`
	RemovedFields []string `json:"removedFields" bson:"removedFields"`
}
