package inmemory

import (
	"context"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InverseIndex[T d] struct {
	sync.RWMutex
	data  map[string][]string
	cache cache[T]
	from  []string
	to    *string
}

func NewInverseIndex[T d](
	data map[string][]string,
	cache cache[T],
	from []string,
	to *string,
) *InverseIndex[T] {
	return &InverseIndex[T]{
		data:  data,
		cache: cache,
		from:  from,
		to:    to,
	}
}

func (s *InverseIndex[T]) Get(ctx context.Context, val string) (ids []string) {
	s.RLock()
	defer s.RUnlock()
	return s.data[val]
}

// Add ...
func (s *InverseIndex[T]) Add(ctx context.Context, it T) {
	s.Lock()
	defer s.Unlock()
	from := getStringFieldValuesByName(it, s.from)
	to := it.ID()
	if s.to != nil {
		to = getStringFieldValueByName(it, *s.to)
	}
	for _, d := range s.data[from] {
		if d == to {
			return
		}
	}
	s.data[from] = append(s.data[from], to)
}

// Update ...
func (s *InverseIndex[T]) Update(ctx context.Context, id primitive.ObjectID, updatedFields T, removedFields []string) {
	s.Lock()
	defer s.Unlock()
	updatedVal := getStringFieldValuesByName(updatedFields, s.from)
	if it, found := s.cache.Get(ctx, id.Hex()); found {
		from := getStringFieldValuesByName(it, s.from)
		to := it.ID()
		if s.to != nil {
			to = getStringFieldValueByName(it, *s.to)
		}
		for k, d := range s.data[from] {
			if d == to {
				s.data[from] = append(s.data[from][:k], s.data[from][k+1:]...)
				break
			}
		}
		s.data[updatedVal] = append(s.data[updatedVal], to)
	}
}

// Delete ...
func (s *InverseIndex[T]) Delete(ctx context.Context, _id primitive.ObjectID) {
	s.Lock()
	defer s.Unlock()
	if it, f := s.cache.Get(ctx, _id.Hex()); f {
		from := getStringFieldValuesByName(it, s.from)
		to := it.ID()
		if s.to != nil {
			to = getStringFieldValueByName(it, *s.to)
		}
		for k, d := range s.data[from] {
			if d == to {
				s.data[from] = append(s.data[from][:k], s.data[from][k+1:]...)
				return
			}
		}
	}
}
