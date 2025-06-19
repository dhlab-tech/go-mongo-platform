package inmemory

import (
	"context"
	"reflect"
	"strings"
	"time"

	mng "go.mongodb.org/mongo-driver/mongo"
)

const (
	InverseIndexType       = "inverse"
	InverseUniqueIndexType = "inverse_unique"
	SortdIndexType         = "sorted"
	SuffixIndexType        = "suffix"
)

type d interface {
	any
	ID() string
	Version() *int64
	SetDeleted(bool)
}

type EventListener[T d] interface {
	StreamEventListener[T]
	AddListener(listener StreamEventListener[T], before bool) (idx int)
}

type InverseIndex[T d] interface {
	StreamEventListener[T]
	Get(ctx context.Context, val string) (ids []string)
}

type InverseUniqueIndex[T d] interface {
	StreamEventListener[T]
	Get(ctx context.Context, val ...string) (id string, found bool)
}

type SortedIndex[T d] interface {
	StreamEventListener[T]
	Intersect(in []string) (res []string)
}

type SuffixIndex[T d] interface {
	StreamEventListener[T]
	Search(ctx context.Context, text string) (items []string)
	Rebuild(ctx context.Context)
}

type MongoDeps struct {
	Client            *mng.Client
	Db                string
	ConnectionTimeout time.Duration
}

// Entity ...
type Entity[T d] struct {
	Collection      string
	BeforeListeners []StreamEventListener[T]
	AfterListeners  []StreamEventListener[T]
	Notify          Notify[T]
	Option          func(InMemory[T])
}

// CacheWithEventListener ...
type CacheWithEventListener[T d] struct {
	Cache                Cache[T]
	EventListener        EventListener[T]
	Notify               Notify[T]
	InverseIndexes       map[string]InverseIndex[T]
	InverseUniqueIndexes map[string]InverseUniqueIndex[T]
	SortedIndexes        map[string]SortedIndex[T]
	SuffixIndexes        map[string]SuffixIndex[T]
	AwaitNotify          Notify[T]
}

// NewCacheWithEventListener ...
func NewCacheWithEventListener[T d](
	beforeListeners []StreamEventListener[T],
	afterListeners []StreamEventListener[T],
	notify Notify[T],
) *CacheWithEventListener[T] {
	c := NewCache[T](map[string]T{})
	l := NewListener[T](c)
	for _, s := range beforeListeners {
		l.AddListener(s, true)
	}
	for _, s := range afterListeners {
		l.AddListener(s, false)
	}
	if notify != nil && !reflect.ValueOf(notify).IsNil() {
		l.AddListener(notify, false)
	}
	// init indexes
	inverseIndexes, inverseUniqueIndexes, sortedIndexes, suffixIndexes := buildIndexes(l, c)
	awaitNotify := NewNotifier[T](
		map[string]map[string]func(){},
		map[string]map[string]func(){},
		map[string]map[string]func(){},
	)
	l.AddListener(awaitNotify, false)
	return &CacheWithEventListener[T]{
		Cache:                c,
		EventListener:        l,
		Notify:               notify,
		InverseIndexes:       inverseIndexes,
		InverseUniqueIndexes: inverseUniqueIndexes,
		SuffixIndexes:        suffixIndexes,
		SortedIndexes:        sortedIndexes,
		AwaitNotify:          awaitNotify,
	}
}

func buildIndexes[T d](l *Listener[T], c Cache[T]) (
	inverseIndexes map[string]InverseIndex[T],
	inverseUniqueIndexes map[string]InverseUniqueIndex[T],
	sortedIndexes map[string]SortedIndex[T],
	suffixIndexes map[string]SuffixIndex[T],
) {
	inverseIndexes = map[string]InverseIndex[T]{}
	inverseUniqueIndexes = map[string]InverseUniqueIndex[T]{}
	sortedIndexes = map[string]SortedIndex[T]{}
	suffixIndexes = map[string]SuffixIndex[T]{}
	var instance T
	t := reflect.TypeOf(instance)
	var v reflect.Value
	if t.Kind() == reflect.Ptr {
		v = reflect.New(t.Elem()).Elem()
	} else {
		v = reflect.New(t).Elem()
	}
	for indexType, idx := range prepareIdxs(v) {
		for indexName, _idx := range idx {
			var to *string
			if _idx.to != "" {
				_to := _idx.to
				to = &_to
			}
			switch indexType {
			case InverseIndexType:
				inverseIndexes[indexName] = NewInverseIndex(map[string][]string{}, c, _idx.from, to)
				l.AddListener(inverseIndexes[indexName], true)
			case InverseUniqueIndexType:
				inverseUniqueIndexes[indexName] = NewInverseUniqIndex(map[string]string{}, c, _idx.from, to)
				l.AddListener(inverseUniqueIndexes[indexName], true)
			case SortdIndexType:
				sortedIndexes[indexName] = NewSortedIndex(NewSorted(1000, []string{}), c, _idx.from, to)
				l.AddListener(sortedIndexes[indexName], true)
			case SuffixIndexType:
				suffixIndexes[indexName] = NewSuffixIndex(c, 1000, _idx.from, to)
				l.AddListener(suffixIndexes[indexName], false)
			}
		}
	}
	return
}

type idx struct {
	from []string
	to   string
}

func prepareIdxs(v reflect.Value) (idxs map[string]map[string]*idx) {
	t := reflect.TypeOf(v.Interface())
	idxs = map[string]map[string]*idx{}
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i).Name
		_v := v.Field(i)
		_t := reflect.TypeOf(_v.Interface())
		if _t.Kind() == reflect.Ptr {
			_v = reflect.New(_t.Elem()).Elem()
		} else {
			_v = reflect.New(_t).Elem()
		}
		if _v.Kind() == reflect.Struct {
			if strings.Compare(_v.Type().Name(), "ObjectID") == 0 ||
				strings.Compare(_v.Type().Name(), "RawMessage") == 0 ||
				strings.Compare(_v.Type().Name(), "Decimal") == 0 {
				continue
			}
			for _indexType, _idxt := range prepareIdxs(_v) {
				for _indexName, _idx := range _idxt {
					if _, ok := idxs[_indexType]; !ok {
						idxs[_indexType] = map[string]*idx{}
					}
					if idxs[_indexType][_indexName] == nil {
						idxs[_indexType][_indexName] = &idx{}
					}
					for _, from := range _idx.from {
						idxs[_indexType][_indexName].from = append(idxs[_indexType][_indexName].from, field+"+"+from)
					}
					if _idx.to != "" {
						idxs[_indexType][_indexName].to = field + "+" + _idx.to
					}
				}
			}
			continue
		}
		for _, index := range strings.Split(t.Field(i).Tag.Get("indexes"), ",") {
			if t.Field(i).Tag.Get("bson") == "-" {
				continue
			}
			_idx := strings.Split(index, ":")
			var indexType string
			indexName := t.Field(i).Tag.Get("bson")
			direction := "from"
			if len(_idx) == 3 {
				indexType, indexName, direction = _idx[0], _idx[1], _idx[2]
			} else if len(_idx) == 2 {
				indexType, indexName = _idx[0], _idx[1]
			} else if len(_idx) == 1 {
				indexType = _idx[0]
			}
			if indexType == "" {
				continue
			}
			_, ok := idxs[indexType]
			if !ok {
				idxs[indexType] = map[string]*idx{}
			}
			if idxs[indexType][indexName] == nil {
				idxs[indexType][indexName] = &idx{}
			}
			if direction == "from" {
				idxs[indexType][indexName].from = append(idxs[indexType][indexName].from, field)
			} else if direction == "to" {
				idxs[indexType][indexName].to = field
			}
		}
	}
	return
}
