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
}

type eventListener[T d] interface {
	EventListener[T]
	AddListener(listener EventListener[T], before bool) (idx int)
}

type inverseIndex[T d] interface {
	EventListener[T]
	Get(ctx context.Context, val string) (ids []string)
}

type inverseUniqueIndex[T d] interface {
	EventListener[T]
	Get(ctx context.Context, val ...string) (id string, found bool)
}

type sortedIndex[T d] interface {
	EventListener[T]
	Intersect(in []string) (res []string)
}

type suffixIndex[T d] interface {
	EventListener[T]
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
	BeforeListeners []EventListener[T]
	AfterListeners  []EventListener[T]
	Notify          Notify[T]
}

// CacheWithEventListener ...
type CacheWithEventListener[T d] struct {
	Cache                cache[T]
	EventListener        eventListener[T]
	Notify               Notify[T]
	InverseIndexes       map[string]inverseIndex[T]
	InverseUniqueIndexes map[string]inverseUniqueIndex[T]
	SortedIndexes        map[string]sortedIndex[T]
	SuffixIndexes        map[string]suffixIndex[T]
	AwaitNotify          Notify[T]
}

// NewCacheWithEventListener ...
func NewCacheWithEventListener[T d](
	beforeListeners []EventListener[T],
	afterListeners []EventListener[T],
	notify Notify[T],
) *CacheWithEventListener[T] {
	c := NewCache[T](0, map[int]string{}, map[string]int{}, map[string]T{})
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
	inverseIndexes := map[string]inverseIndex[T]{}
	inverseUniqueIndexes := map[string]inverseUniqueIndex[T]{}
	sortedIndexes := map[string]sortedIndex[T]{}
	suffixIndexes := map[string]suffixIndex[T]{}
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
				sortedIndexes[indexName] = NewSortedIndex(c, 1000, []string{}, _idx.from, to)
				l.AddListener(sortedIndexes[indexName], true)
			case SuffixIndexType:
				suffixIndexes[indexName] = NewSuffixIndex(c, 1000, _idx.from, to)
				l.AddListener(suffixIndexes[indexName], true)
			}
		}
	}
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
		SortedIndexes:        sortedIndexes,
		AwaitNotify:          awaitNotify,
	}
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
