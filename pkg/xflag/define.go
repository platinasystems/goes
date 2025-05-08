// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xflag

import (
	"encoding"
	"errors"
	"flag"
	"fmt"
	"net"
	"strconv"
	"time"
)

type IsBoolFlagger interface{ IsBoolFlag() bool }

// [DefineIn] [flag.CommandLine]
func Define[T any](ptr *T, name, usage string) {
	DefineIn(flag.CommandLine, ptr, name, usage)
}

// [flag.FlagSet.Var] with [Reference] wrap of generic pointer.
func DefineIn[T any](fs *flag.FlagSet, p *T, name, usage string) {
	r := Reference[T]{p}
	fs.Var(r, name, usage)
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

// [Define] wraps its generic pointer arg with Reference to implement
// [flag.Getter] and [IsBoolFlagger].
type Reference[T any] struct{ p *T }

// If not nil, return the referenced value; otherwise returns its zero.
func (r Reference[T]) Get() any {
	if r.p == nil {
		var z T
		return z
	}
	return *r.p
}

// If reference to bool, returns true; otherwise false.
func (r Reference[T]) IsBoolFlag() bool {
	_, ok := any(r.p).(*bool)
	return ok
}

// If reference to string, assign with arg;
// if to bool and arg is empty, assign with true;
// if to [time.Duration], assign with [time.ParseDuration] results;
// if to [net.HardwareAddr], assign with [net.ParseMAC] results;
// if it's an [encoding.TextUnmarshaler], call its UnmarshalText() with arg;
// otherwise, [fmt.Sscan] it from arg.
func (r Reference[T]) Set(s string) error {
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
func (r Reference[T]) String() string {
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
