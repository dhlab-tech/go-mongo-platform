package inmemory

import (
	"context"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type inverseIndex[T d] struct {
	sync.RWMutex
	data  map[string][]string
	cache Cache[T]
	from  []string
	to    *string
}

func NewInverseIndex[T d](
	data map[string][]string,
	cache Cache[T],
	from []string,
	to *string,
) InverseIndex[T] {
	return &inverseIndex[T]{
		data:  data,
		cache: cache,
		from:  from,
		to:    to,
	}
}

func (s *inverseIndex[T]) Get(ctx context.Context, val string) (ids []string) {
	s.RLock()
	defer s.RUnlock()
	return s.data[val]
}

// Add ...
func (s *inverseIndex[T]) Add(ctx context.Context, it T) {
	s.Lock()
	defer s.Unlock()
	from := updateStringFieldValuesByName(it, s.from)
	if from == nil {
		return
	}
	to := it.ID()
	if s.to != nil {
		_to := updateStringFieldValueByName(it, *s.to)
		if _to != nil {
			to = *_to
		}
	}
	for _, d := range s.data[*from] {
		if d == to {
			return
		}
	}
	s.data[*from] = append(s.data[*from], to)
}

// Update ...
func (s *inverseIndex[T]) Update(ctx context.Context, id primitive.ObjectID, updatedFields T, removedFields []string) {
	s.Lock()
	defer s.Unlock()
	updatedVal := updateStringFieldValuesByName(updatedFields, s.from)
	if updatedVal == nil {
		return
	}
	if it, found := s.cache.Get(ctx, id.Hex()); found {
		_from := updateStringFieldValuesByName(it, s.from)
		if _from == nil {
			return
		}
		from := *_from
		to := it.ID()
		if s.to != nil {
			_to := updateStringFieldValueByName(it, *s.to)
			if _to != nil {
				to = *_to
			}
		}
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
		if from == nil {
			return
		}
		to := it.ID()
		if s.to != nil {
			_to := updateStringFieldValueByName(it, *s.to)
			if _to != nil {
				to = *_to
			}
		}
		for k, d := range s.data[*from] {
			if d == to {
				s.data[*from] = append(s.data[*from][:k], s.data[*from][k+1:]...)
				return
			}
		}
	}
}
