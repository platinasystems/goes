// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xflag

import (
	"encoding"
	"errors"
	"flag"
	"fmt"
	"strings"
	"text/template"
	"time"
	"unicode"
)

var ErrIsUndefinable = errors.New("is undefinable")
var ErrIsNotTextUnmarshaler = errors.New("is not TextUnmarshaler")

// These results are passed to the usage template execution.
var UsageData = func(flags *flag.FlagSet) any {
	return flags
}

// Each usage template includes these functions.
var UsageFuncs = template.FuncMap{
	"flags": SprintDefaults,
}

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

type Definable interface {
	bool | float64 | int | int64 | string | uint | uint64 |
		time.Duration
}

// A flag Description is a generic type wrapping string consisting of a
// [unicode.Space] separated name and usage, e.g.
//
//	const Verbose Description[bool] = "verbose Log everything."
//
// Or with the alias:
//
//	const Verbose Xbool = "verbose Log everything."
//
// Define flag with initial value and any aliases before [flag.Parse].
//
//	Verbose.Define(false)
//
// Then assess flag after [flag.Parse].
//
//	if Verbose.Value() { ... }
type Description[T Definable] string
type Xbool = Description[bool]
type Xduration = Description[time.Duration]
type Xfloat64 = Description[float64]
type Xint = Description[int]
type Xint64 = Description[int64]
type Xstring = Description[string]
type Xuint = Description[uint]
type Xuint64 = Description[uint64]

// [Description.DefineIn] [flag.CommandLine]
func (d Description[T]) Define(val T, aliases ...string) *T {
	return d.DefineIn(flag.CommandLine, val, aliases...)
}

func (d Description[T]) DefineIn(
	flags *flag.FlagSet, val T, aliases ...string,
) *T {
	var name, usage string
	s := d.String()
	for i, r := range s {
		if len(name) == 0 {
			if unicode.IsSpace(r) {
				name = s[:i]
			}
		} else if !unicode.IsSpace(r) {
			usage = s[i:]
			break
		}
	}
	return define(flags, name, val, usage, aliases...).(*T)
}

func (d Description[T]) String() string {
	return string(d)
}

// [Description.ValueIn] [flag.CommandLine]
func (d Description[T]) Value() T {
	return d.ValueIn(flag.CommandLine)
}

func (d Description[T]) ValueIn(flags *flag.FlagSet) T {
	var name string
	s := d.String()
	for i, r := range s {
		if unicode.IsSpace(r) {
			name = s[:i]
			break
		}
	}
	return GetFrom[T](flags, name)
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
	default:
		panic(fmt.Errorf("%T %w", t, ErrIsUndefinable))
	}
	return nil
}

// A flag TextVarDescription is an
// ([encoding.TextUnmarler], [encoding.TextMarler])
// type wrapping string consisting of a
// [unicode.Space] separated name and usage, e.g.
//
//	type AddrPortFlag = TextVarDescription[*netip.AddrPort, netip.AddrPort]
//	const Listen AddrPortFlag = "listen Service {addr}:{port}."
//
// Define flag with initial value and any aliases before [flag.Parse].
//
//	var lap netip.AddrPort
//	Listen.Define(&lap, netip.AddrPortFrom(netip.IPv4Unspecified(), 8080))
type TextVarDescription[P encoding.TextUnmarshaler,
	V encoding.TextMarshaler] string

// [TextVarDescription.DefineIn] [flag.CommandLine]
func (d TextVarDescription[P, V]) Define(p P, v V, aliases ...string) {
	d.DefineIn(flag.CommandLine, p, v, aliases...)
}

func (d TextVarDescription[P, V]) DefineIn(
	flags *flag.FlagSet, p P, v V, aliases ...string,
) {
	var name, usage string
	s := d.String()
	for i, r := range s {
		if len(name) == 0 {
			if unicode.IsSpace(r) {
				name = s[:i]
			}
		} else if !unicode.IsSpace(r) {
			usage = s[i:]
			break
		}
	}
	flags.TextVar(p, name, v, usage)
	if len(aliases) > 0 {
		aka := fmt.Sprint("aka -", name)
		for _, alias := range aliases {
			flags.TextVar(p, alias, v, aka)
		}
	}
}

func (d TextVarDescription[P, V]) String() string {
	return string(d)
}
