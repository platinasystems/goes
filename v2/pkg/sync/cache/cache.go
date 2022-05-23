// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cache

import (
	"fmt"
	"sync"
)

type Cache[T any] struct {
	m      sync.RWMutex
	load   func(*T) error
	loaded bool
	err    error
	v      T
}

func New[T any](load func(*T) error) *Cache[T] {
	return &Cache[T]{load: load}
}

// Preemptive load.
func (c *Cache[T]) Preload(f func(*T)) {
	c.m.Lock()
	defer c.m.Unlock()
	f(&c.v)
	c.loaded = true
	c.err = nil
}

func (c *Cache[T]) Reload() {
	c.m.Lock()
	defer c.m.Unlock()
	c.loaded = true
	c.err = c.load(&c.v)
}

// If not yet loaded, do so and return if error; otherwise, continue with call
// to f() with read+write locked reference and return its results.
func (c *Cache[T]) Ref(f func(*T) error) error {
	c.m.Lock()
	defer c.m.Unlock()
	if !c.loaded {
		c.loaded = true
		if c.load != nil {
			c.err = c.load(&c.v)
		}
	}
	if c.err != nil {
		return c.err
	}
	return f(&c.v)
}

func (c *Cache[T]) String() string {
	v, err := c.Value()
	if err != nil {
		panic(err)
	}
	return fmt.Sprint(v)
}

// If not yet loaded, do so before returning value.  If load failed, this will
// continue to return its error until overwritten by Preload or Reload.
func (c *Cache[T]) Value() (T, error) {
	c.m.RLock()
	if c.loaded {
		defer c.m.RUnlock()
	} else {
		c.m.RUnlock()
		c.m.Lock()
		defer c.m.Unlock()
		if !c.loaded {
			c.loaded = true
			if c.load != nil {
				c.err = c.load(&c.v)
			}
		}
	}
	return c.v, c.err
}
