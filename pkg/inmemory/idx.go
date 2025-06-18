package inmemory

import (
	"sync"
	"sync/atomic"
)

// Idx ...
type Idx struct {
	sync.RWMutex
	maxIdx       atomic.Int64
	itemsByIndex map[int64]string
	indexByID    map[string]int64
}

// GetIDByIndex ...
func (c *Idx) GetIDByIndex(idx int) (id string, found bool) {
	c.RLock()
	defer c.RUnlock()
	id, found = c.itemsByIndex[int64(idx)]
	return
}

// GetIndexByID ...
func (c *Idx) GetIndexByID(id string) (idx int, found bool) {
	c.RLock()
	defer c.RUnlock()
	var _idx int64
	_idx, found = c.indexByID[id]
	idx = int(_idx)
	return
}

// Add ...
func (c *Idx) add(id string) {
	max := c.maxIdx.Add(1)
	c.itemsByIndex[max] = id
	c.indexByID[id] = max
}

// Delete ...
func (c *Idx) deleteByID(id string) {
	if idx, f := c.indexByID[id]; f {
		delete(c.itemsByIndex, idx)
	}
	delete(c.indexByID, id)
}

func (c *Idx) deleteByIdx(idx int) {
	if id, f := c.itemsByIndex[int64(idx)]; f {
		delete(c.indexByID, id)
	}
	delete(c.itemsByIndex, int64(idx))
}

// NewIdx ...
func NewIdx(
	maxIdx atomic.Int64,
	itemsByIndex map[int64]string,
	indexByID map[string]int64,
) *Idx {
	return &Idx{
		maxIdx:       maxIdx,
		itemsByIndex: itemsByIndex,
		indexByID:    indexByID,
	}
}
