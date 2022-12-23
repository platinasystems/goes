// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cache

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

var ErrReadOnly = errors.New("read only")

type Mutexer[T any] interface {
	Mutex(func(*T) error) error
}

type Cache[T any] struct {
	m   sync.RWMutex
	l   func(*T) error
	ok  bool
	ro  bool
	err error
	v   T
}

func New[T any](load func(*T) error) *Cache[T] {
	return &Cache[T]{l: load}
}

func NewReadOnly[T any](load func(*T) error) *Cache[T] {
	return &Cache[T]{
		l:  load,
		ro: true,
	}
}

func (c *Cache[T]) Invalidate() {
	c.m.Lock()
	defer c.m.Unlock()
	c.ok = false
}

func (c *Cache[T]) MarshalJSON() (text []byte, err error) {
	v, err := c.ValErr()
	if err == nil {
		text, err = json.Marshal(v)
	}
	return
}

func (c *Cache[T]) MarshalText() (text []byte, err error) {
	v, err := c.ValErr()
	if err == nil {
		if method, ok := any(v).(encoding.TextMarshaler); ok {
			text, err = method.MarshalText()
		} else {
			text = []byte(fmt.Sprint(v))
		}
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

// Preemptive load.
func (c *Cache[T]) Preload(f func(*T)) {
	c.m.Lock()
	defer c.m.Unlock()
	f(&c.v)
	c.ok = true
	c.err = nil
}

func (c *Cache[T]) UnmarshalJSON(text []byte) (err error) {
	c.m.Lock()
	defer c.m.Unlock()
	if c.ro {
		err = ErrReadOnly
	} else if err = json.Unmarshal(text, &c.v); err == nil {
		c.ok = true
	}
	return
}

func (c *Cache[T]) UnmarshalText(text []byte) (err error) {
	c.m.Lock()
	defer c.m.Unlock()
	if c.ro {
		err = ErrReadOnly
	} else if m, ok := any(&c.v).(encoding.TextUnmarshaler); ok {
		if err = m.UnmarshalText(text); err == nil {
			c.ok = true
		}
	} else if _, err = fmt.Sscan(string(text), &c.v); err == nil {
		c.ok = true
	}
	return
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
