package inmemory

import (
	"sync"
)

// Idx ...
type Idx struct {
	sync.RWMutex
	maxIdx       int
	itemsByIndex map[int]string
	indexByID    map[string]int
}

// GetIDByIndex ...
func (c *Idx) GetIDByIndex(idx int) (id string, found bool) {
	c.RLock()
	defer c.RUnlock()
	id, found = c.itemsByIndex[idx]
	return
}

// GetIndexByID ...
func (c *Idx) GetIndexByID(id string) (idx int, found bool) {
	c.RLock()
	defer c.RUnlock()
	idx, found = c.indexByID[id]
	return
}

// Add ...
func (c *Idx) add(id string) {
	c.maxIdx++
	c.itemsByIndex[c.maxIdx] = id
	c.indexByID[id] = c.maxIdx
}

// Delete ...
func (c *Idx) deleteByID(id string) {
	if idx, f := c.indexByID[id]; f {
		delete(c.itemsByIndex, idx)
	}
	delete(c.indexByID, id)
}

func (c *Idx) deleteByIdx(idx int) {
	if id, f := c.itemsByIndex[idx]; f {
		delete(c.indexByID, id)
	}
	delete(c.itemsByIndex, idx)
}

// NewIdx ...
func NewIdx(
	maxIdx int,
	itemsByIndex map[int]string,
	indexByID map[string]int,
) *Idx {
	return &Idx{
		maxIdx:       maxIdx,
		itemsByIndex: itemsByIndex,
		indexByID:    indexByID,
	}
}
