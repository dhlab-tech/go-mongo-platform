package inmemory

import (
	"context"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type inverseIndex[T d] struct {
	sync.RWMutex
	data    map[string][]string
	nilData []string
	cache   Cache[T]
	from    []string
	to      *string
}

func NewInverseIndex[T d](
	data map[string][]string,
	nilData []string,
	cache Cache[T],
	from []string,
	to *string,
) InverseIndex[T] {
	return &inverseIndex[T]{
		data:    data,
		nilData: nilData,
		cache:   cache,
		from:    from,
		to:      to,
	}
}

func (s *inverseIndex[T]) Get(ctx context.Context, val ...*string) (ids []string) {
	s.RLock()
	defer s.RUnlock()
	zerolog.Ctx(ctx).Debug().
		Any("from", s.from).
		Any("val", val).
		Any("data", s.data).
		Msg("InverseIndex:Get")
	if len(val) >= 1 && val[0] != nil {
		zerolog.Ctx(ctx).Debug().
			Any("from", s.from).
			Any("val", val).
			Any("data", s.data).
			Msg("InverseIndex:Get:val not nil")
		_val := make([]string, len(val))
		for i := 0; i < len(val); i++ {
			if val[i] != nil {
				_val[i] = *val[i]
			}
		}
		return s.data[strings.Join(_val, "")]
	}
	zerolog.Ctx(ctx).Debug().
		Any("from", s.from).
		Any("val", val).
		Any("data", s.data).
		Msg("InverseIndex:Get:nil val")
	return s.nilData
}

// Add ...
func (s *inverseIndex[T]) Add(ctx context.Context, it T) {
	logger := zerolog.Ctx(ctx)
	s.Lock()
	defer s.Unlock()
	logger.Debug().
		Any("id", it.ID()).
		Any("it", it).
		Any("from", s.from).
		Msg("InverseIndex:Add:start")
	fromVal := updateStringFieldValuesByName(it, s.from)
	to := it.ID()
	if s.to != nil {
		_to := updateStringFieldValueByName(it, *s.to)
		if _to != nil {
			to = *_to
		}
	}
	logger.Debug().
		Any("id", it.ID()).
		Any("from", s.from).
		Any("fromVal", fromVal).
		Msg("InverseIndex:Add:after parse from val")
	if fromVal == nil {
		s.nilData = append(s.nilData, to)
		return
	}
	for _, d := range s.data[*fromVal] {
		if d == to {
			return
		}
	}
	s.data[*fromVal] = append(s.data[*fromVal], to)
	logger.Debug().
		Any("id", it.ID()).
		Any("from", s.from).
		Any("fromVal", fromVal).
		Any("to", to).
		Msg("InverseIndex:Add")
}

// Update ...
func (s *inverseIndex[T]) Update(ctx context.Context, id primitive.ObjectID, updatedFields T, removedFields []string) {
	logger := zerolog.Ctx(ctx)
	s.Lock()
	defer s.Unlock()
	logger.Debug().
		Any("id", id.Hex()).
		Any("updatedFields", updatedFields).
		Any("from", s.from).
		Msg("InverseIndex:Update:start")
	// todo Поле может стать nil
	// for _, _ = range removedFields {
	// }
	updatedVal := updateStringFieldValuesByName(updatedFields, s.from)
	if updatedVal == nil {
		return
	}
	if it, found := s.cache.Get(ctx, id.Hex()); found {
		fromVal := updateStringFieldValuesByName(it, s.from)
		to := it.ID()
		if s.to != nil {
			_to := updateStringFieldValueByName(it, *s.to)
			if _to != nil {
				to = *_to
			}
		}
		logger.Debug().
			Any("id", id.Hex()).
			Any("fromVal", fromVal).
			Any("updatedVal", updatedVal).
			Any("updatedFields", updatedFields).
			Any("from", s.from).
			Msg("InverseIndex:Update:after parse from val")
		if fromVal == nil {
			for k, v := range s.nilData {
				if v == to {
					s.nilData = append(s.nilData[:k], s.nilData[k+1:]...)
					break
				}
			}
			s.data[*updatedVal] = append(s.data[*updatedVal], to)
			return
		}
		from := *fromVal
		for k, d := range s.data[from] {
			if d == to {
				s.data[from] = append(s.data[from][:k], s.data[from][k+1:]...)
				break
			}
		}
		s.data[*updatedVal] = append(s.data[*updatedVal], to)
		logger.Debug().
			Any("from", s.from).
			Any("fromVal", fromVal).
			Any("updatedVal", updatedVal).
			Any("data", s.data).
			Any("to", to).
			Msg("InverseIndex:Update")
	}
}

// Delete ...
func (s *inverseIndex[T]) Delete(ctx context.Context, _id primitive.ObjectID) {
	s.Lock()
	defer s.Unlock()
	if it, f := s.cache.Get(ctx, _id.Hex()); f {
		from := updateStringFieldValuesByName(it, s.from)
		to := it.ID()
		if s.to != nil {
			_to := updateStringFieldValueByName(it, *s.to)
			if _to != nil {
				to = *_to
			}
		}
		if from == nil {
			s.nilData = append(s.nilData, to)
			return
		}
		for k, d := range s.data[*from] {
			if d == to {
				s.data[*from] = append(s.data[*from][:k], s.data[*from][k+1:]...)
				return
			}
		}
	}
}
