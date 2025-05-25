package mongo

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog"
)

const (
	// InsertOperationType ...
	InsertOperationType = "insert"
	// UpdateOperationType ...
	UpdateOperationType = "update"
	// DeleteOperationType ...
	DeleteOperationType = "delete"
)

// Listener ...
type Listener[T d] struct {
	collection string
	handler    handler[T]
}

// Listen ...
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
		logger.Debug().Bytes("data", change).Msg("insert income")
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

// NewListener ...
func NewListener[T d](
	collection string,
	handler handler[T],
) *Listener[T] {
	return &Listener[T]{
		collection: collection,
		handler:    handler,
	}
}
