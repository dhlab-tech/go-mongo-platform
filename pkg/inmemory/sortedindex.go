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

type sortedIndex[T d] struct {
	sync.RWMutex
	sorted Sorted
	cache  Cache[T]
	from   []string
	to     *string
}

// Intersect ...
func (s *sortedIndex[T]) Intersect(in []string) (res []string) {
	return s.sorted.Intersect(in)
}

// Add ...
func (s *sortedIndex[T]) Add(ctx context.Context, it T) {
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
	s.sorted.Add(ctx, to, *from)
}

// Update ...
func (s *sortedIndex[T]) Update(ctx context.Context, id primitive.ObjectID, updatedFields T, removedFields []string) {
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
		to := it.ID()
		if s.to != nil {
			_to := updateStringFieldValueByName(it, *s.to)
			if _to != nil {
				to = *_to
			}
		}
		s.sorted.Update(ctx, to, *_from, *updatedVal)
	}
}

// Delete ...
func (s *sortedIndex[T]) Delete(ctx context.Context, _id primitive.ObjectID) {
	s.Lock()
	defer s.Unlock()
	if it, f := s.cache.Get(ctx, _id.Hex()); f {
		_from := updateStringFieldValuesByName(it, s.from)
		if _from == nil {
			return
		}
		to := it.ID()
		if s.to != nil {
			_to := updateStringFieldValueByName(it, *s.to)
			if _to != nil {
				to = *_to
			}
		}
		s.sorted.Delete(ctx, to, *_from)
	}
}

// NewSortedIndex ...
func NewSortedIndex[T d](
	sorted Sorted,
	cache Cache[T],
	from []string,
	to *string,
) SortedIndex[T] {
	return &sortedIndex[T]{
		sorted: sorted,
		cache:  cache,
		from:   from,
		to:     to,
	}
}

type Sorted interface {
	Intersect(in []string) (res []string)
	Add(ctx context.Context, id string, title string)
	Update(ctx context.Context, id string, old string, title string)
	Delete(ctx context.Context, id string, title string)
}

type sorted struct {
	sync.RWMutex
	idx *btree.BTree
	ids []string
}

func NewSorted(degree int, ids []string) Sorted {
	return &sorted{
		idx: btree.New(degree),
		ids: ids,
	}
}

func BuildSorted() Sorted {
	return NewSorted(1000, []string{})
}

func (s *sorted) Intersect(in []string) (res []string) {
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

func (s *sorted) Add(ctx context.Context, id string, title string) {
	s.Lock()
	f := item{
		id:   id,
		text: title,
	}
	if s.idx.Get(f) != nil {
		return
	}
	s.idx.ReplaceOrInsert(f)
	s.Unlock()
	s.fill()
}

// Update ...
func (s *sorted) Update(ctx context.Context, id string, old string, title string) {
	s.Lock()
	f := item{}
	f.id = id
	f.text = old
	s.idx.Delete(f)
	f.text = title
	s.idx.ReplaceOrInsert(f)
	s.Unlock()
	s.fill()
}

// Delete ...
func (s *sorted) Delete(ctx context.Context, id string, title string) {
	s.Lock()
	s.idx.Delete(item{
		id:   id,
		text: title,
	})
	s.Unlock()
	s.fill()
}

func (s *sorted) fill() {
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
