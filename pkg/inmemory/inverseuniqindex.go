package inmemory

import (
	"context"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type inverseUniqueIndex[T d] struct {
	sync.RWMutex
	data  map[string]string
	cache Cache[T]
	from  []string
	to    *string
}

func NewInverseUniqIndex[T d](
	data map[string]string,
	cache Cache[T],
	field []string,
	to *string,
) InverseUniqueIndex[T] {
	return &inverseUniqueIndex[T]{
		data:  data,
		cache: cache,
		from:  field,
		to:    to,
	}
}

func (s *inverseUniqueIndex[T]) Get(ctx context.Context, val ...string) (id string, found bool) {
	s.RLock()
	defer s.RUnlock()
	id, found = s.data[strings.Join(val, "")]
	return
}

// Add ...
func (s *inverseUniqueIndex[T]) Add(ctx context.Context, it T) {
	logger := zerolog.Ctx(ctx)
	s.Lock()
	defer s.Unlock()
	to := it.ID()
	if s.to != nil {
		toVal := getStringFieldValueByName(it, *s.to)
		if toVal != "" {
			to = toVal
		}
	}
	fromVal := getStringFieldValuesByName(it, s.from)
	if fromVal != "" {
		s.data[fromVal] = to
	}
	logger.Debug().
		Any("from", s.from).
		Any("fromVal", fromVal).
		Any("to", to).
		Msg("InverseUniqIndex:Add")
}

// Update ...
func (s *inverseUniqueIndex[T]) Update(ctx context.Context, id primitive.ObjectID, updatedFields T, removedFields []string) {
	logger := zerolog.Ctx(ctx)
	s.Lock()
	defer s.Unlock()
	updatedVal := getStringFieldValuesByName(updatedFields, s.from)
	if updatedVal == "" {
		return
	}
	if it, found := s.cache.Get(ctx, id.Hex()); found {
		to := it.ID()
		if s.to != nil {
			to = getStringFieldValueByName(it, *s.to)
		}
		fromVal := getStringFieldValuesByName(it, s.from)
		if !compareSlices([]rune(fromVal), []rune(updatedVal)) {
			delete(s.data, fromVal)
			s.data[updatedVal] = to
		}
		logger.Debug().
			Any("from", s.from).
			Any("fromVal", fromVal).
			Any("to", to).
			Msg("InverseUniqIndex:Update")
	}
}

// Delete ...
func (s *inverseUniqueIndex[T]) Delete(ctx context.Context, _id primitive.ObjectID) {
	logger := zerolog.Ctx(ctx)
	s.Lock()
	defer s.Unlock()
	if it, f := s.cache.Get(ctx, _id.Hex()); f {
		fromVal := getStringFieldValuesByName(it, s.from)
		if fromVal != "" {
			delete(s.data, fromVal)
		}
		logger.Debug().
			Any("from", s.from).
			Any("fromVal", fromVal).
			Msg("InverseUniqIndex:Delete")
	}
}

func compareSlices(slice1, slice2 []rune) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}
