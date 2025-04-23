// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xflag

import (
	"encoding"
	"errors"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"time"
	"unicode"
)

var ErrIsUndefinable = errors.New("is undefinable")
var ErrIsNotTextUnmarshaler = errors.New("is not TextUnmarshaler")

func Get[T any](name string) T {
	return GetFrom[T](flag.CommandLine, name)
}

func GetFrom[T any](flags *flag.FlagSet, name string) (v T) {
	if f := flags.Lookup(name); f != nil {
		if getter, ok := f.Value.(flag.Getter); ok {
			v, _ = getter.Get().(T)
		}
	}
	return
}

// A [unicode.Space] separated flag key and usage, e.g.
//
//	const VerboseFlag KeyUsage[bool] = "v Verbose output."
type KeyUsage[T any] string

// [KeyUsage.DefineIn] [flag.CommandLine]
func (ku KeyUsage[T]) Define(val T, aliases ...string) *T {
	return ku.DefineIn(flag.CommandLine, val, aliases...)
}

func (ku KeyUsage[T]) DefineIn(flags *flag.FlagSet, val T, aliases ...string) *T {
	var key, usage string
	for i, r := range string(ku) {
		if len(key) == 0 {
			if unicode.IsSpace(r) {
				key = string(ku)[:i]
			}
		} else if !unicode.IsSpace(r) {
			usage = string(ku)[i:]
			break
		}
	}
	return define(flags, key, val, usage, aliases...).(*T)
}

// [KeyUsage.ValueIn] [flag.CommandLine]
func (ku KeyUsage[T]) Value() T {
	return ku.ValueIn(flag.CommandLine)
}

func (ku KeyUsage[T]) ValueIn(flags *flag.FlagSet) T {
	var key string
	for i, r := range string(ku) {
		if unicode.IsSpace(r) {
			key = string(ku)[:i]
			break
		}
	}
	return GetFrom[T](flags, key)
}

func LastName(flags *flag.FlagSet) string {
	name := flags.Name()
	if i := strings.LastIndex(name, " "); i > 0 {
		name = name[i+1:]
	}
	return name
}

// [flag.FlagSet.Init] wht new name but current [flag.ErrorHandling].
func Rename(flags *flag.FlagSet, name string) {
	flags.Init(name, flags.ErrorHandling())
}

// Redirect [flag.FlagSet.PrintDefaults] to string.
func SprintDefaults(flags *flag.FlagSet) string {
	save := flags.Output()
	defer flags.SetOutput(save)
	var sb strings.Builder
	flags.SetOutput(&sb)
	flags.PrintDefaults()
	return sb.String()
}

// [TemplateUsageIn] [flag.CommandLine]
func TemplateUsage(tmpl string) {
	TemplateUsageIn(flag.CommandLine, tmpl)
}

// Assign [flag.FlagSet.Usage] to a closure that creates a new [text/template]
// with [UsageFuncs]; parses “tmpl”; then [text/template.Template.Execute]'s
// to [flag.FlagSet.Output] with [UsageData].
func TemplateUsageIn(flags *flag.FlagSet, tmpl string) {
	flags.Usage = func() {
		w := flags.Output()
		tmpl = strings.TrimLeft(tmpl, " \t\n")
		tt, err := template.New("usage").Funcs(UsageFuncs).Parse(tmpl)
		if err == nil {
			err = tt.Execute(w, UsageData(flags))
		}
		if err != nil {
			name := strings.Replace(flags.Name(), " ", ":", -1)
			fmt.Fprint(w, name, ":usage:", err, "\n")
		}
	}
}

// These results are passed to the usage template execution.
var UsageData = func(flags *flag.FlagSet) any {
	return flags
}

// The usage template will include these functions.
var UsageFuncs = template.FuncMap{
	"flags": SprintDefaults,
}

func define(flags *flag.FlagSet, name string, val any, usage string,
	aliases ...string,
) any {
	switch t := val.(type) {
	case bool:
		p := flags.Bool(name, t, usage)
		if len(aliases) > 0 {
			aka := fmt.Sprint("aka -", name)
			for _, alias := range aliases {
				flags.BoolVar(p, alias, *p, aka)
			}
		}
		return p
	case time.Duration:
		p := flags.Duration(name, t, usage)
		if len(aliases) > 0 {
			aka := fmt.Sprint("aka -", name)
			for _, alias := range aliases {
				flags.DurationVar(p, alias, *p, aka)
			}
		}
		return p
	case float64:
		p := flags.Float64(name, t, usage)
		if len(aliases) > 0 {
			aka := fmt.Sprint("aka -", name)
			for _, alias := range aliases {
				flags.Float64Var(p, alias, *p, aka)
			}
		}
		return p
	case int:
		p := flags.Int(name, t, usage)
		if len(aliases) > 0 {
			aka := fmt.Sprint("aka -", name)
			for _, alias := range aliases {
				flags.IntVar(p, alias, *p, aka)
			}
		}
		return p
	case int64:
		p := flags.Int64(name, t, usage)
		if len(aliases) > 0 {
			aka := fmt.Sprint("aka -", name)
			for _, alias := range aliases {
				flags.Int64Var(p, alias, *p, aka)
			}
		}
		return p
	case string:
		p := flags.String(name, t, usage)
		if len(aliases) > 0 {
			aka := fmt.Sprint("aka -", name)
			for _, alias := range aliases {
				flags.StringVar(p, alias, *p, aka)
			}
		}
		return p
	case uint:
		p := flags.Uint(name, t, usage)
		if len(aliases) > 0 {
			aka := fmt.Sprint("aka -", name)
			for _, alias := range aliases {
				flags.UintVar(p, alias, *p, aka)
			}
		}
		return p
	case uint64:
		p := flags.Uint64(name, t, usage)
		if len(aliases) > 0 {
			aka := fmt.Sprint("aka -", name)
			for _, alias := range aliases {
				flags.Uint64Var(p, alias, *p, aka)
			}
		}
		return p
	case encoding.TextMarshaler:
		tum, ok := t.(encoding.TextUnmarshaler)
		if !ok {
			rval := reflect.ValueOf(t)
			if rval.Kind() == reflect.Ptr {
				rval = reflect.Indirect(rval)
			}
			p := reflect.New(rval.Type()).Interface()
			tum, ok = p.(encoding.TextUnmarshaler)
			if !ok {
				panic(fmt.Errorf("%T %w", t,
					ErrIsNotTextUnmarshaler))
			}
		}
		flags.TextVar(tum, name, t, usage)
		if len(aliases) > 0 {
			aka := fmt.Sprint("aka -", name)
			for _, alias := range aliases {
				flags.TextVar(tum, alias, t, aka)
			}
		}
		return tum
	default:
		panic(fmt.Errorf("%T %w", t, ErrIsUndefinable))
	}
	return nil
}
