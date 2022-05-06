// Copyright © 2015-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
)

var Prog = filepath.Base(os.Args[0])

type Func = func(context.Context, ...string) error

type Selection map[string]Func

var BuiltIn = Selection{
	"build-id": func(ctx context.Context, args ...string) error {
		OutputOf(ctx).Println(ThisBuildId)
		return nil
	},
	"build-info": func(ctx context.Context, args ...string) error {
		if bi, ok := debug.ReadBuildInfo(); ok {
			OutputOf(ctx).Print(bi)
		}
		return nil
	},
	"cat":  Cat,
	"echo": Echo,
	"version": func(ctx context.Context, args ...string) error {
		if bi, ok := debug.ReadBuildInfo(); ok {
			OutputOf(ctx).Println(bi.Main.Version)
		}
		return nil
	},
}

func (m Selection) Keys() []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (m Selection) Main() {
	StyleLog()
	ctx, stop := signal.NotifyContext(context.Background(),
		TerminationSignals...)
	defer stop()
	for k, v := range BuiltIn {
		if _, ok := m[k]; !ok {
			m[k] = v
		}
	}
	ctx = WithRoot(ctx, m)
	ctx = WithInput(ctx, os.Stdin)
	ctx = WithOutput(ctx, os.Stdout)
	ctx = WithPath(ctx, Prog)
	timeout := flag.Duration("timeout", 0,
		"Terminate command if incomplete by non-zero limit.")
	flag.CommandLine.Init(Prog, flag.ContinueOnError)
	flag.Usage = func() {
		Usage(ctx, "COMMAND [OPTION]...\n",
			"\n",
			flag.CommandLine,
			m)
	}
	ctx = WithUsage(ctx, flag.Usage)
	if err := flag.CommandLine.Parse(os.Args[1:]); err == flag.ErrHelp {
		return
	}
	if *timeout != 0 {
		t, cancel := context.WithTimeout(ctx, *timeout)
		defer cancel()
		ctx = t
	}
	args := flag.Args()
	ctx, args = Preempt(ctx, args)
	f, found := m[os.Args[0]] // e.g. /init
	if !found {
		f = m.Select
	}
	defer recovery()
	if err := f(ctx, args...); err != nil {
		PlainLog()
		Fatal(err)
	}
}

func (m Selection) Select(ctx context.Context, args ...string) error {
	if len(args) == 0 {
		switch Preemption(ctx) {
		case "":
			if f, found := m[""]; found {
				return f(ctx)
			} else {
				return fmt.Errorf("incomplete")
			}
		case "complete":
			m.complete(ctx)
		case "help":
			m.usage(ctx)
		}
	} else if f, found := m[args[0]]; found {
		ctx = WithPath(ctx, args[0])
		return f(ctx, args[1:]...)
	} else {
		switch Preemption(ctx) {
		case "":
			if p, perr := exec.LookPath(args[0]); perr == nil {
				cmd := exec.Command(p, args[1:]...)
				cmd.Stdin = InputOf(ctx)
				cmd.Stdout = OutputOf(ctx)
				cmd.Stderr = os.Stderr
				return cmd.Run()
			} else {
				return perr
			}
		case "complete":
			m.complete(ctx, args...)
		case "help":
			m.usage(ctx)
		}
	}
	return nil
}

func (m Selection) complete(ctx context.Context, args ...string) {
	o := OutputOf(ctx)
	for _, s := range CompleteStrings(m.Keys(), args) {
		o.Println(s)
	}
}

func (m Selection) usage(ctx context.Context) {
	if usage := UsageOf(ctx); usage != nil {
		usage()
	} else {
		Usage(ctx, "[COMMAND [OPTION]...]...\n", m)
	}
}

func recovery() {
	r := recover()
	if r == nil {
		return
	}
	sb := new(strings.Builder)
	fmt.Fprintln(sb, r)
	pcs := make([]uintptr, 64)
	if n := runtime.Callers(2, pcs); n > 0 {
		frames := runtime.CallersFrames(pcs[:n])
		for {
			f, more := frames.Next()
			if len(f.Function) == 0 {
				break
			}
			if strings.Contains(f.File, "runtime/") {
				continue
			}
			fmt.Fprint(sb, "    ", f.Function, "()\n")
			fmt.Fprint(sb, "        ", f.File, ":", f.Line, "\n")
			if !more {
				break
			}
		}
	}
	PlainLog()
	Fatal(sb)
}
