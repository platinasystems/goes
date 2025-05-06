// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xflag

import (
	"encoding"
	. "flag"
	"time"
)

type Union interface {
	bool | time.Duration | float64 | int | int64 | string | uint | uint64
}

// [DefineIn] [CommandLine]
func Define[T Union](ptr *T, name, usage string) {
	DefineIn(CommandLine, ptr, name, usage)
}

func DefineIn[T Union](fs *FlagSet, ptr *T, name, usage string) {
	define(fs, ptr, name, usage)
}

func define(fs *FlagSet, ptr any, name, usage string) {
	switch t := ptr.(type) {
	case *bool:
		fs.BoolVar(t, name, *t, usage)
	case *time.Duration:
		fs.DurationVar(t, name, *t, usage)
	case *float64:
		fs.Float64Var(t, name, *t, usage)
	case *int:
		fs.IntVar(t, name, *t, usage)
	case *int64:
		fs.Int64Var(t, name, *t, usage)
	case *string:
		fs.StringVar(t, name, *t, usage)
	case *uint:
		fs.UintVar(t, name, *t, usage)
	case *uint64:
		fs.Uint64Var(t, name, *t, usage)
	}
}

// [DefineTextIn] [CommandLine]
func DefineText[T encoding.TextMarshaler](ptr *T, name, usage string) {
	DefineTextIn(CommandLine, ptr, name, usage)
}

func DefineTextIn[T encoding.TextMarshaler](
	fs *FlagSet, ptr *T, name, usage string,
) {
	fs.TextVar(any(ptr).(encoding.TextUnmarshaler), name, *ptr, usage)
}
