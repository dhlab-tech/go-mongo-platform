package mongo

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog"
)

const (
	// InsertOperationType represents an insert operation in MongoDB Change Streams.
	InsertOperationType = "insert"
	// UpdateOperationType represents an update operation in MongoDB Change Streams.
	UpdateOperationType = "update"
	// DeleteOperationType represents a delete operation in MongoDB Change Streams.
	DeleteOperationType = "delete"
)

// Listener processes MongoDB Change Stream events and applies them to the in-memory projection.
// It decodes Change Stream events (insert, update, delete) and calls the appropriate
// handler methods (Add, Update, Delete) to keep the in-memory cache synchronized.
type Listener[T d] struct {
	collection string
	handler    handler[T]
}

// Listen processes a Change Stream event from MongoDB.
// It decodes the event JSON, determines the operation type (insert/update/delete),
// and calls the appropriate handler method to update the in-memory projection.
func (s *Listener[T]) Listen(ctx context.Context, change []byte) (err error) {
	logger := zerolog.Ctx(ctx)
	// A new event variable should be declared for each event.
	var tp StreamType
	if e := json.Unmarshal(change, &tp); e != nil {
		logfWithError(logger, change, e, "error while decoding type from %s collection stream", s.collection)
		return
	}
	logf(logger, change, "%s stream %s", s.collection, tp.OperationType)
	switch tp.OperationType {
	case InsertOperationType:
		var decoded StreamInsert[T]
		if e := json.Unmarshal(change, &decoded); e != nil {
			logfWithError(logger, change, e, "error while decoding insert op from %s collection", s.collection)
			return
		}
		s.handler.Add(ctx, decoded.FullDocument)
	case UpdateOperationType:
		var decoded StreamUpdate[T]
		if e := json.Unmarshal(change, &decoded); e != nil {
			logfWithError(logger, change, e, "error while decoding update op from %s collection", s.collection)
			return
		}
		s.handler.Update(
			ctx,
			decoded.DocumentKey.ID,
			decoded.UpdateDescription.UpdatedFields,
			decoded.UpdateDescription.RemovedFields,
		)
	case DeleteOperationType:
		var decoded StreamDelete
		if e := json.Unmarshal(change, &decoded); e != nil {
			logfWithError(logger, change, e, "error while decoding delete operation from %s collection", s.collection)
			return
		}
		s.handler.Delete(context.Background(), decoded.DocumentKey.ID)
	}
	return
}

// NewListener creates a new Listener for processing Change Stream events.
// The handler is called for each Change Stream event to update the in-memory projection.
func NewListener[T d](
	collection string,
	handler handler[T],
) *Listener[T] {
	return &Listener[T]{
		collection: collection,
		handler:    handler,
	}
}
