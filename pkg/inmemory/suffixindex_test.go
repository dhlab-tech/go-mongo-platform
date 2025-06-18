package inmemory

import (
	"sync"
	"testing"

	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

func TestS_Put_1(t *testing.T) {
	a := &S{
		RWMutex:   sync.RWMutex{},
		intersect: NewIntersect(),
		data:      btree.New(100),
		pool:      new(Pool),
	}
	a.Put("Булочка с вишней", 1)
	a.Put("БАК ФАСОВКА Булгур", 2)
	res := a.Search("Було")
	assert.Equal(t, []int{1}, res)
}

func TestS_Put_2(t *testing.T) {
	a := &S{
		RWMutex:   sync.RWMutex{},
		intersect: NewIntersect(),
		data:      btree.New(100),
		pool:      new(Pool),
	}
	a.Put("Булочка с вишней", 1)
	a.Put("БАК ФАСОВКА Булгур", 2)
	a.Put("Булочка с корицей", 3)
	res := a.Search("Було")
	assert.Equal(t, []int{1, 3}, res)
}

func TestS_Put_3(t *testing.T) {
	a := &S{
		RWMutex:   sync.RWMutex{},
		intersect: NewIntersect(),
		data:      btree.New(100),
		pool:      new(Pool),
	}
	a.Put("Булочка с вишней", 1)
	a.Put("БАК ФАСОВКА Булгур", 2)
	a.Put("Булочка с корицей", 3)
	res := a.Search("фас")
	assert.Equal(t, []int{2}, res)
}
