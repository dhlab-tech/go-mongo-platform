package mongo_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xiyuantang/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/dhlab-tech/go-mongo-platform/pkg/inmemory"
	"github.com/dhlab-tech/go-mongo-platform/pkg/mongo"
)

var (
	name1              = "test name 1"
	name2              = "test name 2"
	number1    int64   = 1
	number2    int64   = 2
	unumber1   uint64  = 1
	unumber2   uint64  = 2
	bool1              = true
	bool2              = false
	float1     float64 = 1.12
	float2     float64 = 1.24
	slice1             = []string{"_test1_", "_test2_"}
	slice1_res         = []interface{}{"_test1_", "_test2_"}
	slice2             = []string{"_test21_", "_test22_", "_test23_"}
	slice2_res         = []interface{}{"_test21_", "_test22_", "_test23_"}
	map1               = map[string]interface{}{"1": "_map_1", "2": "_map_2"}
	map2               = map[string]interface{}{"12": "_map_12", "22": "_map_22", "23": "_map_23"}
	s                  = S{Name: nil}
	s1                 = S{Name: &name1}
	s2                 = S{Name: &name2}
	s3                 = S{Name: &name2, G: G{In: C{Name: &name1}}}
	g1                 = G{In: C{Name: &name1}, Out: &C{Name: &name2}}
	g2                 = G{In: C{Name: nil}, Out: nil}
	version1   int64   = 1
	version2   int64   = 2
	id0        uint8   = 0
	id1        uint8   = 1
	dec1, _            = decimal.NewFromString("100.15")
	dec2, _            = decimal.NewFromString("123.45")
	width1             = 100
	height1            = 100
)

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

type Config struct {
	D
	Schema   *string           `json:"schema" bson:"schema"`
	Host     *string           `json:"host" bson:"host" indexes:"inverse_unique:host:from"`
	Path     map[string]string `json:"path" bson:"path"`
	HTMLRoot *string           `json:"htmlRoot" bson:"htmlRoot"`
	CSSRoot  *string           `json:"cssRoot" bson:"cssRoot"`
	SJSRoot  *string           `json:"sjsRoot" bson:"sjsRoot"`
	JSRoot   *string           `json:"jsRoot" bson:"jsRoot"`
	JSVars   map[string]string `json:"jsVars" bson:"jsVars"`
	PageSize *int              `json:"pageSize" bson:"pageSize"`
	Menu     map[string]string `json:"menu" bson:"menu"`
	MSConfig *string           `json:"msConfig" bson:"msConfig"`
	Fits     []Fit             `json:"fits" bson:"fits"`
	Labels   []string          `json:"labels" bson:"labels"`
}

type Fit struct {
	Width  int `json:"width" bson:"width"`
	Height int `json:"height" bson:"height"`
}

type V struct {
	D
	Name     *string                `bson:"name"`
	Number   *int64                 `bson:"number"`
	UNumber  *uint64                `bson:"unumber"`
	Bool     *bool                  `bson:"bool"`
	Float    *float64               `bson:"float"`
	Slice    []string               `bson:"slice"`
	Map      map[string]interface{} `bson:"map"`
	Map2     map[string]string      `bson:"map2"`
	S        S                      `bson:"s"`
	S1       S                      `bson:"s1"`
	S2       S                      `bson:"s2"`
	PName    string                 `bson:"pname"`
	PNumber  int64                  `bson:"pnumber"`
	PUNumber uint64                 `bson:"punumber"`
	PBool    bool                   `bson:"pbool"`
	PFloat   float64                `bson:"pfloat"`
	SP1      *S                     `bson:"spointer1"`
	GP1      *G                     `bson:"gpointer1"`
	GP2      *G                     `bson:"gpointer2"`
	GP3      *G                     `bson:"gpointer3"`
	SS       []S                    `bson:"sslice1"`
	Dec1     decimal.Decimal        `bson:"dec1"`
	Dec2     decimal.Decimal        `bson:"dec2"`
	JS       json.RawMessage        `bson:"js"`
}

type S struct {
	Name *string `bson:"name"`
	G    G       `bson:"gg"`
}

type G struct {
	In  C  `bson:"in"`
	Out *C `bson:"out"`
}

type C struct {
	Name *string `bson:"name"`
}

func TestProcessor_PrepareCreate(t *testing.T) {
	p := mongo.NewProcessor[*V](nil, nil, nil, nil)
	id := primitive.ObjectID{id1, id0, id0, id0, id0, id0, id0, id0, id0, id0, id0, id0}
	v := &V{
		D: D{
			Id: id,
		},
		Name:     &name1,
		Number:   &number1,
		UNumber:  &unumber1,
		Bool:     &bool1,
		Float:    &float1,
		Slice:    slice1,
		Map:      map1,
		S:        s1,
		PName:    name1,
		PNumber:  number1,
		PUNumber: unumber1,
		PBool:    bool1,
		PFloat:   float1,
		SS:       []S{{Name: &name1}, {Name: &name2}},
		JS:       json.RawMessage{100, 101, 102, 103},
	}
	pr, doc, err := p.PrepareCreate(context.Background(), v)
	assert.NoError(t, err)
	assert.Equal(t, &V{
		D: D{
			Id: id,
		},
		Name:     &name1,
		Number:   &number1,
		UNumber:  &unumber1,
		Bool:     &bool1,
		Float:    &float1,
		Slice:    slice1,
		Map:      map1,
		S:        s1,
		PName:    name1,
		PNumber:  number1,
		PUNumber: unumber1,
		PBool:    bool1,
		PFloat:   float1,
		SS:       []S{{Name: &name1}, {Name: &name2}},
		JS:       json.RawMessage{100, 101, 102, 103},
	}, pr)
	assert.Equal(t, bson.D{
		bson.E{Key: "_id", Value: primitive.ObjectID{id1, id0, id0, id0, id0, id0, id0, id0, id0, id0, id0, id0}},
		bson.E{Key: "version", Value: nil},
		bson.E{Key: "deleted", Value: nil},
		bson.E{Key: "name", Value: name1},
		bson.E{Key: "number", Value: number1},
		bson.E{Key: "unumber", Value: unumber1},
		bson.E{Key: "bool", Value: bool1},
		bson.E{Key: "float", Value: float1},
		bson.E{Key: "slice", Value: bson.A{"_test1_", "_test2_"}},
		bson.E{Key: "map", Value: bson.D{bson.E{Key: "1", Value: "_map_1"}, bson.E{Key: "2", Value: "_map_2"}}},
		bson.E{Key: "s", Value: bson.D{
			bson.E{Key: "name", Value: name1},
			bson.E{Key: "gg", Value: bson.D{
				bson.E{Key: "in", Value: bson.D{
					bson.E{Key: "name", Value: nil},
				}},
				bson.E{Key: "out", Value: nil},
			}},
		}},
		bson.E{Key: "s1", Value: bson.D{
			bson.E{Key: "name", Value: nil},
			bson.E{Key: "gg", Value: bson.D{
				bson.E{Key: "in", Value: bson.D{
					bson.E{Key: "name", Value: nil},
				}},
				bson.E{Key: "out", Value: nil},
			}},
		}},
		bson.E{Key: "s2", Value: bson.D{
			bson.E{Key: "name", Value: nil},
			bson.E{Key: "gg", Value: bson.D{
				bson.E{Key: "in", Value: bson.D{
					bson.E{Key: "name", Value: nil},
				}},
				bson.E{Key: "out", Value: nil},
			}},
		}},
		bson.E{Key: "pname", Value: name1},
		bson.E{Key: "pnumber", Value: number1},
		bson.E{Key: "punumber", Value: unumber1},
		bson.E{Key: "pbool", Value: bool1},
		bson.E{Key: "pfloat", Value: float1},
		bson.E{Key: "spointer1", Value: nil},
		bson.E{Key: "gpointer1", Value: nil},
		bson.E{Key: "gpointer2", Value: nil},
		bson.E{Key: "gpointer3", Value: nil},
		bson.E{Key: "sslice1", Value: bson.A{
			bson.D{
				bson.E{Key: "name", Value: name1},
				bson.E{Key: "gg", Value: bson.D{
					bson.E{Key: "in", Value: bson.D{
						bson.E{Key: "name", Value: nil},
					}},
					bson.E{Key: "out", Value: nil},
				}},
			},
			bson.D{
				bson.E{Key: "name", Value: name2},
				bson.E{Key: "gg", Value: bson.D{
					bson.E{Key: "in", Value: bson.D{
						bson.E{Key: "name", Value: nil},
					}},
					bson.E{Key: "out", Value: nil},
				}},
			},
		}},
		bson.E{Key: "dec1", Value: "0"},
		bson.E{Key: "dec2", Value: "0"},
		bson.E{Key: "js", Value: bson.A{uint64(100), uint64(101), uint64(102), uint64(103)}},
	}, doc)
}

func TestProcessor_PrepareUpdate(t *testing.T) {
	id := primitive.ObjectID{id1, id0, id0, id0, id0, id0, id0, id0, id0, id0, id0, id0}
	c := inmemory.NewCache[*V](map[string]*V{})
	c.Add(context.Background(), &V{
		D: D{
			Id: id,
			V:  &version2,
		},
		Name:     &name1,
		Number:   &number1,
		UNumber:  &unumber1,
		Bool:     &bool1,
		Float:    &float1,
		Slice:    slice1,
		Map:      map1,
		S:        s1,
		PName:    name1,
		PNumber:  number1,
		PUNumber: unumber1,
		PBool:    bool1,
		PFloat:   float1,
		GP2:      &g1,
		GP3:      &g1,
		SS:       []S{{Name: &name1, G: G{In: C{Name: &name1}, Out: &C{Name: &name2}}}, {Name: &name2}},
		Dec2:     dec1,
	})
	p := mongo.NewProcessor[*V](c, nil, nil, nil)
	pr, set, _, err := p.PrepareUpdate(context.Background(), &V{
		D: D{
			Id: id,
			V:  &version2,
		},
		Name:     &name2,
		Number:   &number2,
		UNumber:  nil,
		Bool:     &bool2,
		Float:    &float2,
		Slice:    slice2,
		Map:      map2,
		S:        s2,
		S1:       s1,
		S2:       s,
		PName:    name2,
		PNumber:  number2,
		PUNumber: unumber2,
		PBool:    bool2,
		PFloat:   float2,
		SP1:      &s1,
		GP1:      &g1,
		GP3:      &g2,
		SS:       []S{{Name: &name2, G: G{In: C{Name: &name2}, Out: &C{Name: &name1}}}, {Name: &name1}},
		Dec1:     dec1,
		Dec2:     dec1,
	})
	assert.NoError(t, err)
	assert.Equal(t, &V{
		D: D{
			Id: id,
			V:  &version2,
		},
		Name:     &name2,
		Number:   &number2,
		UNumber:  nil,
		Bool:     &bool2,
		Float:    &float2,
		Slice:    slice2,
		Map:      map2,
		S:        s2,
		S1:       s1,
		S2:       s,
		PName:    name2,
		PNumber:  number2,
		PUNumber: unumber2,
		PBool:    bool2,
		PFloat:   float2,
		SP1:      &s1,
		GP1:      &g1,
		GP3:      &g2,
		SS:       []S{{Name: &name2, G: G{In: C{Name: &name2}, Out: &C{Name: &name1}}}, {Name: &name1}},
		Dec1:     dec1,
		Dec2:     dec1,
	}, pr)
	assert.Equal(t, bson.D{
		bson.E{Key: "name", Value: name2},
		bson.E{Key: "number", Value: number2},
		bson.E{Key: "unumber", Value: nil},
		bson.E{Key: "bool", Value: bool2},
		bson.E{Key: "float", Value: float2},
		bson.E{Key: "slice", Value: bson.A{"_test21_", "_test22_", "_test23_"}},
		bson.E{Key: "map", Value: bson.D{bson.E{Key: "12", Value: "_map_12"}, bson.E{Key: "22", Value: "_map_22"}, bson.E{Key: "23", Value: "_map_23"}}},
		bson.E{Key: "s", Value: bson.D{
			bson.E{Key: "name", Value: name2},
			bson.E{Key: "gg", Value: bson.D{
				bson.E{Key: "in", Value: bson.D{
					bson.E{Key: "name", Value: nil},
				}},
				bson.E{Key: "out", Value: nil},
			}},
		}},
		bson.E{Key: "s1", Value: bson.D{
			bson.E{Key: "name", Value: name1},
			bson.E{Key: "gg", Value: bson.D{
				bson.E{Key: "in", Value: bson.D{
					bson.E{Key: "name", Value: nil},
				}},
				bson.E{Key: "out", Value: nil},
			}},
		}},
		bson.E{Key: "pname", Value: name2},
		bson.E{Key: "pnumber", Value: number2},
		bson.E{Key: "punumber", Value: unumber2},
		bson.E{Key: "pbool", Value: bool2},
		bson.E{Key: "pfloat", Value: float2},
		bson.E{Key: "spointer1", Value: bson.D{
			bson.E{Key: "name", Value: name1},
			bson.E{Key: "gg", Value: bson.D{
				bson.E{Key: "in", Value: bson.D{
					bson.E{Key: "name", Value: nil},
				}},
				bson.E{Key: "out", Value: nil},
			}},
		}},
		bson.E{Key: "gpointer1", Value: bson.D{bson.E{Key: "in", Value: bson.D{bson.E{Key: "name", Value: name1}}}, bson.E{Key: "out", Value: bson.D{bson.E{Key: "name", Value: name2}}}}},
		bson.E{Key: "gpointer2", Value: nil},
		bson.E{Key: "gpointer3", Value: bson.D{bson.E{Key: "in", Value: bson.D{bson.E{Key: "name", Value: nil}}}, bson.E{Key: "out", Value: nil}}},
		bson.E{Key: "sslice1", Value: bson.A{
			bson.D{
				bson.E{Key: "name", Value: name2},
				bson.E{Key: "gg", Value: bson.D{
					bson.E{Key: "in", Value: bson.D{
						bson.E{Key: "name", Value: name2},
					}},
					bson.E{Key: "out", Value: bson.D{
						bson.E{Key: "name", Value: name1},
					},
					},
				}},
			},
			bson.D{
				bson.E{Key: "name", Value: name1},
				bson.E{Key: "gg", Value: bson.D{
					bson.E{Key: "in", Value: bson.D{
						bson.E{Key: "name", Value: nil},
					}},
					bson.E{Key: "out", Value: nil},
				}},
			},
		}},
		bson.E{Key: "dec1", Value: "100.15"},
	}, set)
}

func TestProcessor_PrepareUpdate_Empty(t *testing.T) {
	id := primitive.ObjectID{id1, id0, id0, id0, id0, id0, id0, id0, id0, id0, id0, id0}
	c := inmemory.NewCache[*V](map[string]*V{})
	c.Add(context.Background(), &V{
		D: D{
			Id: id,
			V:  &version1,
		},
		Name:     &name1,
		Number:   &number1,
		UNumber:  &unumber1,
		Bool:     &bool1,
		Float:    &float1,
		Slice:    slice1,
		Map:      map1,
		S:        s3,
		PName:    name1,
		PNumber:  number1,
		PUNumber: unumber1,
		PBool:    bool1,
		PFloat:   float1,
		GP2:      &g1,
		GP3:      &g1,
		SS:       []S{{Name: &name1}, {Name: &name2}},
		Dec1:     dec1,
	})
	p := mongo.NewProcessor[*V](c, nil, nil, nil)
	pr, set, _, err := p.PrepareUpdate(context.Background(), &V{
		D: D{
			Id: id,
			V:  &version1,
		},
		Name:     &name1,
		Number:   &number1,
		UNumber:  &unumber1,
		Bool:     &bool1,
		Float:    &float1,
		Slice:    slice1,
		Map:      map1,
		S:        s3,
		PName:    name1,
		PNumber:  number1,
		PUNumber: unumber1,
		PBool:    bool1,
		PFloat:   float1,
		GP2:      &g1,
		GP3:      &g1,
		SS:       []S{{Name: &name1}, {Name: &name2}},
		Dec1:     dec1,
	})
	assert.NoError(t, err)
	assert.Equal(t, &V{
		D: D{
			Id: id,
			V:  &version1,
		},
		Name:     &name1,
		Number:   &number1,
		UNumber:  &unumber1,
		Bool:     &bool1,
		Float:    &float1,
		Slice:    slice1,
		Map:      map1,
		S:        s3,
		PName:    name1,
		PNumber:  number1,
		PUNumber: unumber1,
		PBool:    bool1,
		PFloat:   float1,
		GP2:      &g1,
		GP3:      &g1,
		SS:       []S{{Name: &name1}, {Name: &name2}},
		Dec1:     dec1,
	}, pr)
	var empty bson.D
	assert.Equal(t, empty, set)
}

func TestProcessor_PrepareUpdate_Slice(t *testing.T) {
	id := primitive.ObjectID{id1, id0, id0, id0, id0, id0, id0, id0, id0, id0, id0, id0}
	c := inmemory.NewCache[*V](map[string]*V{})
	c.Add(context.Background(), &V{
		D: D{
			Id: id,
			V:  &version1,
		},
		SS: []S{{Name: &name1}, {Name: &name2}},
	})
	p := mongo.NewProcessor[*V](c, nil, nil, nil)
	pr, set, _, err := p.PrepareUpdate(context.Background(), &V{
		D: D{
			Id: id,
			V:  &version1,
		},
		SS: []S{},
	})
	assert.NoError(t, err)
	assert.Equal(t, &V{
		D: D{
			Id: id,
			V:  &version1,
		},
		SS: []S{},
	}, pr)
	assert.Equal(t, bson.D{
		bson.E{Key: "sslice1", Value: bson.A{}},
	}, set)
}

func TestProcessor_PrepareUpdate_Config(t *testing.T) {
	id := primitive.ObjectID{id1, id0, id0, id0, id0, id0, id0, id0, id0, id0, id0, id0}
	c := inmemory.NewCache[*Config](map[string]*Config{})
	c.Add(context.Background(), &Config{
		D: D{
			Id: id,
			V:  &version1,
		},
	})
	p := mongo.NewProcessor[*Config](c, nil, nil, nil)
	pr, set, _, err := p.PrepareUpdate(context.Background(), &Config{
		D: D{
			Id: id,
			V:  &version1,
		},
		Fits: []Fit{
			{Width: 64, Height: 85},
			{Width: 172, Height: 230},
			{Width: 400, Height: 533},
			{Width: 600, Height: 800},
		},
		Labels: []string{"__test"},
	})
	assert.NoError(t, err)
	assert.Equal(t, &Config{
		D: D{
			Id: id,
			V:  &version1,
		},
		Fits: []Fit{
			{Width: 64, Height: 85},
			{Width: 172, Height: 230},
			{Width: 400, Height: 533},
			{Width: 600, Height: 800},
		},
		Labels: []string{"__test"},
	}, pr)
	assert.Equal(t, bson.D{
		bson.E{Key: "fits", Value: bson.A{
			bson.D{
				bson.E{Key: "width", Value: int64(64)},
				bson.E{Key: "height", Value: int64(85)},
			},
			bson.D{
				bson.E{Key: "width", Value: int64(172)},
				bson.E{Key: "height", Value: int64(230)},
			},
			bson.D{
				bson.E{Key: "width", Value: int64(400)},
				bson.E{Key: "height", Value: int64(533)},
			},
			bson.D{
				bson.E{Key: "width", Value: int64(600)},
				bson.E{Key: "height", Value: int64(800)},
			},
		}},
		bson.E{Key: "labels", Value: bson.A{
			"__test",
		}},
	}, set)
}

func TestProcessor_PrepareUpdate_Config_Labels(t *testing.T) {
	id := primitive.ObjectID{id1, id0, id0, id0, id0, id0, id0, id0, id0, id0, id0, id0}
	c := inmemory.NewCache[*Config](map[string]*Config{})
	c.Add(context.Background(), &Config{
		D: D{
			Id: id,
			V:  &version1,
		},
		Labels: []string{"__test1", "__test2"},
	})
	p := mongo.NewProcessor[*Config](c, nil, nil, nil)
	pr, set, _, err := p.PrepareUpdate(context.Background(), &Config{
		D: D{
			Id: id,
			V:  &version1,
		},
		Labels: []string{"__test1", "__test2", "__test3"},
	})
	assert.NoError(t, err)
	assert.Equal(t, &Config{
		D: D{
			Id: id,
			V:  &version1,
		},
		Labels: []string{"__test1", "__test2", "__test3"},
	}, pr)
	assert.Equal(t, bson.D{
		bson.E{Key: "labels", Value: bson.A{
			"__test1", "__test2", "__test3",
		}},
	}, set)
}

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

func TestProcessor_PrepareUpdate_Delete(t *testing.T) {
	id := primitive.ObjectID{id1, id0, id0, id0, id0, id0, id0, id0, id0, id0, id0, id0}
	c := inmemory.NewCache[*Image](map[string]*Image{})
	im := Image{
		D: D{
			Id:      id,
			V:       &version1,
			Deleted: &bool2,
		},
		Name:   &name1,
		Orig:   &name2,
		Width:  &width1,
		Height: &height1,
		Mime:   &name1,
		Ext:    &name1,
	}
	im.Properties = []Property{}
	c.Add(context.Background(), &im)
	p := mongo.NewProcessor[*Image](c, nil, nil, nil)
	nim := Image{
		D: D{
			Id:      id,
			V:       &version1,
			Deleted: &bool1,
		},
		Name:   &name1,
		Orig:   &name2,
		Width:  &width1,
		Height: &height1,
		Mime:   &name1,
		Ext:    &name1,
	}
	nim.Properties = []Property{}
	pr, set, _, err := p.PrepareUpdate(context.Background(), &nim)
	assert.NoError(t, err)
	assert.Equal(t, &nim, pr)
	assert.Equal(t, bson.D{
		bson.E{Key: "_id", Value: id},
		bson.E{Key: "version", Value: version1},
		bson.E{Key: "deleted", Value: true},
	}, set)
}
