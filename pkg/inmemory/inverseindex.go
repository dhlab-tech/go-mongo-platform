package inmemory

import (
	"context"
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

func (s *inverseIndex[T]) Get(ctx context.Context, val *string) (ids []string) {
	s.RLock()
	defer s.RUnlock()
	if val != nil {
		return s.data[*val]
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

// Update ...
func (s *inverseIndex[T]) Update(ctx context.Context, id primitive.ObjectID, updatedFields T, removedFields []string) {
	s.Lock()
	defer s.Unlock()
	updatedVal := updateStringFieldValuesByName(updatedFields, s.from)
	if it, found := s.cache.Get(ctx, id.Hex()); found {
		_from := updateStringFieldValuesByName(it, s.from)
		to := it.ID()
		if s.to != nil {
			_to := updateStringFieldValueByName(it, *s.to)
			if _to != nil {
				to = *_to
			}
		}
		if updatedVal == nil && _from == nil {
			return
		}
		if _from == nil {
			for k, v := range s.nilData {
				if v == to {
					s.nilData = append(s.nilData[:k], s.nilData[k+1:]...)
					break
				}
			}
			s.data[*updatedVal] = append(s.data[*updatedVal], to)
			return
		}
		if updatedVal == nil {
			from := *_from
			for k, d := range s.data[from] {
				if d == to {
					s.data[from] = append(s.data[from][:k], s.data[from][k+1:]...)
					break
				}
			}
			s.nilData = append(s.nilData, to)
			return
		}
		from := *_from
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
