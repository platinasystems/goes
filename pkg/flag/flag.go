// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package flag

import (
	"flag"
	"io"
	"reflect"
)

const helpmsg = "Print options."

type Flag = flag.Flag
type FlagSet = flag.FlagSet

var CommandLine = flag.CommandLine

// Make a new flag.FlagSet that doesn't output during Parse and instead sets an
// associate `help` flag.
func New(name string) *FlagSet {
	fs := flag.NewFlagSet(name, 0)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}
	help := fs.Bool("help", false, helpmsg)
	fs.BoolVar(help, "h", *help, helpmsg)
	return fs
}

// Return the elemental value of a flag or the generic's zero value if the flag
// isn't w/in FlagSet or it's value can't be converted. Usage,
//
//	if Eval[bool](CommandLine, "verbose") { ... }
func Eval[T any](fs *FlagSet, name string) (t T) {
	if fs != nil {
		if f := fs.Lookup(name); f != nil {
			elem := reflect.ValueOf(f.Value).Elem()
			ttype := reflect.TypeOf(t)
			if elem.CanConvert(ttype) {
				t = elem.Convert(ttype).Interface().(T)
			}
		}
	}
	return
}
