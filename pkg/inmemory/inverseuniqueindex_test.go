package inmemory_test

import (
	"context"
	"testing"

	"github.com/dhlab-tech/go-mongo-platform/pkg/inmemory"
	"github.com/stretchr/testify/assert"
)

var (
	group1 = "test1"
	group2 = "test2"
	group3 = "test3"
	group4 = "test4"
)

type Group struct {
	D
	P
	S
	Title   *string  `json:"title" bson:"title"`
	Members []string `json:"members" bson:"members"` // user_ids
	CtxID   *string  `json:"ctxId" bson:"ctxId"`
	Carts   []string `json:"carts" bson:"carts" indexes:"inverse_unique:cartId:from"`
}

func TestInverseUniqueIndex_Slice(t *testing.T) {
	c := inmemory.NewCache[*Group](make(map[string]*Group))
	idx := inmemory.NewInverseUniqIndex(map[string]string{}, c, []string{"Carts"}, nil)
	grp := Group{Carts: []string{"test1", "test2", "test3"}}
	idx.Add(context.Background(), &grp)
	c.Add(context.Background(), &grp)
	id, found := idx.Get(context.Background(), group1)
	assert.Equal(t, true, found)
	assert.Equal(t, grp.ID(), id)
	id, found = idx.Get(context.Background(), group2)
	assert.Equal(t, true, found)
	assert.Equal(t, grp.ID(), id)
	id, found = idx.Get(context.Background(), group3)
	assert.Equal(t, true, found)
	assert.Equal(t, grp.ID(), id)
	_, found = idx.Get(context.Background(), group4)
	assert.Equal(t, false, found)
}
