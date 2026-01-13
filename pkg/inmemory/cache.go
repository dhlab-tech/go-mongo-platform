package inmemory

import (
	"context"
	"reflect"
	"strings"
	"sync/atomic"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Cache provides an in-memory storage for typed entities with indexing support.
// It maintains entities in memory and provides fast lookup by ID or index.
// The cache is synchronized with MongoDB through Change Streams.
type Cache[T d] interface {
	All(ctx context.Context) (ids []string)
	Get(ctx context.Context, id string) (b T, found bool)
	GetByIndex(ctx context.Context, idx int) (r T, f bool)
	GetIndexByID(id string) (idx int, found bool)
	GetIDByIndex(idx int) (id string, found bool)
	Add(ctx context.Context, v T)
	Update(ctx context.Context, _id primitive.ObjectID, updatedFields T, removedFields []string)
	Delete(ctx context.Context, _id primitive.ObjectID)
}

// cache ...
type cache[T d] struct {
	Idx
	data map[string]T
}

// All returns all entity IDs currently stored in the cache.
func (c *cache[T]) All(ctx context.Context) (ids []string) {
	c.RLock()
	defer c.RUnlock()
	ids = make([]string, len(c.data))
	i := 0
	for _, d := range c.data {
		ids[i] = d.ID()
		i++
	}
	return
}

// Get retrieves an entity by its ID from the cache.
// Returns the entity and a boolean indicating if it was found.
func (c *cache[T]) Get(ctx context.Context, id string) (r T, f bool) {
	c.RLock()
	defer c.RUnlock()
	if d, found := c.data[id]; found {
		ps, err := c.prepareCreate(ctx, reflect.ValueOf(d))
		if err != nil {
			return
		}
		r = ps.Interface().(T)
		f = found
		return
	}
	return
}

// GetByIndex retrieves an entity by its index position and removes it from the cache.
// This is useful for processing entities in order and removing them as they are processed.
func (c *cache[T]) GetByIndex(ctx context.Context, idx int) (r T, f bool) {
	var (
		id string
	)
	c.RLock()
	defer c.RUnlock()
	if id, f = c.GetIDByIndex(idx); f {
		if r, f = c.Get(ctx, id); !f {
			return
		}
		c.deleteByID(id)
	}
	c.deleteByIdx(idx)
	return
}

// Add adds an entity to the cache.
// If an entity with the same ID already exists, it is replaced.
func (c *cache[T]) Add(ctx context.Context, v T) {
	c.Lock()
	defer c.Unlock()
	c.data[v.ID()] = v
	c.add(v.ID())
}

// Update updates an existing entity in the cache with the provided changes.
// UpdatedFields contains the new field values, and removedFields lists fields that should be cleared.
func (c *cache[T]) Update(ctx context.Context, _id primitive.ObjectID, updatedFields T, removedFields []string) {
	c.Lock()
	defer c.Unlock()
	if it, ok := c.data[_id.Hex()]; ok {
		ufv := reflect.ValueOf(updatedFields).Elem()
		uft := ufv.Type()
		itv := reflect.ValueOf(it).Elem()
		c._upd(itv, ufv)
		for _, fieldName := range removedFields {
			for i := 0; i < ufv.NumField(); i++ {
				fieldValue := ufv.Field(i)
				fieldType := uft.Field(i)
				if fieldType.Tag.Get("bson") == fieldName {
					itv.FieldByName(fieldType.Name).Set(reflect.Zero(fieldValue.Type()))
				}
			}
		}
	}
}

func (c *cache[T]) _upd(itv reflect.Value, v reflect.Value) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := t.Field(i)
		if fieldValue.Kind() == reflect.Ptr ||
			fieldValue.Kind() == reflect.Slice ||
			fieldValue.Kind() == reflect.Map {
			if !fieldValue.IsNil() && !fieldValue.IsZero() {
				itv.FieldByName(fieldType.Name).Set(fieldValue)
			}
			continue
		}
		if fieldValue.Kind() == reflect.Array {
			if strings.Compare(fieldValue.Type().String(), "primitive.ObjectID") == 0 {
				if fieldValue.Interface().(primitive.ObjectID).IsZero() {
					continue
				}
			}
		}
		if fieldValue.Kind() == reflect.Struct && fieldType.Tag.Get("bson") == "" {
			c._upd(itv.FieldByName(fieldType.Name), v.Field(i))
			continue
		}
		itv.FieldByName(fieldType.Name).Set(fieldValue)
		continue
	}
}

// Delete ...
func (c *cache[T]) Delete(ctx context.Context, _id primitive.ObjectID) {
	c.Lock()
	defer c.Unlock()
	delete(c.data, _id.Hex())
	c.deleteByID(_id.Hex())
}

func (p *cache[T]) prepareCreate(ctx context.Context, ps reflect.Value) (prepared reflect.Value, err error) {
	switch ps.Kind() {
	case reflect.Ptr:
		if ps.IsNil() {
			prepared = ps
			return
		}
		pr, rerr := p.prepareCreate(ctx, ps.Elem())
		if rerr != nil {
			err = rerr
			return
		}
		prepared = pr
		return prepared.Addr(), nil
	case reflect.Struct:
		_prepared := reflect.New(ps.Type()).Interface()
		prepared = reflect.ValueOf(_prepared).Elem()
		uft := reflect.TypeOf(ps.Interface())
		for i := 0; i < ps.NumField(); i++ {
			if uft.Field(i).Tag.Get("bson") == "-" {
				continue
			}
			if !(ps.Field(i).Kind() == reflect.Map && ps.Field(i).IsNil()) {
				pr, err := p.prepareCreateForStruct(ctx, ps.Field(i), uft.Field(i))
				if err != nil {
					continue
				}
				prepared.Field(i).Set(pr)
			}
		}
	default:
		prepared = ps
	}
	return
}

func (p *cache[T]) prepareCreateForStruct(ctx context.Context, fieldValue reflect.Value, fieldType reflect.StructField) (prepared reflect.Value, err error) {
	switch fieldValue.Kind() {
	case reflect.Ptr:
		if fieldValue.IsNil() {
			prepared = fieldValue
			return
		}
		_prepared, _err := p.prepareCreateForStruct(ctx, fieldValue.Elem(), fieldType)
		if _err != nil {
			err = _err
			return
		}
		prepared = _prepared.Addr()
	case reflect.Array:
		if p.isPrimitiveObjectID(fieldValue.Type()) {
			prepared = fieldValue
			return
		}
		prepared, err = p.setArray(fieldValue)
		if err != nil {
			return
		}
	case reflect.Slice:
		prepared, err = p.setSlice(ctx, fieldValue)
		if err != nil {
			return
		}
	case reflect.Map:
		prepared, err = p.setMap(fieldValue)
		if err != nil {
			return
		}
	case reflect.Struct:
		if p.isDecimalType(fieldValue.Type()) {
			prepared = fieldValue
			return
		}
		prepared, err = p.prepareCreate(ctx, fieldValue)
		if err != nil {
			return
		}
	default:
		prepared = fieldValue
	}
	return
}

func (p *cache[T]) prepareCreateForSlice(ctx context.Context, fieldValue reflect.Value) (prepared reflect.Value, err error) {
	switch fieldValue.Kind() {
	case reflect.Ptr:
		if fieldValue.IsNil() {
			prepared = fieldValue
			return
		}
		_prepared, _err := p.prepareCreateForSlice(ctx, fieldValue.Elem())
		if _err != nil {
			err = _err
			return
		}
		prepared = _prepared.Addr()
	case reflect.Array:
		if p.isPrimitiveObjectID(fieldValue.Type()) {
			prepared = fieldValue
			return
		}
		prepared, err = p.setArray(fieldValue)
		if err != nil {
			return
		}
	case reflect.Slice:
		if p.isJsonRawMessage(fieldValue.Type()) {
			prepared = fieldValue
			return
		}
		prepared, err = p.setSlice(ctx, fieldValue)
		if err != nil {
			return
		}
	case reflect.Map:
		prepared, err = p.setMap(fieldValue)
		if err != nil {
			return
		}
	case reflect.Struct:
		if p.isDecimalType(fieldValue.Type()) {
			prepared = fieldValue
			return
		}
		prepared, err = p.prepareCreate(ctx, fieldValue)
		if err != nil {
			return
		}
	default:
		prepared = fieldValue
	}
	return
}

func (p *cache[T]) setArray(ps reflect.Value) (prepared reflect.Value, err error) {
	prepared = reflect.New(ps.Type()).Elem()
	for i := 0; i < ps.Len(); i++ {
		prepared.Index(i).Set(ps.Index(i))
	}
	return
}

func (p *cache[T]) setSlice(ctx context.Context, ps reflect.Value) (prepared reflect.Value, err error) {
	if ps.IsNil() {
		return ps, nil
	}
	prepared = reflect.MakeSlice(ps.Type(), ps.Len(), ps.Cap())
	for j := 0; j < ps.Len(); j++ {
		_prepared, err := p.prepareCreateForSlice(ctx, ps.Index(j))
		if err != nil {
			return prepared, err
		}
		prepared.Index(j).Set(_prepared)
	}
	return
}

func (p *cache[T]) setMap(ps reflect.Value) (prepared reflect.Value, err error) {
	if !ps.IsNil() {
		prepared = reflect.MakeMap(ps.Type())
		for _, key := range ps.MapKeys() {
			value := ps.MapIndex(key)
			prepared.SetMapIndex(key, value)
		}
	}
	return
}

func (p *cache[T]) isDecimalType(t reflect.Type) bool {
	return strings.Contains(t.String(), "Decimal")
}

func (p *cache[T]) isPrimitiveObjectID(t reflect.Type) bool {
	return strings.Compare(t.String(), "primitive.ObjectID") == 0
}

func (p *cache[T]) isJsonRawMessage(t reflect.Type) bool {
	return strings.Compare(t.String(), "json.RawMessage") == 0
}

// NewCache creates a new Cache instance with the provided initial data.
// The cache maintains entities in memory and provides fast lookup by ID or index.
func NewCache[T d](
	data map[string]T,
) Cache[T] {
	return &cache[T]{
		Idx: Idx{
			maxIdx:       &atomic.Int64{},
			itemsByIndex: map[int64]string{},
			indexByID:    map[string]int64{},
		},
		data: data,
	}
}
