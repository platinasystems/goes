// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package flag

import (
	"flag"
	"io"
	"reflect"
)

type Flag = flag.Flag
type FlagSet = flag.FlagSet

var CommandLine = flag.CommandLine

// Make a new flag.FlagSet that doesn't output during Parse and instead sets an
// associate `help` flag.
func New(name string) *FlagSet {
	fs := flag.NewFlagSet(name, 0)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}
	help := fs.Bool("help", false, "Print options.")
	fs.BoolVar(help, "h", *help, fs.Lookup("help").Usage)
	return fs
}

// Return the first non-zero flag value within the referenced `flagsets` or the
// generic's zero value if unavailable or non-convertible. Usage,
//
//	if Search[bool]("help", fs) { ... }
//
// If there are less than two `flagsets`, the CommandLine is also searched but
// this may be avoided an extra nil, e.g.
//
//	if Search[bool]("help", fs, nil) { ... }
func Search[T comparable](name string, flagsets ...*FlagSet) T {
	var z T
	ztype := reflect.TypeOf(z)
	if len(flagsets) < 2 {
		flagsets = append(flagsets, CommandLine)
	}
	for _, fs := range flagsets {
		if fs == nil {
			continue
		}
		f := fs.Lookup(name)
		if f == nil {
			continue
		}
		elem := reflect.ValueOf(f.Value).Elem()
		if !elem.CanConvert(ztype) {
			continue
		}
		v := elem.Convert(ztype).Interface().(T)
		if v != z {
			return v
		}
	}
	return z
}
