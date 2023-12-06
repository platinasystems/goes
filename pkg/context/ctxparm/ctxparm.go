// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package ctxparm

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
)

var (
	Reader  = New(io.Reader(os.Stdin))
	Writer  = New(io.Writer(os.Stdout))
	Strings = New([]string{})
	Map     = New(map[string]any{})
	Flags   = New(flag.CommandLine)
)

func AppendStringsIn(ctx context.Context, strings ...string) context.Context {
	return Strings.With(ctx, append(Strings.In(ctx), strings...))
}

func MapKeysIn(ctx context.Context) string {
	m := Map.In(ctx)
	keys := make([]string, 0, len(m))
	for k := range m {
		if k != "daemon" && !strings.HasPrefix(k, "_") {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		fmt.Fprintln(&sb, " ", k)
	}
	return sb.String()
}

// Return the value of the named flag within the context's FlagSet, or the zero
// value if unavailable or non-convertible. Usage,
//
//	if SearchFlagsIn[bool](ctx, "help") { ... }
func SearchFlagsIn[T comparable](ctx context.Context, name string) (v T) {
	flags := Flags.In(ctx)
	f := flags.Lookup(name)
	if f != nil {
		elem := reflect.ValueOf(f.Value).Elem()
		vtype := reflect.TypeOf(v)
		if elem.CanConvert(vtype) {
			v = elem.Convert(vtype).Interface().(T)
		}
	}
	return
}

func SprintFlagsIn(ctx context.Context) string {
	flags := Flags.In(ctx)
	w := new(strings.Builder)
	fmt.Fprintln(w)
	flags.SetOutput(w)
	flags.PrintDefaults()
	flags.SetOutput(io.Discard)
	return w.String()
}

// A Parameter provides a key, type, default value and methods to get and
// put values in context.
type Parameter[T any] struct {
	Default T
}

func New[T any](defval T) *Parameter[T] {
	return &Parameter[T]{defval}
}

// If present, return parameter value in context, otherwise, return the
// parameter default.
func (p *Parameter[T]) In(ctx context.Context) T {
	if t := ctx.Value(p); t != nil {
		if v, ok := t.(T); ok {
			return v
		}
	}
	return p.Default
}

func (p *Parameter[T]) With(ctx context.Context, v T) context.Context {
	return context.WithValue(ctx, p, v)
}
