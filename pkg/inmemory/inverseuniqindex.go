package inmemory

import (
	"context"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type inverseUniqueIndex[T d] struct {
	sync.RWMutex
	data  map[string]string
	cache Cache[T]
	from  []string
	to    *string
}

// NewInverseUniqIndex creates a new InverseUniqueIndex instance.
// The index maps field values (from) to a single entity ID, enforcing uniqueness.
// If 'to' is specified, it maps to a specific field value instead of the entity ID.
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
	s.Lock()
	defer s.Unlock()
	fromVal := _updateStringFieldValuesByName(it, s.from)
	if len(fromVal) == 0 {
		return
	}
	to := it.ID()
	if s.to != nil {
		toVal := updateStringFieldValueByName(it, *s.to)
		if toVal != nil {
			to = *toVal
		}
	}
	for _, fv := range fromVal {
		s.data[fv] = to
	}
}

// Update ...
func (s *inverseUniqueIndex[T]) Update(ctx context.Context, id primitive.ObjectID, updatedFields T, removedFields []string) {
	s.Lock()
	defer s.Unlock()
	updatedVal := _updateStringFieldValuesByName(updatedFields, s.from)
	if len(updatedVal) == 0 {
		return
	}
	if it, found := s.cache.Get(ctx, id.Hex()); found {
		fromVal := _updateStringFieldValuesByName(it, s.from)
		if len(fromVal) == 0 {
			return
		}
		to := it.ID()
		if s.to != nil {
			toVal := updateStringFieldValueByName(it, *s.to)
			if toVal != nil {
				to = *toVal
			}
		}
		_updVals := map[string]struct{}{}
		for _, uv := range updatedVal {
			_updVals[uv] = struct{}{}
		}
		_fromVals := map[string]struct{}{}
		for _, fv := range fromVal {
			_fromVals[fv] = struct{}{}
			if _, ok := _updVals[fv]; !ok {
				delete(s.data, fv)
			}
		}
		for _, uv := range updatedVal {
			if _, ok := _fromVals[uv]; !ok {
				s.data[uv] = to
				continue
			}
		}
	}
}

// Delete ...
func (s *inverseUniqueIndex[T]) Delete(ctx context.Context, _id primitive.ObjectID) {
	s.Lock()
	defer s.Unlock()
	if it, f := s.cache.Get(ctx, _id.Hex()); f {
		fromVal := updateStringFieldValuesByName(it, s.from)
		if fromVal != nil {
			delete(s.data, *fromVal)
		}
	}
}
