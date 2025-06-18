package inmemory

import (
	"context"
	"sync"
	"unicode"

	"github.com/google/btree"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Suffix[T d] struct {
	M
	cache Cache[T]
	from  []string
	to    *string
}

func NewSuffix[T d](index M, cache Cache[T], from []string, to *string) SuffixIndex[T] {
	return &Suffix[T]{
		M:     index,
		cache: cache,
		from:  from,
		to:    to,
	}
}

// Rebuild ...
func (s *Suffix[T]) Rebuild(ctx context.Context) {
	s.M.Start()
	for _, id := range s.cache.All(ctx) {
		if it, found := s.cache.Get(ctx, id); found {
			from := updateStringFieldValuesByName(it, s.from)
			if from == nil {
				continue
			}
			to := it.ID()
			if s.to != nil {
				_to := updateStringFieldValueByName(it, *s.to)
				if _to != nil {
					to = *_to
				}
			}
			s.M.Rebuild(to, *from)
		}
	}
	s.M.Commit()
}

// Search ...
func (s *Suffix[T]) Search(ctx context.Context, text string) (items []string) {
	return s.M.S(ctx, text)
}

// Add ...
func (s *Suffix[T]) Add(ctx context.Context, it T) {
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
	s.M.Add(to, *from)
}

// Update ...
func (s *Suffix[T]) Update(ctx context.Context, id primitive.ObjectID, updatedFields T, removedFields []string) {
	it, ok := s.cache.Get(context.Background(), id.Hex())
	if !ok {
		return
	}
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
	s.M.Update(to, *from)
}

// Delete ...
func (s *Suffix[T]) Delete(ctx context.Context, _id primitive.ObjectID) {
}

type M interface {
	Start()
	Rebuild(id string, title string)
	Commit()
	Add(id string, title string)
	Update(id string, title string)
	S(ctx context.Context, text string) (items []string)
}

type suffixTree interface {
	Reset()
	Put(in string, idx int)
	Search(in string) (out []int)
}

type m[T d] struct {
	sync.RWMutex
	cache Cache[T]
	tree  suffixTree
	old   suffixTree
}

func (s *m[T]) Start() {
	s.Lock()
	s.old.Reset()
	s.Unlock()
}

func (s *m[T]) Rebuild(id string, text string) {
	idx, found := s.cache.GetIndexByID(id)
	if !found {
		return
	}
	s.old.Put(text, idx)
}

func (s *m[T]) Commit() {
	s.Lock()
	s.tree, s.old = s.old, s.tree
	s.Unlock()
}

func (s *m[T]) Add(id string, text string) {
	s.Lock()
	defer s.Unlock()
	idx, found := s.cache.GetIndexByID(id)
	if !found {
		return
	}
	s.tree.Put(text, idx)
}

func (s *m[T]) Update(id string, text string) {
	s.Lock()
	defer s.Unlock()
	idx, found := s.cache.GetIndexByID(id)
	if !found {
		return
	}
	s.tree.Put(text, idx)
}

func (s *m[T]) S(ctx context.Context, text string) (items []string) {
	s.RLock()
	defer s.RUnlock()
	var (
		id    string
		found bool
	)
	idxs := s.tree.Search(text)
	items = make([]string, len(idxs))
	for k, idx := range idxs {
		if id, found = s.cache.GetIDByIndex(idx); found {
			items[k] = id
		}
	}
	return
}

func NewM[T d](
	cache Cache[T],
	tree suffixTree,
	old suffixTree,
) M {
	return &m[T]{
		cache: cache,
		tree:  tree,
		old:   old,
	}
}

// S ...
type S struct {
	sync.RWMutex
	intersect
	data *btree.BTree
	pool *Pool
}

// Reset ...
func (a *S) Reset() {
	a.Lock()
	defer a.Unlock()
	items := make([]*F, 0, a.data.Len())
	a.data.Ascend(func(i btree.Item) bool {
		switch b := i.(type) {
		case *F:
			items = append(items, b)
		}
		return false
	})
	for k := 0; k < len(items); k++ {
		a.data.Delete(items[k])
		t := items[k]
		items[k] = nil
		a.pool.Release(t)
	}
}

// Put ...
func (a *S) Put(in string, idx int) {
	a.Lock()
	defer a.Unlock()
	i := []rune(in)
	if len(i) < 3 {
		return
	}
	i = a.toLowerRuneSlice(i)
	if len(i) == 3 {
		a.set(i, idx)
		return
	}
	for k := 0; k < len(i)-3; k++ {
		a.set(i[k:k+3], idx)
	}
}

func (a *S) toLowerRuneSlice(input []rune) []rune {
	for i, r := range input {
		input[i] = unicode.ToLower(r)
	}
	return input
}

// Search ...
func (a *S) Search(in string) (out []int) {
	a.RLock()
	defer a.RUnlock()
	i := []rune(in)
	if len(i) < 3 {
		return
	}
	i = a.toLowerRuneSlice(i)
	if len(i) == 3 {
		p := a.pool.Acquire()
		p.key[0] = i[0]
		p.key[1] = i[1]
		p.key[2] = i[2]
		out = append(out, a.get(p)...)
		a.pool.Release(p)
		return
	}
	p := a.pool.Acquire()
	p.key[0] = i[0]
	p.key[1] = i[1]
	p.key[2] = i[2]
	out = append(out, a.get(p)...)
	if len(i) == 4 {
		p.key[0] = i[1]
		p.key[1] = i[2]
		p.key[2] = i[3]
		out = a.IntersectInt(out, a.get(p))
		return
	}
	for k := 1; k < len(i)-3; k++ {
		p.key[0] = i[k]
		p.key[1] = i[k+1]
		p.key[2] = i[k+2]
		out = a.IntersectInt(out, a.get(p))
	}
	a.pool.Release(p)
	return
}

func (a *S) set(i []rune, idx int) {
	var g btree.Item
	p := a.pool.Acquire()
	p.key[0] = i[0]
	p.key[1] = i[1]
	p.key[2] = i[2]
	g = a.data.Get(p)
	if g != nil {
		a.pool.Release(p)
		switch b := g.(type) {
		case *F:
			for k := 0; k < len(b.data); k++ {
				if b.data[k] == idx {
					return
				}
			}
			b.data = append(b.data, idx)
			return
		}
		return
	}
	p.data = append(p.data, idx)
	a.data.ReplaceOrInsert(p)
}

func (a *S) get(p *F) []int {
	g := a.data.Get(p)
	if g != nil {
		switch b := g.(type) {
		case *F:
			return b.data
		}
	}
	return nil
}

// NewS ...
func NewS(
	intersect intersect,
	data *btree.BTree,
	pool *Pool,
) *S {
	return &S{
		intersect: intersect,
		data:      data,
		pool:      pool,
	}
}

type s [3]rune

// F ...
type F struct {
	key  s
	data []int
}

// Less ...
func (a *F) Less(than btree.Item) bool {
	switch b := than.(type) {
	case *F:
		if a.key[0] < b.key[0] {
			return true
		}
		if a.key[0] > b.key[0] {
			return false
		}
		if a.key[1] < b.key[1] {
			return true
		}
		if a.key[1] > b.key[1] {
			return false
		}
		return a.key[2] < b.key[2]
	}
	return false
}

// Reset ...
func (a *F) Reset() {
	a.key[0] = 0
	a.key[1] = 0
	a.key[2] = 0
	a.data = nil
}

// Pool ...
type Pool struct {
	sync.Pool
}

// Acquire ...
func (p *Pool) Acquire() *F {
	v := p.Get()
	if v == nil {
		return &F{}
	}
	return v.(*F)
}

// Release ...
func (p *Pool) Release(req *F) {
	req.Reset()
	p.Put(req)
}

// NewPool ...
func NewPool() *Pool {
	return &Pool{
		sync.Pool{
			New: func() interface{} { return new(F) },
		},
	}
}

func NewSuffixIndex[T d](cache Cache[T], btreeDegree int, from []string, to *string) SuffixIndex[T] {
	sorterIntersector := NewIntersect()
	suffixPool := NewPool()
	return NewSuffix(
		NewM(
			cache,
			NewS(sorterIntersector, btree.New(btreeDegree), suffixPool),
			NewS(sorterIntersector, btree.New(btreeDegree), suffixPool),
		),
		cache,
		from,
		to,
	)
}

func BuildM[T d](cache Cache[T]) M {
	sorterIntersector := NewIntersect()
	suffixPool := NewPool()
	return NewM(
		cache,
		NewS(sorterIntersector, btree.New(1000), suffixPool),
		NewS(sorterIntersector, btree.New(1000), suffixPool),
	)
}
