// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cache

import (
	"context"
	"fmt"
	"sync"
)

type Cache[T any] struct {
	m   sync.RWMutex
	l   func(*T) error
	lc  func(context.Context, *T) error
	ok  bool
	err error
	v   T
}

func New[T any](load func(*T) error) *Cache[T] {
	return &Cache[T]{l: load}
}

func NewContext[T any](load func(context.Context, *T) error) *Cache[T] {
	return &Cache[T]{lc: load}
}

func (c *Cache[T]) Invalidate() {
	c.m.Lock()
	defer c.m.Unlock()
	c.ok = false
}

func (c *Cache[T]) MarshalText() (text []byte, err error) {
	v, err := c.ValErr()
	if err == nil {
		text = []byte(fmt.Sprint(v))
	}
	return
}

func (c *Cache[T]) MarshalTextContext(ctx context.Context) (
	text []byte,
	err error,
) {
	v, err := c.ValErrContext(ctx)
	if err == nil {
		text = []byte(fmt.Sprint(v))
	}
	return
}

// If not yet loaded, do so and return if error; otherwise, continue with call
// to f() with locked reference and return f's results.
func (c *Cache[T]) Mutex(f func(*T) error) error {
	c.m.Lock()
	defer c.m.Unlock()
	if !c.ok {
		c.ok = true
		if c.l != nil {
			c.err = c.l(&c.v)
		}
	}
	if c.err != nil {
		return c.err
	}
	return f(&c.v)
}

func (c *Cache[T]) MutexContext(
	ctx context.Context,
	f func(context.Context, *T) error) error {
	c.m.Lock()
	defer c.m.Unlock()
	if !c.ok {
		c.ok = true
		if c.lc != nil {
			c.err = c.lc(ctx, &c.v)
		}
	}
	if c.err != nil {
		return c.err
	}
	return f(ctx, &c.v)
}

// Preemptive load.
func (c *Cache[T]) Preload(f func(*T)) {
	c.m.Lock()
	defer c.m.Unlock()
	f(&c.v)
	c.ok = true
	c.err = nil
}

// If not yet loaded, do so before returning value.  If load failed,
// return its zero value and Err continues to return the failure until
// overwritten by Preload or Reload.
func (c *Cache[T]) ValErr() (T, error) {
	c.m.RLock()
	if c.ok {
		defer c.m.RUnlock()
	} else {
		c.m.RUnlock()
		c.m.Lock()
		defer c.m.Unlock()
		if !c.ok {
			c.ok = true
			if c.l != nil {
				c.err = c.l(&c.v)
			}
		}
	}
	return c.v, c.err
}

func (c *Cache[T]) Value() T {
	t, err := c.ValErr()
	if err != nil {
		panic(err)
	}
	return t
}

func (c *Cache[T]) ValErrContext(ctx context.Context) (T, error) {
	c.m.RLock()
	if c.ok {
		defer c.m.RUnlock()
	} else {
		c.m.RUnlock()
		c.m.Lock()
		defer c.m.Unlock()
		if !c.ok {
			c.ok = true
			if c.lc != nil {
				c.err = c.lc(ctx, &c.v)
			}
		}
	}
	return c.v, c.err
}

func (c *Cache[T]) ValueContext(ctx context.Context) T {
	t, err := c.ValErrContext(ctx)
	if err != nil {
		panic(err)
	}
	return t
}
