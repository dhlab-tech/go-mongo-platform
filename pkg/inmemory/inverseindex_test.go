package inmemory_test

import (
	"context"
	"testing"

	"github.com/dhlab-tech/go-mongo-platform/pkg/inmemory"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	parent1 = "test_parent_1"
	parent2 = "test_parent_2"
)

type Image struct {
	D
	P
	R
	Name   *string `json:"name" bson:"name" indexes:"sorted:title:from,suffix:title:from"`                   // название файла
	Orig   *string `json:"orig" bson:"orig" indexes:"inverse_unique:origWidthHeight:from,inverse:orig:from"` // id оригинального изображения
	Width  *int    `json:"width" bson:"width" indexes:"inverse_unique:origWidthHeight:from"`                 // ширина изображения
	Height *int    `json:"height" bson:"height" indexes:"inverse_unique:origWidthHeight:from"`               // высота изображения
	Mime   *string `json:"mime" bson:"mime"`
	Ext    *string `json:"ext" bson:"ext"`
}

type D struct {
	Id      primitive.ObjectID `bson:"_id"`
	V       *int64             `bson:"version"`
	Deleted *bool              `bson:"deleted"`
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

type P struct {
	Properties []Property `json:"properties" bson:"properties"`
}

type Property struct {
	Tags  string `json:"tags" bson:"tags"`
	Value string `json:"value" bson:"value"`
}

type R struct {
	Parent *string `json:"parent" bson:"parent" indexes:"inverse:parent_id:from"`
}

func TestInverseIndex_ParentUpdate(t *testing.T) {
	c := inmemory.NewCache[*Image](make(map[string]*Image))
	idx := inmemory.NewInverseIndex(map[string][]string{}, []string{}, c, []string{"R+Parent"}, nil)
	img := Image{R: R{Parent: &parent1}}
	idx.Add(context.Background(), &img)
	c.Add(context.Background(), &img)
	ids := idx.Get(context.Background(), &parent1)
	assert.Equal(t, []string{img.ID()}, ids)
	img2 := Image{R: R{Parent: &parent2}}
	idx.Update(context.Background(), img.Id, &img2, nil)
	c.Add(context.Background(), &img2)
	ids = idx.Get(context.Background(), &parent1)
	assert.Equal(t, []string{}, ids)
	ids = idx.Get(context.Background(), &parent2)
	assert.Equal(t, []string{img.ID()}, ids)
}
