package inmemory

import (
	"context"
	"sync"

	"github.com/google/btree"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type item struct {
	id   string
	text string
}

func (s item) Less(than btree.Item) bool {
	switch a := than.(type) {
	case item:
		return s.text+s.id < a.text+a.id
	}
	return false
}

type SortedIndex[T d] struct {
	sync.RWMutex
	idx   *btree.BTree
	cache cache[T]
	ids   []string
	from  []string
	to    *string
}

// Intersect ...
func (s *SortedIndex[T]) Intersect(in []string) (res []string) {
	t := make(map[string]struct{}, len(in))
	for _, v := range in {
		t[v] = struct{}{}
	}
	res = make([]string, 0, len(in))
	for _, d := range s.ids {
		if _, ok := t[d]; ok {
			res = append(res, d)
		}
	}
	return
}

// Add ...
func (s *SortedIndex[T]) Add(ctx context.Context, it T) {
	s.Lock()
	defer s.Unlock()
	to := it.ID()
	if s.to != nil {
		to = getStringFieldValueByName(it, *s.to)
	}
	f := item{
		id:   to,
		text: getStringFieldValuesByName(it, s.from),
	}
	if s.idx.Get(f) != nil {
		return
	}
	s.idx.ReplaceOrInsert(f)
	s.fill()
}

// Update ...
func (s *SortedIndex[T]) Update(ctx context.Context, id primitive.ObjectID, updatedFields T, removedFields []string) {
	s.Lock()
	defer s.Unlock()
	if it, found := s.cache.Get(ctx, id.Hex()); found {
		to := it.ID()
		if s.to != nil {
			to = getStringFieldValueByName(it, *s.to)
		}
		from := getStringFieldValuesByName(it, s.from)
		f := item{}
		f.id = to
		if from != "" {
			f.text = from
			s.idx.Delete(f)
		}
		f.text = getStringFieldValuesByName(updatedFields, s.from)
		s.idx.ReplaceOrInsert(f)
		s.fill()
	}
}

// Delete ...
func (s *SortedIndex[T]) Delete(ctx context.Context, _id primitive.ObjectID) {
	s.Lock()
	defer s.Unlock()
	if it, f := s.cache.Get(ctx, _id.Hex()); f {
		to := it.ID()
		if s.to != nil {
			to = getStringFieldValueByName(it, *s.to)
		}
		s.idx.Delete(item{
			id:   to,
			text: getStringFieldValuesByName(it, s.from),
		})
		s.fill()
	}
}

func (s *SortedIndex[T]) fill() {
	ids := make([]string, s.idx.Len())
	s.idx.Ascend(func(i btree.Item) bool {
		switch a := i.(type) {
		case item:
			ids = append(ids, a.id)
			return true
		}
		return false
	})
	s.ids = ids
}

// NewSortedIndex ...
func NewSortedIndex[T d](
	cache cache[T],
	btreeDegree int,
	ids []string,
	from []string,
	to *string,
) *SortedIndex[T] {
	return &SortedIndex[T]{
		ids:   ids,
		idx:   btree.New(btreeDegree),
		cache: cache,
		from:  from,
		to:    to,
	}
}
