package inmemory

import (
	"context"
	"sort"
	"sync"
	"unicode"

	"github.com/google/btree"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateSuffix is a SuffixIndex wrapper that handles updates by removing old values before cache updates.
// This ensures the cache always has the correct data when the main suffix index processes the update.
type UpdateSuffix[T d] struct {
	M
	cache Cache[T]
	from  []string
	to    *string
}

// NewUpdateSuffix creates a new UpdateSuffix instance for handling suffix index updates.
func NewUpdateSuffix[T d](index M, cache Cache[T], from []string, to *string) SuffixIndex[T] {
	return &Suffix[T]{
		M:     index,
		cache: cache,
		from:  from,
		to:    to,
	}
}

func (s *UpdateSuffix[T]) Search(ctx context.Context, text string) (items []string) {
	return
}

func (s *UpdateSuffix[T]) Add(ctx context.Context, it T) {
}

// Update removes the old value from the index.
// This is done because the main suffix index is included in the chain after the cache update,
// but we can only delete data before the cache update to have the old data in the cache.
// With this approach, we avoid the need to rebuild the cache - it's always up-to-date.
func (s *UpdateSuffix[T]) Update(ctx context.Context, id primitive.ObjectID, updatedFields T, removedFields []string) {
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
	s.M.Delete(to, *from)
}

func (s *UpdateSuffix[T]) Delete(ctx context.Context, _id primitive.ObjectID) {
}

// Suffix provides full-text search capabilities using suffix matching.
// Search performs exact text matching, while Find uses trigram-based fuzzy search.
type Suffix[T d] struct {
	M
	cache Cache[T]
	from  []string
	to    *string
}

// NewSuffix creates a new Suffix instance for full-text search.
func NewSuffix[T d](index M, cache Cache[T], from []string, to *string) SuffixIndex[T] {
	return &Suffix[T]{
		M:     index,
		cache: cache,
		from:  from,
		to:    to,
	}
}

// Search uses exact text matching
func (s *Suffix[T]) Search(ctx context.Context, text string) (items []string) {
	return s.M.Search(ctx, text)
}

// Find uses trigram-based search
// and results are sorted by match frequency
func (s *Suffix[T]) Find(ctx context.Context, text string) (items []string) {
	return s.M.Find(ctx, text)
}

// Add ...
func (s *Suffix[T]) Add(ctx context.Context, it T) {
	fromVal := updateStringFieldValuesByName(it, s.from)
	if fromVal == nil {
		return
	}
	to := it.ID()
	if s.to != nil {
		_to := updateStringFieldValueByName(it, *s.to)
		if _to != nil {
			to = *_to
		}
	}
	s.M.Add(to, *fromVal)
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
	it, ok := s.cache.Get(context.Background(), _id.Hex())
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
	s.M.Delete(to, *from)
}

// M provides the core suffix index operations for full-text search.
// It maintains a suffix tree for efficient text matching.
type M interface {
	Add(id string, title string)
	Update(id string, title string)
	Delete(id string, text string)
	Search(ctx context.Context, text string) (items []string)
	Find(ctx context.Context, text string) (items []string)
}

type suffixTree interface {
	Reset()
	Put(in string, idx int)
	Delete(in string, idx int)
	Search(in string) (out []int)
	Find(in string) (out []int)
}

type m[T d] struct {
	sync.RWMutex
	cache Cache[T]
	tree  suffixTree
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

func (s *m[T]) Delete(id string, text string) {
	s.Lock()
	defer s.Unlock()
	idx, found := s.cache.GetIndexByID(id)
	if !found {
		return
	}
	s.tree.Delete(text, idx)
}

func (s *m[T]) Search(ctx context.Context, text string) (items []string) {
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

func (s *m[T]) Find(ctx context.Context, text string) (items []string) {
	s.RLock()
	defer s.RUnlock()
	var (
		id    string
		found bool
	)
	idxs := s.tree.Find(text)
	items = make([]string, len(idxs))
	for k, idx := range idxs {
		if id, found = s.cache.GetIDByIndex(idx); found {
			items[k] = id
		}
	}
	return
}

// NewM creates a new M instance (suffix index) with the specified cache and suffix tree.
func NewM[T d](
	cache Cache[T],
	tree suffixTree,
) M {
	return &m[T]{
		cache: cache,
		tree:  tree,
	}
}

// S implements a suffix tree using a B-tree for efficient suffix-based text search.
type S struct {
	sync.RWMutex
	intersect
	data *btree.BTree
	pool *Pool
	ms   map[rune]struct{}
}

// Reset clears all data from the suffix tree and returns items to the pool.
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
	for k := 0; k <= len(i)-3; k++ {
		a.set(i[k:k+3], idx)
	}
}

func (a *S) Delete(in string, idx int) {
	a.Lock()
	defer a.Unlock()
	i := []rune(in)
	if len(i) < 3 {
		return
	}
	i = a.toLowerRuneSlice(i)
	if len(i) == 3 {
		a.delete(i, idx)
		return
	}
	for k := 0; k <= len(i)-3; k++ {
		a.delete(i[k:k+3], idx)
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
	for k := 1; k <= len(i)-3; k++ {
		p.key[0] = i[k]
		p.key[1] = i[k+1]
		p.key[2] = i[k+2]
		out = a.IntersectInt(out, a.get(p))
	}
	a.pool.Release(p)
	return
}

func (a *S) Find(in string) (out []int) {
	_in := []rune(in)
	for i := len(_in) - 1; i >= 0; i-- {
		if _, ok := a.ms[_in[i]]; ok {
			_in = append(_in[:i], _in[i+1:]...)
		}
	}
	type pp struct {
		Idx   int
		Count int
	}
	res := map[int]pp{}
	for k := 0; k <= len(_in)-3; k++ {
		for _, idx := range a.Search(string(_in[k : k+3])) {
			_pp := res[idx]
			_pp.Idx = idx
			_pp.Count++
			res[idx] = _pp
		}
	}
	_res := []pp{}
	for _, _pp := range res {
		_res = append(_res, _pp)
	}
	sort.Slice(_res, func(i, j int) bool {
		return _res[i].Count > _res[j].Count
	})
	for _, v := range _res {
		out = append(out, v.Idx)
	}
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

func (a *S) delete(i []rune, idx int) {
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
					b.data = append(b.data[:k], b.data[k+1:]...)
					a.data.ReplaceOrInsert(b)
					return
				}
			}
		}
	}
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

// NewS creates a new S instance (suffix tree) with the specified components.
func NewS(
	intersect intersect,
	data *btree.BTree,
	pool *Pool,
) *S {
	a := S{
		intersect: intersect,
		data:      data,
		pool:      pool,
	}
	s := append([]rune(` _-=+()*&^%$#@!~!"№;%:?[]{}\|/,.><`), []rune("`")...)
	a.ms = map[rune]struct{}{}
	for _, v := range s {
		a.ms[v] = struct{}{}
	}
	return &a
}

type s [3]rune

// F represents a trigram (3-character sequence) in the suffix tree with associated entity indices.
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

// Pool provides object pooling for F instances to reduce allocations.
type Pool struct {
	sync.Pool
}

// Acquire gets an F instance from the pool or creates a new one if the pool is empty.
func (p *Pool) Acquire() *F {
	v := p.Get()
	if v == nil {
		return &F{}
	}
	return v.(*F)
}

// Release returns an F instance to the pool after resetting it.
func (p *Pool) Release(req *F) {
	req.Reset()
	p.Put(req)
}

// NewPool creates a new Pool for F instances.
func NewPool() *Pool {
	return &Pool{
		sync.Pool{
			New: func() interface{} { return new(F) },
		},
	}
}

// NewSuffixIndex creates a pair of SuffixIndex instances: one for regular operations and one for updates.
// The update index handles removing old values before cache updates to maintain consistency.
func NewSuffixIndex[T d](cache Cache[T], btreeDegree int, from []string, to *string) (SuffixIndex[T], SuffixIndex[T]) {
	sorterIntersector := NewIntersect()
	suffixPool := NewPool()
	m := NewM(
		cache,
		NewS(sorterIntersector, btree.New(btreeDegree), suffixPool),
	)
	return NewSuffix(m, cache, from, to), NewUpdateSuffix(m, cache, from, to)
}

// BuildM creates a new M instance (suffix index) with default settings.
func BuildM[T d](cache Cache[T]) M {
	sorterIntersector := NewIntersect()
	suffixPool := NewPool()
	return NewM(
		cache,
		NewS(sorterIntersector, btree.New(1000), suffixPool),
	)
}
