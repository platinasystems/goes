// Copyright © 2015-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"flag"
)

var (
	usageMark int
	usageKey  = &usageMark
)

// Usage prints this formatted text.
//
//	usage: PATH ARGS...
//
// Where PATH is the space separated elements pushed onto the context.
// The ARGS are fomatted with `fmt.Print` WITHOUT space separation so
// include this w/in if desired.
//
// A *flag.FlagSet arg is // printed with FlagSet.PrintDefault,
// and similarly for a Selection arg.
func Usage(ctx context.Context, args ...interface{}) {
	o := OutputOf(ctx)
	o.Print("usage:")
	p := PathOf(ctx)
	if n := len(p); n >= 2 && p[1] == "help" {
		copy(p[1:], p[2:])
		p = p[:n-1]
	}
	for _, s := range p {
		o.Print(" ", s)
	}
	end := "\n"
	if len(args) == 0 {
		o.Print(end)
		return
	}
	o.Print(" ")
	for _, v := range args {
		if fs, ok := v.(*flag.FlagSet); ok {
			end = ""
			fs.SetOutput(o)
			fs.PrintDefaults()
		} else if sel, ok := v.(Selection); ok {
			end = ""
			for _, s := range sel.Keys() {
				if len(s) > 0 {
					o.Println(" ", s)
				}
			}
		} else {
			o.Print(v)
		}
	}
	o.Print(end)
}

func UsageOf(ctx context.Context) func() {
	if v := ctx.Value(usageKey); v != nil {
		return v.(func())
	}
	return nil
}

func WithUsage(ctx context.Context, f func()) context.Context {
	return context.WithValue(ctx, usageKey, f)
}
