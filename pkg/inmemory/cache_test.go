package inmemory_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/dhlab-tech/go-mongo-platform/pkg/inmemory"
)

var (
	name1          = "test name 1"
	name2          = "test name 2"
	number1        = 1
	number2        = 2
	version1 int64 = 1
	version2 int64 = 2
	slice1         = []string{"_test1_", "_test2_"}
	slice2         = []string{"_test21_", "_test22_", "_test23_"}
	map1           = map[string]string{"1": "_map_1", "2": "_map_2"}
	map2           = map[string]string{"12": "_map_12", "22": "_map_22", "23": "_map_23"}
)

type V struct {
	Id     primitive.ObjectID `bson:"_id"`
	Name   *string            `bson:"name"`
	Number *int               `bson:"number"`
	Slice  []string           `bson:"slice"`
	Map    map[string]string  `bson:"map"`
	Ver    *int64             `bson:"version"`
	S      S                  `bson:"s"`
}

func (v *V) ID() string {
	return v.Id.Hex()
}

func (v *V) Version() *int64 {
	return v.Ver
}

func (v *V) SetDeleted(d bool) {
}

type S struct {
	CS  C
	CSS []C
}

type C struct {
	Name string
}

func TestCache_Update(t *testing.T) {
	c := inmemory.NewCache[*V](0, map[int]string{}, map[string]int{}, map[string]*V{})
	id := primitive.NewObjectIDFromTimestamp(time.Now())
	c.Add(context.Background(), &V{Id: id})
	c.Update(context.Background(), id, &V{
		Name:   &name1,
		Number: &number1,
		Slice:  slice1,
		Map:    map1,
		Ver:    &version1,
	}, []string{})
	v, f := c.Get(context.Background(), id.Hex())
	assert.Equal(t, true, f)
	assert.NotNil(t, v.Name)
	assert.Equal(t, name1, *v.Name)
	assert.NotNil(t, v.Number)
	assert.Equal(t, number1, *v.Number)
	assert.NotNil(t, v.Slice)
	assert.Equal(t, slice1, v.Slice)
	assert.NotNil(t, v.Map)
	assert.Equal(t, map1, v.Map)
	assert.NotNil(t, v.Ver)
	assert.Equal(t, version1, *v.Ver)
	c.Update(context.Background(), id, &V{
		Name:   &name2,
		Number: &number2,
		Slice:  slice2,
		Map:    map2,
	}, []string{"version"})
	v, f = c.Get(context.Background(), id.Hex())
	assert.Equal(t, true, f)
	assert.NotNil(t, v.Name)
	assert.Equal(t, name2, *v.Name)
	assert.NotNil(t, v.Number)
	assert.Equal(t, number2, *v.Number)
	assert.NotNil(t, v.Slice)
	assert.Equal(t, slice2, v.Slice)
	assert.NotNil(t, v.Map)
	assert.Equal(t, map2, v.Map)
	assert.Nil(t, v.Ver)
	c.Update(context.Background(), id, &V{}, []string{})
	v, f = c.Get(context.Background(), id.Hex())
	assert.Equal(t, true, f)
	assert.NotNil(t, v.Name)
	assert.Equal(t, name2, *v.Name)
	assert.NotNil(t, v.Number)
	assert.Equal(t, number2, *v.Number)
	assert.NotNil(t, v.Slice)
	assert.Equal(t, slice2, v.Slice)
	assert.NotNil(t, v.Map)
	assert.Equal(t, map2, v.Map)
	c.Update(context.Background(), id, &V{}, []string{"name", "number", "slice", "map"})
	v, f = c.Get(context.Background(), id.Hex())
	assert.Equal(t, true, f)
	assert.Nil(t, v.Name)
	assert.Nil(t, v.Number)
	assert.Nil(t, v.Slice)
	assert.Nil(t, v.Map)
}

// TestCache_Get проверяем кейсы:
// 1) то что метод возвращает копию хранимых данных
func TestCache_Get(t *testing.T) {
	c := inmemory.NewCache[*V](0, map[int]string{}, map[string]int{}, map[string]*V{})
	id := primitive.NewObjectIDFromTimestamp(time.Now())
	c.Add(context.Background(), &V{
		Id:     id,
		Name:   &name1,
		Number: &number1,
		Slice:  slice1,
		Map:    map1,
		S: S{
			CS: C{
				Name: name1,
			},
		},
	})
	v, f := c.Get(context.Background(), id.Hex())
	assert.Equal(t, true, f)
	assert.NotNil(t, v.Name)
	assert.Equal(t, name1, *v.Name)
	assert.NotNil(t, v.Number)
	assert.Equal(t, number1, *v.Number)
	assert.NotNil(t, v.Slice)
	assert.Equal(t, slice1, v.Slice)
	assert.NotNil(t, v.Map)
	assert.Equal(t, map1, v.Map)
	assert.Equal(t, S{CS: C{Name: name1}}, v.S)
	assert.Equal(t, name1, v.S.CS.Name)
	c.Update(context.Background(), id, &V{S: S{CS: C{Name: name2}}}, []string{})
	v, f = c.Get(context.Background(), id.Hex())
	assert.Equal(t, S{CS: C{Name: name2}}, v.S)
	assert.Equal(t, name2, v.S.CS.Name)
	c.Update(context.Background(), id, &V{}, []string{"name", "number", "slice", "map"})
	assert.Equal(t, true, f)
	assert.NotNil(t, v.Name)
	assert.Equal(t, name1, *v.Name)
	assert.NotNil(t, v.Number)
	assert.Equal(t, number1, *v.Number)
	assert.NotNil(t, v.Slice)
	assert.Equal(t, slice1, v.Slice)
	assert.NotNil(t, v.Map)
	assert.Equal(t, map1, v.Map)
	v, f = c.Get(context.Background(), id.Hex())
	assert.Equal(t, true, f)
	assert.Nil(t, v.Name)
	assert.Nil(t, v.Number)
	assert.Nil(t, v.Slice)
	assert.Nil(t, v.Map)
}
