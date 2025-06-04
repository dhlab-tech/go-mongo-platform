package inmemory

import (
	"context"
	"reflect"
	"strings"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

// All ...
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

// Get ...
func (c *cache[T]) Get(ctx context.Context, id string) (r T, f bool) {
	logger := zerolog.Ctx(ctx)
	c.RLock()
	defer c.RUnlock()
	if d, found := c.data[id]; found {
		r = reflect.New(reflect.TypeOf(r).Elem()).Interface().(T)
		v1 := reflect.ValueOf(d).Elem()
		v2 := reflect.ValueOf(r).Elem()
		for i := 0; i < v1.NumField(); i++ {
			if v1.Field(i).Kind() == reflect.Ptr ||
				v1.Field(i).Kind() == reflect.Slice ||
				v1.Field(i).Kind() == reflect.Map {
				if !v1.Field(i).IsNil() {
					v2.Field(i).Set(v1.Field(i))
				}
				continue
			}
			v2.Field(i).Set(v1.Field(i))
		}
		f = found
		logger.Debug().Any("cached item", r).Msg("Cache:Get")
		return
	}
	return
}

// GetByIndex ...
func (c *cache[T]) GetByIndex(ctx context.Context, idx int) (r T, f bool) {
	var (
		id string
	)
	c.RLock()
	defer c.RUnlock()
	if id, f = c.GetIDByIndex(idx); f {
		if m, found := c.data[id]; found {
			r = reflect.New(reflect.TypeOf(r).Elem()).Interface().(T)
			v1 := reflect.ValueOf(m).Elem()
			v2 := reflect.ValueOf(r).Elem()
			for i := 0; i < v1.NumField(); i++ {
				if v1.Field(i).Kind() == reflect.Ptr ||
					v1.Field(i).Kind() == reflect.Slice ||
					v1.Field(i).Kind() == reflect.Map {
					if !v1.Field(i).IsNil() {
						v2.Field(i).Set(v1.Field(i))
					}
					continue
				}
				v2.Field(i).Set(v1.Field(i))
			}
			f = found
			return
		}
		c.deleteByID(id)
	}
	c.deleteByIdx(idx)
	return
}

// Add ...
func (c *cache[T]) Add(ctx context.Context, v T) {
	c.Lock()
	defer c.Unlock()
	c.data[v.ID()] = v
	c.add(v.ID())
}

// Update ...
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

// NewCache ...
func NewCache[T d](
	maxIdx int,
	itemsByIndex map[int]string,
	indexByID map[string]int,
	data map[string]T,
) Cache[T] {
	return &cache[T]{
		Idx: Idx{
			maxIdx:       maxIdx,
			itemsByIndex: itemsByIndex,
			indexByID:    indexByID,
		},
		data: data,
	}
}
