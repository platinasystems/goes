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
	"reflect"
	"time"
)

type Enabler = func(string) error
type Initializer = func() any
type Handler func(string) error
type handler = func(string) error

type isBooler interface{ IsBoolFlag() bool }
type setter interface{ Set(string) error }

var (
	ErrNilItem     = errors.New("<nil>")
	ErrUndefinable = errors.New("undefinable")
)

/*
A Label associates a flag Name and Usage with an Item that is later defined
in a [flag.FlagSet].
The Item may be a variable pointer,
a boolean [Enabler],
a generic [Handler]
or an [Initializer].

An [Enabler] or [Handler] is called each time the named boolean or generic flag
is parsed.

The non-error, [Initializer] pointer result
overrides the Item before the flag is defined.

See [Labels].
*/
type Label struct {
	Name,
	Usage string
	Item any
}

// [Label.DefineIn] [flag.CommandLine]
func (lbl *Label) Define() error { return lbl.DefineIn(flag.CommandLine) }

func (lbl *Label) DefineIn(fs *flag.FlagSet) error {
	if lbl.Item == nil {
		return fmt.Errorf("%s: %w", lbl.Name, ErrNilItem)
	}
	var err error
	switch t := lbl.Item.(type) {
	case *bool:
		fs.BoolVar(t, lbl.Name, *t, lbl.Usage)
	case func() *bool:
		lbl.Item = t()
		err = lbl.DefineIn(fs)
	case *time.Duration:
		fs.DurationVar(t, lbl.Name, *t, lbl.Usage)
	case *float64:
		fs.Float64Var(t, lbl.Name, *t, lbl.Usage)
	case *int:
		fs.IntVar(t, lbl.Name, *t, lbl.Usage)
	case func() *int:
		lbl.Item = t()
		err = lbl.DefineIn(fs)
	case *int64:
		fs.Int64Var(t, lbl.Name, *t, lbl.Usage)
	case func() *int64:
		lbl.Item = t()
		err = lbl.DefineIn(fs)
	case *string:
		fs.StringVar(t, lbl.Name, *t, lbl.Usage)
	case func() *string:
		lbl.Item = t()
		err = lbl.DefineIn(fs)
	case *uint:
		fs.UintVar(t, lbl.Name, *t, lbl.Usage)
	case func() *uint:
		lbl.Item = t()
		err = lbl.DefineIn(fs)
	case *uint64:
		fs.Uint64Var(t, lbl.Name, *t, lbl.Usage)
	case func() *uint64:
		lbl.Item = t()
		err = lbl.DefineIn(fs)
	case Enabler:
		fs.BoolFunc(lbl.Name, lbl.Usage, t)
	case Initializer:
		lbl.Item = t()
		if ie, oops := lbl.Item.(error); oops {
			err = ie
		} else {
			err = lbl.DefineIn(fs)
		}
	case Handler:
		fs.Func(lbl.Name, lbl.Usage, (handler)(t))
	case encoding.TextUnmarshaler:
		v := reflect.ValueOf(t)
		if k := v.Kind(); k != reflect.Ptr {
			err = fmt.Errorf("(%v)%s: %w",
				k, lbl.Name, ErrUndefinable)
		} else {
			elem := v.Elem().Interface()
			if tm, ok := elem.(encoding.TextMarshaler); ok {
				fs.TextVar(t, lbl.Name, tm, lbl.Usage)
			} else {
				err = fmt.Errorf("*(%v)%s: %w",
					k, lbl.Name, ErrUndefinable)
			}
		}
	default:
		v := reflect.ValueOf(t)
		if k := v.Kind(); k != reflect.Ptr {
			err = fmt.Errorf("(%v)%s %wv",
				k, lbl.Name, ErrUndefinable)
		} else if val, ok := t.(flag.Value); ok {
			fs.Var(val, lbl.Name, lbl.Usage)
		} else {
			fs.Var(lbl, lbl.Name, lbl.Usage)
		}
	}
	return err
}

func (lbl *Label) Get() any {
	if getter, ok := lbl.Item.(flag.Getter); ok {
		return getter.Get()
	}
	return reflect.ValueOf(lbl.Item).Elem().Interface()
}

func (lbl *Label) IsBoolFlag() bool {
	if _, ok := lbl.Item.(*bool); ok {
		return ok
	}
	if _, ok := lbl.Item.(Enabler); ok {
		return ok
	}
	if isBooler, ok := lbl.Item.(isBooler); ok {
		return isBooler.IsBoolFlag()
	}
	return false
}

func (lbl *Label) Set(s string) (err error) {
	if t, ok := lbl.Item.(*net.HardwareAddr); ok {
		*t, err = net.ParseMAC(s)
	} else {
		_, err = fmt.Sscan(s, lbl.Item)
	}
	return err
}

func (lbl *Label) String() string {
	return fmt.Sprint(lbl.Item)
}

type Labels []Label

// [Labels.DefineIn] [flag.CommandLine]
func (labels Labels) Define() error { return labels.DefineIn(flag.CommandLine) }

// Iteratively [Label.DefineIn] until done or error.
func (labels Labels) DefineIn(fs *flag.FlagSet) (err error) {
	for _, lbl := range labels {
		if err = lbl.DefineIn(fs); err != nil {
			break
		}
	}
	return
}
