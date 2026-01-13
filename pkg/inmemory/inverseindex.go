package inmemory

import (
	"context"
	"strings"
	"sync"

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

// NewInverseIndex creates a new InverseIndex instance.
// The index maps field values (from) to lists of entity IDs, allowing multiple entities per value.
// If 'to' is specified, it maps to a specific field value instead of the entity ID.
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
	if len(val) >= 1 && val[0] != nil {
		_val := make([]string, len(val))
		for i := 0; i < len(val); i++ {
			if val[i] != nil {
				_val[i] = *val[i]
			}
		}
		return s.data[strings.Join(_val, "")]
	}
	return s.nilData
}

// Add ...
func (s *inverseIndex[T]) Add(ctx context.Context, it T) {
	s.Lock()
	defer s.Unlock()
	fromVal := updateStringFieldValuesByName(it, s.from)
	to := it.ID()
	if s.to != nil {
		_to := updateStringFieldValueByName(it, *s.to)
		if _to != nil {
			to = *_to
		}
	}
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
}

// Update updates the index when entity fields change.
// Note: If a field from the index key (s.from) is removed (becomes nil),
// the index entry is not automatically removed. This is a known limitation.
// The index will be updated correctly on the next full resync or when the entity
// is recreated with the field populated.
func (s *inverseIndex[T]) Update(ctx context.Context, id primitive.ObjectID, updatedFields T, removedFields []string) {
	s.Lock()
	defer s.Unlock()
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
