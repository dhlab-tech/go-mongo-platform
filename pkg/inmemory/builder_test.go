package inmemory

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DocSetTitle struct {
	D
	CatalogID *string `bson:"catalog_id" indexes:"inverse:catalog_id:from,inverse_unique:catalogItem:from"`
	ItemID    *string `bson:"item_id" indexes:"inverse:item_id:from,inverse_unique:catalogItem:from"`
	Title     *string `bson:"title" indexes:"suffix:title:from,sorted:title:from"`
}

type Image struct {
	D
	Name   *string `json:"name" bson:"name"`                                                                 // название файла
	Orig   *string `json:"orig" bson:"orig" indexes:"inverse_unique:origWidthHeight:from,inverse:orig:from"` // id оригинального изображения
	Width  *int    `json:"width" bson:"width" indexes:"inverse_unique:origWidthHeight:from"`                 // ширина изображения
	Height *int    `json:"height" bson:"height" indexes:"inverse_unique:origWidthHeight:from"`               // высота изображения
	Mime   *string `json:"mime" bson:"mime"`
	Ext    *string `json:"ext" bson:"ext"`
}

type D struct {
	Id      primitive.ObjectID `json:"_id" bson:"_id"`
	V       *int64             `json:"version" bson:"version"`
	Deleted *bool              `json:"deleted" bson:"deleted" indexes:"inverse:deleted:from"`
}

func (v *D) ID() string {
	if v.Id.IsZero() {
		v.Id = primitive.NewObjectID()
	}
	return v.Id.Hex()
}

func (v *D) Version() *int64 {
	return v.V
}

func (v *D) IsDeleted() bool {
	return v.Deleted != nil && *v.Deleted
}

func (v *D) SetDeleted(d bool) {
	_d := d
	v.Deleted = &_d
}

type FT struct {
	from []string
	to   string
}

var (
	emptyTo         = ""
	boolFalse       = false
	version   int64 = 100
	catalogID       = "test_cat"
	itemID          = "test_item"
	title           = "Название на русском"
)

var expectedIdxsForDocSetTitle = map[string]map[string]FT{
	"suffix": {
		"title": {
			from: []string{"Title"},
			to:   emptyTo,
		},
	},
	"sorted": {
		"title": {
			from: []string{"Title"},
			to:   emptyTo,
		},
	},
	"inverse": {
		"deleted": {
			from: []string{"D+Deleted"},
			to:   emptyTo,
		},
		"item_id": {
			from: []string{"ItemID"},
			to:   emptyTo,
		},
		"catalog_id": {
			from: []string{"CatalogID"},
			to:   emptyTo,
		},
	},
	"inverse_unique": {
		"catalogItem": {
			from: []string{"CatalogID", "ItemID"},
			to:   emptyTo,
		},
	},
}

var expectedIdxsForImage = map[string]map[string]FT{
	"inverse": {
		"deleted": {
			from: []string{"D+Deleted"},
			to:   emptyTo,
		},
		"orig": {
			from: []string{"Orig"},
			to:   emptyTo,
		},
	},
	"inverse_unique": {
		"origWidthHeight": {
			from: []string{"Orig", "Width", "Height"},
			to:   emptyTo,
		},
	},
}

func TestBuilder_prepareIdxs_for_DocSetTitle(t *testing.T) {
	var doc DocSetTitle
	for indexType, idx := range prepareIdxs(reflect.ValueOf(doc)) {
		for indexName, _idx := range idx {
			assert.Equal(t, expectedIdxsForDocSetTitle[indexType][indexName].from, _idx.from)
			assert.Equal(t, expectedIdxsForDocSetTitle[indexType][indexName].to, _idx.to)
		}
	}
}

func TestBuilder_prepareIdxs_for_Image(t *testing.T) {
	var image Image
	for indexType, idx := range prepareIdxs(reflect.ValueOf(image)) {
		for indexName, _idx := range idx {
			assert.Equal(t, expectedIdxsForImage[indexType][indexName].from, _idx.from)
			assert.Equal(t, expectedIdxsForImage[indexType][indexName].to, _idx.to)
		}
	}
}

func TestInverseUniqueIndex(t *testing.T) {
	c := NewCache[*DocSetTitle](0, map[int]string{}, map[string]int{}, map[string]*DocSetTitle{})
	var to *string
	if expectedIdxsForDocSetTitle["inverse_unique"]["catalogItem"].to != "" {
		_to := expectedIdxsForDocSetTitle["inverse_unique"]["catalogItem"].to
		to = &_to
	}
	iui := NewInverseUniqIndex(map[string]string{}, c, expectedIdxsForDocSetTitle["inverse_unique"]["catalogItem"].from, to)
	_id := primitive.NewObjectID()
	toIndex := DocSetTitle{
		D: D{
			Id:      _id,
			Deleted: &boolFalse,
			V:       &version,
		},
		CatalogID: &catalogID,
		ItemID:    &itemID,
		Title:     &title,
	}
	iui.Add(context.Background(), &toIndex)
	id, found := iui.Get(context.Background(), catalogID, itemID)
	assert.Equal(t, true, found)
	assert.Equal(t, _id.Hex(), id)
}
