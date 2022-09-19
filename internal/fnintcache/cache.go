package fnintcache

import "sync"

const idxBorder = 1 << 16 // 65536

type Cache[T any] struct {
	mx        *sync.Mutex
	smallInts [idxBorder]*T // 65536 items
	bigInts   map[int]*T    // for indexes above idxBorder
}

func New[T any]() *Cache[T] {
	return &Cache[T]{
		bigInts: make(map[int]*T, 1024),
		mx:      &sync.Mutex{},
	}
}

func (c *Cache[T]) Get(i int) (*T, bool) {
	if i < 0 {
		return nil, false
	}

	if i < idxBorder {
		if c.smallInts[i] == nil {
			return nil, false
		}
		return c.smallInts[i], true
	}

	c.mx.Lock()
	obj, found := c.bigInts[i]
	c.mx.Unlock()
	return obj, found
}

func (c *Cache[T]) Set(i int, res T) {
	if i < idxBorder {
		c.smallInts[i] = &res
		return
	}

	c.mx.Lock()
	c.bigInts[i] = &res
	c.mx.Unlock()
}
