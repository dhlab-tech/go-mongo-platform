package mongo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/rs/zerolog"
	"github.com/xiyuantang/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrNothingToCreate = errors.New("nothing to create")
	ErrNothingToUpdate = errors.New("nothing to update")
)

// Processor ...
type Processor[T d] struct {
	cache   cache[T]
	creator creator
	updater updater
	remover remover
}

// Create ...
func (p *Processor[T]) Create(ctx context.Context, ps T) (id string, err error) {
	var (
		doc bson.D
		_id primitive.ObjectID
	)
	// создаем все сущности по умолчанию не удаленными
	ps.SetDeleted(false)
	_, doc, err = p.PrepareCreate(ctx, ps)
	if err != nil {
		return
	}
	if len(doc) > 0 {
		_id, err = p.creator.Create(ctx, doc)
		if err != nil {
			return
		}
		id = _id.Hex()
		return
	}
	err = ErrNothingToCreate
	return
}

// Update ...
func (p *Processor[T]) Update(ctx context.Context, ps T) (T, error) {
	var (
		set, unset bson.D
		f          bool
		err        error
	)
	ps, set, unset, err = p.PrepareUpdate(ctx, ps)
	if err != nil {
		return ps, err
	}
	if len(set) > 0 {
		logger := zerolog.Ctx(ctx)
		logger.Debug().Any("set", set).Msg("processor:Update")
		if f, err = p.updater.UpdateOne(ctx, ps.ID(), ps.Version(), set, unset); !f {
			return ps, err
		}
		if err != nil {
			return ps, err
		}
		return ps, nil
	}
	return ps, ErrNothingToUpdate
}

func (p *Processor[T]) Delete(ctx context.Context, id string) (err error) {
	logger := zerolog.Ctx(ctx)
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.Err(err).Str("id", id).Msg("Mongo:Processor:Delete:parse objectID from Hex")
		return
	}
	_, err = p.remover.Remove(ctx, bson.D{
		bson.E{
			Key:   "_id",
			Value: _id,
		},
	})
	if err != nil {
		logger.Err(err).Str("id", id).Msg("Mongo:Processor:Delete:remover:Remove")
	}
	return
}

// PrepareCreate ...
func (p *Processor[T]) PrepareCreate(ctx context.Context, ps T) (prepared T, doc bson.D, err error) {
	logger := zerolog.Ctx(ctx)
	var pr reflect.Value
	in := reflect.ValueOf(ps)
	var _t reflect.Type
	var _v interface{}
	if in.Kind() == reflect.Ptr {
		_v = in.Elem().Interface()
		_t = reflect.TypeOf(in.Elem().Interface())
	} else {
		_v = in.Interface()
		_t = reflect.TypeOf(in.Interface())
	}
	logger.Debug().
		Any("item for create", _v).
		Any("type", _t.Name()).
		Msg("Processor:PrepareCreate")
	pr, doc, err = p.prepareCreate(ctx, in)
	if err != nil {
		return
	}
	if in.Kind() == reflect.Ptr {
		_v = pr.Elem().Interface()
	} else {
		_v = pr.Interface()
	}
	logger.Debug().
		Str("_id", ps.ID()).
		Any("item prepared for create", _v).
		Any("type", _t.Name()).
		Any("doc set", doc).
		Msg("Processor:PrepareCreate")
	return pr.Interface().(T), doc, nil
}

func (p *Processor[T]) prepareCreate(ctx context.Context, ps reflect.Value) (prepared reflect.Value, doc bson.D, err error) {
	switch ps.Kind() {
	case reflect.Ptr:
		if ps.IsNil() {
			prepared = ps
			return
		}
		pr, rd, rerr := p.prepareCreate(ctx, ps.Elem())
		if rerr != nil {
			err = rerr
			return
		}
		prepared = pr
		doc = rd
		return prepared.Addr(), doc, nil
	case reflect.Struct:
		_prepared := reflect.New(ps.Type()).Interface()
		prepared = reflect.ValueOf(_prepared).Elem()
		uft := reflect.TypeOf(ps.Interface())
		for i := 0; i < ps.NumField(); i++ {
			if uft.Field(i).Tag.Get("bson") == "-" {
				continue
			}
			if !(ps.Field(i).Kind() == reflect.Map && ps.Field(i).IsNil()) {
				pr, prDoc, err := p.prepareCreateForStruct(ctx, ps.Field(i), uft.Field(i))
				if err != nil {
					continue
				}
				prepared.Field(i).Set(pr)
				doc = append(doc, prDoc...)
			}
		}
	default:
		prepared = ps
	}
	return
}

func (p *Processor[T]) prepareCreateForStruct(ctx context.Context, fieldValue reflect.Value, fieldType reflect.StructField) (prepared reflect.Value, doc bson.D, err error) {
	switch fieldValue.Kind() {
	case reflect.Ptr:
		if fieldValue.IsNil() {
			prepared = fieldValue
			doc = append(doc, bson.E{
				Key:   fieldType.Tag.Get("bson"),
				Value: nil,
			})
			return
		}
		_prepared, prDoc, _err := p.prepareCreateForStruct(ctx, fieldValue.Elem(), fieldType)
		if _err != nil {
			err = _err
			return
		}
		prepared = _prepared.Addr()
		doc = append(doc, prDoc...)
	case reflect.Array:
		if p.isPrimitiveObjectID(fieldValue.Type()) {
			prepared = fieldValue
			doc = append(doc, bson.E{
				Key:   fieldType.Tag.Get("bson"),
				Value: fieldValue.Interface(),
			})
			return
		}
		var prDoc bson.A
		prepared, prDoc, err = p.setArray(fieldValue)
		if err != nil {
			return
		}
		doc = append(doc, bson.E{
			Key:   fieldType.Tag.Get("bson"),
			Value: prDoc,
		})
	case reflect.Slice:
		var prDoc bson.A
		prepared, prDoc, err = p.setSlice(ctx, fieldValue)
		if err != nil {
			return
		}
		doc = append(doc, bson.E{
			Key:   fieldType.Tag.Get("bson"),
			Value: prDoc,
		})
	case reflect.Map:
		var prDoc bson.D
		prepared, prDoc, err = p.setMap(fieldValue)
		if err != nil {
			return
		}
		doc = append(doc, bson.E{
			Key:   fieldType.Tag.Get("bson"),
			Value: prDoc,
		})
	case reflect.Struct:
		if p.isDecimalType(fieldValue.Type()) {
			switch tp := fieldValue.Interface().(type) {
			case decimal.Decimal:
				prepared = fieldValue
				doc = append(doc, bson.E{
					Key:   fieldType.Tag.Get("bson"),
					Value: tp.String(),
				})
			}
			return
		}
		var prDoc bson.D
		prepared, prDoc, err = p.prepareCreate(ctx, fieldValue)
		if err != nil {
			return
		}
		tag := fieldType.Tag.Get("bson")
		if tag != "" {
			doc = append(doc, bson.E{
				Key:   tag,
				Value: prDoc,
			})
		} else {
			doc = append(doc, prDoc...)
		}
	default:
		prepared = fieldValue
		doc = append(doc, p.set(fieldValue, fieldType))
	}
	return
}

func (p *Processor[T]) prepareCreateForSlice(ctx context.Context, fieldValue reflect.Value) (prepared reflect.Value, doc any, err error) {
	switch fieldValue.Kind() {
	case reflect.Ptr:
		if fieldValue.IsNil() {
			prepared = fieldValue
			doc = nil
			return
		}
		_prepared, prDoc, _err := p.prepareCreateForSlice(ctx, fieldValue.Elem())
		if _err != nil {
			err = _err
			return
		}
		prepared = _prepared.Addr()
		doc = prDoc
	case reflect.Array:
		if p.isPrimitiveObjectID(fieldValue.Type()) {
			prepared = fieldValue
			doc = fieldValue.Interface()
			return
		}
		var prDoc bson.A
		prepared, prDoc, err = p.setArray(fieldValue)
		if err != nil {
			return
		}
		doc = prDoc
	case reflect.Slice:
		var prDoc bson.A
		prepared, prDoc, err = p.setSlice(ctx, fieldValue)
		if err != nil {
			return
		}
		doc = prDoc
	case reflect.Map:
		var prDoc bson.D
		prepared, prDoc, err = p.setMap(fieldValue)
		if err != nil {
			return
		}
		doc = prDoc
	case reflect.Struct:
		if p.isDecimalType(fieldValue.Type()) {
			switch tp := fieldValue.Interface().(type) {
			case decimal.Decimal:
				prepared = fieldValue
				doc = tp.String()
			}
			return
		}
		var prDoc bson.D
		prepared, prDoc, err = p.prepareCreate(ctx, fieldValue)
		if err != nil {
			return
		}
		doc = prDoc
	default:
		prepared = fieldValue
		doc = p.get(fieldValue)
	}
	return
}

func (p *Processor[T]) setArray(ps reflect.Value) (prepared reflect.Value, doc bson.A, err error) {
	doc = bson.A{}
	prepared = reflect.New(ps.Type()).Elem()
	for i := 0; i < ps.Len(); i++ {
		prepared.Index(i).Set(ps.Index(i))
		doc = append(doc, ps.Index(i).Interface())
	}
	return
}

func (p *Processor[T]) setSlice(ctx context.Context, ps reflect.Value) (prepared reflect.Value, doc bson.A, err error) {
	if ps.IsNil() {
		return ps, doc, nil
	}
	prepared = reflect.MakeSlice(ps.Type(), ps.Len(), ps.Cap())
	doc = bson.A{}
	for j := 0; j < ps.Len(); j++ {
		_prepared, prDoc, err := p.prepareCreateForSlice(ctx, ps.Index(j))
		if err != nil {
			return prepared, doc, err
		}
		prepared.Index(j).Set(_prepared)
		doc = append(doc, prDoc)
	}
	return
}

func (p *Processor[T]) isArraysEqual(ctx context.Context, _id string, oldValue, newValue reflect.Value) bool {
	if oldValue.Len() != newValue.Len() {
		return false
	}
	for j := 0; j < oldValue.Len(); j++ {
		if !p.isEquals(ctx, _id, oldValue.Index(j), newValue.Index(j)) {
			return false
		}
	}
	return true
}

func (p *Processor[T]) isSlicesEqual(ctx context.Context, _id string, oldValue, newValue reflect.Value) bool {
	if oldValue.IsNil() && newValue.IsNil() {
		return true
	}
	if (oldValue.IsNil() && !newValue.IsNil()) ||
		(!oldValue.IsNil() && newValue.IsNil()) {
		return false
	}
	if oldValue.Len() != newValue.Len() {
		return false
	}
	for j := 0; j < oldValue.Len(); j++ {
		if !p.isEquals(ctx, _id, oldValue.Index(j), newValue.Index(j)) {
			return false
		}
	}
	return true
}

func (p *Processor[T]) isEqualsStruct(ctx context.Context, _id string, oldValue, newValue reflect.Value) bool {
	switch newValue.Kind() {
	case reflect.Ptr:
		if newValue.IsNil() && oldValue.IsNil() {
			return true
		}
		if (!newValue.IsNil() && oldValue.IsNil()) || (newValue.IsNil() && !oldValue.IsNil()) {
			return false
		}
		return p.isEqualsStruct(ctx, _id, oldValue.Elem(), newValue.Elem())
	case reflect.Struct:
		for i := 0; i < newValue.NumField(); i++ {
			if !p.isEquals(ctx, _id, oldValue.Field(i), newValue.Field(i)) {
				return false
			}
		}
		return true
	}
	return false
}

func (p *Processor[T]) isEquals(ctx context.Context, _id string, oldData, newData reflect.Value) bool {
	logger := zerolog.Ctx(ctx)
	switch newData.Kind() {
	case reflect.Ptr:
		if newData.IsNil() && oldData.IsNil() {
			return true
		}
		if (!newData.IsNil() && oldData.IsNil()) || (newData.IsNil() && !oldData.IsNil()) {
			return false
		}
		return p.isEquals(ctx, _id, oldData.Elem(), newData.Elem())
	case reflect.Array:
		if p.isPrimitiveObjectID(newData.Type()) {
			switch tp := newData.Interface().(type) {
			case primitive.ObjectID:
				switch op := oldData.Interface().(type) {
				case primitive.ObjectID:
					return tp.Hex() == op.Hex()
				}
			}
			return false
		}
		return p.isArraysEqual(ctx, _id, oldData, newData)
	case reflect.Slice:
		if p.isJsonRawMessage(newData.Type()) {
			switch tp := newData.Interface().(type) {
			case json.RawMessage:
				switch op := oldData.Interface().(type) {
				case json.RawMessage:
					if !bytes.Equal(tp, op) {
						logger.Debug().
							Str("_id", _id).
							Any("tp len", len(tp)).
							Any("op len", len(op)).
							Any("newData", newData.Interface()).
							Any("oldData", oldData.Interface()).
							Any("tp", tp).
							Any("op", op).
							Any("tp type", reflect.TypeOf(tp).Name()).
							Any("op type", reflect.TypeOf(op).Name()).
							Msg("Processor:isEquals:JsonRawMessage")
						return false
					}
					return true
				}
			}
			return false
		}
		return p.isSlicesEqual(ctx, _id, oldData, newData)
	case reflect.Map:
		return p.isMapsEqual(ctx, _id, oldData, newData)
	case reflect.Struct:
		if p.isDecimalType(newData.Type()) {
			switch tp := newData.Interface().(type) {
			case decimal.Decimal:
				switch op := oldData.Interface().(type) {
				case decimal.Decimal:
					return tp.Equal(op)
				}
			}
			return false
		}
		return p.isEqualsStruct(ctx, _id, oldData, newData)
	default:
		return oldData.Equal(newData)
	}
}

func (p *Processor[T]) setMap(ps reflect.Value) (prepared reflect.Value, doc bson.D, err error) {
	if !ps.IsNil() {
		prepared = reflect.MakeMap(ps.Type())
		for _, key := range ps.MapKeys() {
			value := ps.MapIndex(key)
			prepared.SetMapIndex(key, value)
			doc = append(doc, bson.E{Key: key.String(), Value: value.Interface()})
		}
	}
	return
}

func (p *Processor[T]) isMapsEqual(ctx context.Context, _id string, oldValue, newValue reflect.Value) bool {
	if oldValue.IsNil() && newValue.IsNil() {
		return true
	}
	if (oldValue.IsNil() && !newValue.IsNil()) ||
		(!oldValue.IsNil() && newValue.IsNil()) {
		return false
	}
	if oldValue.Len() != newValue.Len() {
		return false
	}
	for _, key := range oldValue.MapKeys() {
		if !p.isEquals(ctx, _id, oldValue.MapIndex(key), newValue.MapIndex(key)) {
			return false
		}
	}
	return true
}

func (p *Processor[T]) PrepareUpdate(ctx context.Context, ps T) (prepared T, set bson.D, unset bson.D, err error) {
	logger := zerolog.Ctx(ctx)
	var (
		found bool
	)
	if prepared, found = p.cache.Get(ctx, ps.ID()); !found {
		err = ErrNotFound
		return
	}
	in := reflect.ValueOf(ps)
	var _t reflect.Type
	var _v interface{}
	if in.Kind() == reflect.Ptr {
		_v = in.Elem().Interface()
		_t = reflect.TypeOf(in.Elem().Interface())
	} else {
		_v = in.Interface()
		_t = reflect.TypeOf(in.Interface())
	}
	logger.Debug().
		Any("item for update", _v).
		Any("item from cache", prepared).
		Any("type", _t.Name()).
		Msg("Processor:PrepareUpdate")
	pr, s, err := p.prepareUpdate(ctx, ps.ID(), reflect.ValueOf(ps), reflect.ValueOf(prepared))
	if err != nil {
		return
	}
	if in.Kind() == reflect.Ptr {
		_v = pr.Elem().Interface()
	} else {
		_v = pr.Interface()
	}
	logger.Debug().
		Str("_id", ps.ID()).
		Any("item prepared for update", _v).
		Any("type", _t.Name()).
		Any("doc set", s).
		Msg("Processor:PrepareUpdate")
	return pr.Interface().(T), s, nil, nil
}

func (p *Processor[T]) prepareUpdate(ctx context.Context, _id string, newData, oldData reflect.Value) (prepared reflect.Value, set bson.D, err error) {
	switch newData.Kind() {
	case reflect.Ptr:
		if newData.IsNil() {
			prepared = newData
			return
		}
		pr, rset, rerr := p.prepareUpdate(ctx, _id, newData.Elem(), oldData.Elem())
		if rerr != nil {
			err = rerr
			return
		}
		prepared = pr
		set = rset
		return prepared.Addr(), set, nil
	case reflect.Struct:
		prepared = reflect.ValueOf(reflect.New(newData.Type()).Interface()).Elem()
		uft := reflect.TypeOf(newData.Interface())
		for i := 0; i < newData.NumField(); i++ {
			if uft.Field(i).Tag.Get("bson") == "-" {
				continue
			}
			var (
				prField reflect.Value
				prSet   bson.D
			)
			if !(newData.Field(i).Kind() == reflect.Map && newData.Field(i).IsNil()) {
				prField, prSet, err = p.prepareUpdateForStruct(ctx, _id, newData.Field(i), oldData.Field(i), uft.Field(i))
				if err != nil {
					continue
				}
				prepared.Field(i).Set(prField)
				if len(prSet) > 0 {
					set = append(set, prSet...)
				}
			}
		}
	}
	return
}

func (p *Processor[T]) prepareUpdateForStruct(ctx context.Context, _id string, newValue, oldValue reflect.Value, fieldType reflect.StructField) (prepared reflect.Value, set bson.D, err error) {
	switch newValue.Kind() {
	case reflect.Ptr:
		if newValue.IsNil() && oldValue.IsNil() {
			prepared = newValue
			return
		}
		if newValue.IsNil() && !oldValue.IsNil() {
			prepared = newValue
			set = append(set, bson.E{
				Key:   fieldType.Tag.Get("bson"),
				Value: nil,
			})
			return
		}
		var (
			_prepared reflect.Value
			prSet     bson.D
		)
		if !newValue.IsNil() && !oldValue.IsNil() {
			_prepared, prSet, err = p.prepareUpdateForStruct(ctx, _id, newValue.Elem(), oldValue.Elem(), fieldType)
			if err != nil {
				return
			}
		} else if !newValue.IsNil() && oldValue.IsNil() {
			_prepared, prSet, err = p.prepareCreateForStruct(ctx, newValue.Elem(), fieldType)
			if err != nil {
				return
			}
		}
		prepared = _prepared.Addr()
		if len(prSet) > 0 {
			set = append(set, prSet...)
		}
	case reflect.Array:
		if p.isPrimitiveObjectID(newValue.Type()) {
			prepared = newValue
			if !oldValue.Equal(newValue) {
				set = append(set, bson.E{
					Key:   fieldType.Tag.Get("bson"),
					Value: newValue.Interface(),
				})
			}
			return
		}
		if !p.isSlicesEqual(ctx, _id, oldValue, newValue) {
			var prDoc bson.A
			prepared, prDoc, err = p.setArray(newValue)
			if err != nil {
				return
			}
			set = append(set, bson.E{
				Key:   fieldType.Tag.Get("bson"),
				Value: prDoc,
			})
		} else {
			prepared, _, err = p.setSlice(ctx, newValue)
			if err != nil {
				return
			}
		}
	case reflect.Slice:
		if p.isJsonRawMessage(newValue.Type()) {
			prepared = newValue
			switch tp := newValue.Interface().(type) {
			case json.RawMessage:
				switch op := oldValue.Interface().(type) {
				case json.RawMessage:
					if !bytes.Equal(tp, op) {
						set = append(set, bson.E{
							Key:   fieldType.Tag.Get("bson"),
							Value: newValue.Interface(),
						})
					}
				}
			}
			return
		}
		if !p.isSlicesEqual(ctx, _id, oldValue, newValue) {
			var prDoc bson.A
			prepared, prDoc, err = p.setSlice(ctx, newValue)
			if err != nil {
				return
			}
			set = append(set, bson.E{
				Key:   fieldType.Tag.Get("bson"),
				Value: prDoc,
			})
		} else {
			prepared, _, err = p.setSlice(ctx, newValue)
			if err != nil {
				return
			}
		}
	case reflect.Map:
		if !p.isMapsEqual(ctx, _id, oldValue, newValue) {
			var prDoc bson.D
			prepared, prDoc, err = p.setMap(newValue)
			if err != nil {
				return
			}
			set = append(set, bson.E{
				Key:   fieldType.Tag.Get("bson"),
				Value: prDoc,
			})
		} else {
			prepared, _, err = p.setMap(newValue)
			if err != nil {
				return
			}
		}
	case reflect.Struct:
		if p.isDecimalType(newValue.Type()) {
			switch tp := newValue.Interface().(type) {
			case decimal.Decimal:
				prepared = newValue
				switch op := oldValue.Interface().(type) {
				case decimal.Decimal:
					if !tp.Equal(op) {
						set = append(set, bson.E{
							Key:   fieldType.Tag.Get("bson"),
							Value: tp.String(),
						})
					}
				}
			}
			return
		}
		if !p.isEqualsStruct(ctx, _id, oldValue, newValue) {
			var prSet bson.D
			prepared, prSet, err = p.prepareCreate(ctx, newValue)
			if err != nil {
				return
			}
			if len(prSet) > 0 {
				tag := fieldType.Tag.Get("bson")
				if tag != "" {
					set = append(set, bson.E{
						Key:   tag,
						Value: prSet,
					})
				} else {
					set = append(set, prSet...)
				}
			}
		} else {
			prepared = newValue
		}
	default:
		if !oldValue.Equal(newValue) {
			prepared = newValue
			set = append(set, p.set(newValue, fieldType))
		} else {
			prepared = newValue
		}
	}
	return
}

func (p *Processor[T]) set(fieldValue reflect.Value, fieldType reflect.StructField) bson.E {
	switch fieldValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return bson.E{Key: fieldType.Tag.Get("bson"), Value: fieldValue.Int()}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return bson.E{Key: fieldType.Tag.Get("bson"), Value: fieldValue.Uint()}
	case reflect.Bool:
		return bson.E{Key: fieldType.Tag.Get("bson"), Value: fieldValue.Bool()}
	case reflect.Float32, reflect.Float64:
		return bson.E{Key: fieldType.Tag.Get("bson"), Value: fieldValue.Float()}
	case reflect.String:
		return bson.E{Key: fieldType.Tag.Get("bson"), Value: fieldValue.String()}
	default:
		return bson.E{Key: fieldType.Tag.Get("bson"), Value: fieldValue.Interface()}
	}
}

func (p *Processor[T]) get(fieldValue reflect.Value) any {
	switch fieldValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fieldValue.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fieldValue.Uint()
	case reflect.Bool:
		return fieldValue.Bool()
	case reflect.Float32, reflect.Float64:
		return fieldValue.Float()
	case reflect.String:
		return fieldValue.String()
	default:
		return fieldValue.Interface()
	}
}

func (p *Processor[T]) isDecimalType(t reflect.Type) bool {
	return strings.Contains(t.String(), "Decimal")
}

func (p *Processor[T]) isPrimitiveObjectID(t reflect.Type) bool {
	return strings.Compare(t.String(), "primitive.ObjectID") == 0
}

func (p *Processor[T]) isJsonRawMessage(t reflect.Type) bool {
	return strings.Compare(t.String(), "json.RawMessage") == 0
}

// NewProcessor ...
func NewProcessor[T d](
	cache cache[T],
	creator creator,
	updater updater,
	remover remover,
) *Processor[T] {
	return &Processor[T]{
		cache:   cache,
		creator: creator,
		updater: updater,
		remover: remover,
	}
}
