// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xflag

import (
	"encoding"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Definer interface{ Define() }
type DefineInner interface{ DefineIn(*flag.FlagSet) }

// Instantiate package scoped variables that are defined flags w/in mains.
func New[T any](name, usage string, init func() T) *Generic[T] {
	return &Generic[T]{
		name:  name,
		usage: usage,
		init:  init,
	}
}

type Generic[T any] struct {
	name,
	usage string
	init func() T
	once struct {
		init, define sync.Once
	}
	val T
}

func (g *Generic[T]) Define() {
	g.DefineIn(flag.CommandLine)
}

func (g *Generic[T]) DefineIn(fs *flag.FlagSet) {
	g.once.init.Do(g.valInit)
	g.once.define.Do(func() {
		if !define(fs, &g.val, g.name, g.usage) {
			r := reference[T]{&g.val}
			fs.Var(r, g.name, g.usage)
		}
	})
}

func (g *Generic[T]) Override(v T) {
	g.once.init.Do(func() {})
	g.once.define.Do(func() {})
	g.val = v
}

func (g *Generic[T]) String() string {
	return fmt.Sprint(g.Value())
}

func (g *Generic[T]) Value() T {
	g.once.init.Do(g.valInit)
	return g.val
}

func (g *Generic[T]) valInit() {
	if g.init != nil {
		g.val = g.init()
	}
}

// Instantiate package scoped flag with [Dir.File] method.
type Dir struct{ Generic[string] }

func NewDir(name, usage string, init func() string) *Dir {
	return &Dir{Generic[string]{
		name:  name,
		usage: usage,
		init:  init,
	}}
}

// If given string doesn't equal "-" or have a [filepath.Separator],
// then [filepath.Join] it to [Dir.Value];
// otherwise, return unchanged.
func (gd *Dir) File(s string) string {
	if s != "-" && strings.IndexRune(s, filepath.Separator) < 0 {
		if _, err := os.Stat(s); errors.Is(err, fs.ErrNotExist) {
			if val := gd.Value(); len(val) > 0 {
				s = filepath.Join(val, s)
			}
		}
	}
	return s
}

// If the [flag] package doesn't provide a [flag.FlagSet] method for the [Generic] type, wrap it with this reference to implement [flag.Getter] and IsBoolFlag.
type reference[T any] struct{ p *T }

// If not nil, return the referenced value; otherwise returns its zero.
func (r reference[T]) Get() any {
	if r.p == nil {
		var z T
		return z
	}
	return *r.p
}

// If reference to bool, returns true; otherwise false.
func (r reference[T]) IsBoolFlag() bool {
	_, ok := any(r.p).(*bool)
	return ok
}

// If reference to string, assign with arg;
// if to bool and arg is empty, assign with true;
// if to [time.Duration], assign with [time.ParseDuration] results;
// if to [net.HardwareAddr], assign with [net.ParseMAC] results;
// if it's an [encoding.TextUnmarshaler], call its UnmarshalText() with arg;
// otherwise, [fmt.Sscan] it from arg.
func (r reference[T]) Set(s string) error {
	return refset(r.p, s)
}

func refset(p any, s string) (err error) {
	if p == nil {
		err = errors.New("<nil>")
	} else if t, ok := p.(*string); ok {
		*t = s
	} else if t, ok := p.(*bool); ok && len(s) == 0 {
		*t = true
	} else if t, ok := p.(*time.Duration); ok {
		*t, err = time.ParseDuration(s)
	} else if t, ok := p.(*net.HardwareAddr); ok {
		*t, err = net.ParseMAC(s)
	} else if t, ok := p.(encoding.TextUnmarshaler); ok {
		err = t.UnmarshalText([]byte(s))
	} else {
		_, err = fmt.Sscan(s, p)
	}
	return
}

// If the referenced value is a [fmt.Stringer], return its String() results;
// if it's a string, return that value;
// if it's a uint of any size, return its base 10 string format;
// otherwise, return its [fmt.Sprint] results.
func (r reference[T]) String() string {
	if r.p == nil {
		return ""
	}
	return refstring(*r.p)
}

func refstring(v any) string {
	if t, ok := v.(fmt.Stringer); ok {
		return t.String()
	}
	switch t := v.(type) {
	case string:
		return t
	case uint:
		return strconv.FormatUint(uint64(t), 10)
	case uint8:
		return strconv.FormatUint(uint64(t), 10)
	case uint16:
		return strconv.FormatUint(uint64(t), 10)
	case uint32:
		return strconv.FormatUint(uint64(t), 10)
	case uint64:
		return strconv.FormatUint(uint64(t), 10)
	default:
		return fmt.Sprint(t)
	}
	return ""
}

// If available, first try to define with the [flag] builit-in type functions.
func define(fs *flag.FlagSet, v any, name, usage string) bool {
	switch p := v.(type) {
	case *bool:
		fs.BoolVar(p, name, *p, usage)
	case *time.Duration:
		fs.DurationVar(p, name, *p, usage)
	case *float64:
		fs.Float64Var(p, name, *p, usage)
	case *int:
		fs.IntVar(p, name, *p, usage)
	case *int64:
		fs.Int64Var(p, name, *p, usage)
	case *string:
		fs.StringVar(p, name, *p, usage)
	case *uint:
		fs.UintVar(p, name, *p, usage)
	case *uint64:
		fs.Uint64Var(p, name, *p, usage)
	default:
		return false
	}
	return true
}

func Define[T any](ptr *T, name, usage string) {
	DefineIn(flag.CommandLine, ptr, name, usage)
}

func DefineIn[T any](fs *flag.FlagSet, p *T, name, usage string) {
	if !define(fs, p, name, usage) {
		r := reference[T]{p}
		fs.Var(r, name, usage)
	}
}

// [DefineValIn] [flag.CommandLine]
func DefineVal[T any](name string, val T, usage string) *T {
	return DefineValIn(flag.CommandLine, name, val, usage)
}

// [DefineIn] with clone.
func DefineValIn[T any](fs *flag.FlagSet, name string, v T, usage string) *T {
	p := new(T)
	*p = v
	DefineIn(fs, p, name, usage)
	return p
}

// [DefineVarIn] [flag.CommandLine]
func DefineVar[T any](p *T, name string, v T, usage string) {
	DefineVarIn(flag.CommandLine, p, name, v, usage)
}

// Load value then [DefineIn].
func DefineVarIn[T any](
	fs *flag.FlagSet, p *T, name string, v T, usage string,
) {
	*p = v
	DefineIn(fs, p, name, usage)
}
