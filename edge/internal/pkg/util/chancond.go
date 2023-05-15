package util

import (
	"sync"
)

type ChanCond struct {
	m  sync.Mutex
	ch chan struct{}
}

func (c *ChanCond) InitBuffered(size uint) {
	c.m.Lock()
	defer c.m.Unlock()
	if c.ch == nil {
		c.ch = make(chan struct{}, size)
	}
}

func (c *ChanCond) Wait() <-chan struct{} {
	c.m.Lock()
	defer c.m.Unlock()

	if c.ch == nil {
		c.ch = make(chan struct{})
	}

	return c.ch
}

func (c *ChanCond) Signal() {
	c.m.Lock()
	defer c.m.Unlock()
	select {
	case c.ch <- struct{}{}:
	default:
	}
}

func (c *ChanCond) Broadcast() {
	c.m.Lock()
	defer c.m.Unlock()
	if c.ch == nil {
		return
	}
	close(c.ch)
	c.ch = nil
}
