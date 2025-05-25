package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	errStreamFailed = errors.New("stream failed")
)

// Stream ...
type Stream struct {
	sync.RWMutex
	change    *mongo.ChangeStream
	listeners map[string]map[string]StreamListener // listeners by db and collection
}

// AddListener ...
func (s *Stream) AddListener(ctx context.Context, db, col string, listener StreamListener) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.listeners[db]; !ok {
		s.listeners[db] = map[string]StreamListener{}
	}
	s.listeners[db][col] = listener
}

// Listen ...
func (s *Stream) Listen(ctx context.Context) (err error) {
	var (
		p  map[string]StreamListener
		k  StreamListener
		ok bool
	)
	logger := zerolog.Ctx(ctx)
	logger.Debug().Msg("start streaming")
	defer s.change.Close(ctx)
	defer logger.Debug().Msg("stop streaming")
	for {
		if s.change.TryNext(ctx) {
			// A new event variable should be declared for each event.
			var tp StreamingNS
			if e := s.change.Decode(&tp); e != nil {
				logWithError(logger, s.change.Current, e, "error while decoding ns from stream")
				continue
			}
			s.RLock()
			if p, ok = s.listeners[tp.NS.Db]; !ok {
				s.RUnlock()
				continue
			}
			if k, ok = p[tp.NS.Coll]; !ok {
				s.RUnlock()
				continue
			}
			s.RUnlock()
			var bsonDocument bson.D
			var temporaryBytes []byte
			err = s.change.Decode(&bsonDocument)
			temporaryBytes, err = bson.MarshalExtJSON(bsonDocument, false, false)
			if err != nil {
				logger.Err(err).Msg("processing stream: unmarshal from bson to json")
				continue
			}
			logger.Debug().Bytes("income", temporaryBytes).Msg("streaming")
			err = k.Listen(ctx, temporaryBytes)
			if err != nil {
				logger.Err(err).Msg("processing stream")
				continue
			}
		}
		// If TryNext returns false, the next change is not yet available, the change stream was closed by the server,
		// or an error occurred. TryNext should only be called again for the empty batch case.
		if err = s.change.Err(); err != nil {
			logger.Err(err).Msg("change error")
			return
		}
		if s.change.ID() == 0 {
			err = errStreamFailed
			logger.Err(err).Msg("streaming failed")
			return
		}
	}
}

// NewStream ...
func NewStream(
	change *mongo.ChangeStream,
	listeners map[string]map[string]StreamListener,
) *Stream {
	return &Stream{
		change:    change,
		listeners: listeners,
	}
}

func logf(logger *zerolog.Logger, data []byte, format string, args ...interface{}) {
	fields, e := unmarshal(data)
	if e != nil {
		logger.Err(e).Fields(string(data)).Msg(fmt.Sprintf(format, args...))
	} else {
		logger.Debug().Fields(fields).Msg(fmt.Sprintf(format, args...))
	}
}

func logWithError(logger *zerolog.Logger, data []byte, err error, msg string) {
	fields, e := unmarshal(data)
	if e != nil {
		logger.Err(e).Fields(string(data)).Msg(msg)
	} else {
		logger.Err(err).Fields(fields).Msg(msg)
	}
}

func logfWithError(logger *zerolog.Logger, data []byte, err error, format string, args ...interface{}) {
	fields, e := unmarshal(data)
	if e != nil {
		logger.Err(e).Fields(string(data)).Msg(fmt.Sprintf(format, args...))
	} else {
		logger.Err(err).Fields(fields).Msg(fmt.Sprintf(format, args...))
	}
}

func unmarshal(data []byte) (fields map[string]interface{}, err error) {
	err = json.Unmarshal(data, &fields)
	return
}
