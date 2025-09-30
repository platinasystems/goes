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
	"strconv"
	"time"
)

type Enabler = func() error
type Initializer = func() any

type isBooler interface{ IsBoolFlag() bool }
type setter interface{ Set(string) error }

var (
	ErrNilItem    = errors.New("<nil>")
	ErrNonPointer = errors.New("can't define")
)

/*
A Label associates a flag Name and Usage with an Item that is later defined
in a [flag.FlagSet].
The Item may be a package or function scoped variable pointer,
an [Enabler],
or an [Initializer].

An [Enabler] is called when the named boolean flag is set true.

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
		return ErrNilItem
	}
	// first try to define with the built-in type functions.
	switch t := lbl.Item.(type) {
	case *bool:
		fs.BoolVar(t, lbl.Name, *t, lbl.Usage)
	case *time.Duration:
		fs.DurationVar(t, lbl.Name, *t, lbl.Usage)
	case *float64:
		fs.Float64Var(t, lbl.Name, *t, lbl.Usage)
	case *int:
		fs.IntVar(t, lbl.Name, *t, lbl.Usage)
	case *int64:
		fs.Int64Var(t, lbl.Name, *t, lbl.Usage)
	case *string:
		fs.StringVar(t, lbl.Name, *t, lbl.Usage)
	case *uint:
		fs.UintVar(t, lbl.Name, *t, lbl.Usage)
	case *uint64:
		fs.Uint64Var(t, lbl.Name, *t, lbl.Usage)
	default:
		if _, ok := lbl.Item.(Enabler); !ok {
			if f, ok := lbl.Item.(Initializer); ok {
				v := f()
				if err, oops := v.(error); oops {
					return err
				}
				lbl.Item = v
			}
			k := reflect.ValueOf(lbl.Item).Kind()
			if k != reflect.Ptr {
				return fmt.Errorf("%w %v", ErrNonPointer, k)
			}
		}
		fs.Var(lbl, lbl.Name, lbl.Usage)
	}
	return nil
}

func (lbl *Label) Get() any {
	if lbl.Item == nil {
		return nil
	}
	if getter, ok := lbl.Item.(flag.Getter); ok {
		return getter.Get()
	}
	if _, ok := lbl.Item.(Enabler); ok {
		return false
	}
	if v := reflect.ValueOf(lbl.Item); v.Kind() == reflect.Ptr {
		return v.Elem().Interface()
	}
	return nil
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
	if lbl.Item == nil {
		err = ErrNilItem
	} else if f, ok := lbl.Item.(Enabler); ok {
		if len(s) > 0 {
			if t, err := strconv.ParseBool(s); !t || err != nil {
				return err
			}
		}
		return f()
	} else if setter, ok := lbl.Item.(setter); ok {
		return setter.Set(s)
	} else if t, ok := lbl.Item.(*string); ok {
		*t = s
	} else if t, ok := lbl.Item.(*bool); ok && len(s) == 0 {
		*t = true
	} else if t, ok := lbl.Item.(*time.Duration); ok {
		*t, err = time.ParseDuration(s)
	} else if t, ok := lbl.Item.(*net.HardwareAddr); ok {
		*t, err = net.ParseMAC(s)
	} else if t, ok := lbl.Item.(encoding.TextUnmarshaler); ok {
		err = t.UnmarshalText([]byte(s))
	} else {
		_, err = fmt.Sscan(s, lbl.Item)
	}
	return err
}

func (lbl *Label) String() string {
	if lbl.Item == nil {
		return ""
	}
	if _, ok := lbl.Item.(Enabler); ok {
		return "false"
	}
	if _, ok := lbl.Item.(Initializer); ok {
		return "uninitialized"
	}
	if t, ok := lbl.Item.(fmt.Stringer); ok {
		return t.String()
	}
	switch t := lbl.Item.(type) {
	case *string:
		return *t
	case *uint:
		return strconv.FormatUint(uint64(*t), 10)
	case *uint8:
		return strconv.FormatUint(uint64(*t), 10)
	case *uint16:
		return strconv.FormatUint(uint64(*t), 10)
	case *uint32:
		return strconv.FormatUint(uint64(*t), 10)
	case *uint64:
		return strconv.FormatUint(uint64(*t), 10)
	}
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
