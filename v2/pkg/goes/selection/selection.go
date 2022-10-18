// Copyright © 2015-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package selection

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/goes/complete"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/termination"
)

var (
	ErrIncomplete = errors.New("incomplete")
	ErrNotFound   = errors.New("not found")
	Fatal         = style.Plain.Errata.Fatal
)

func HasHelp(args []string) bool {
	if len(args) > 0 && strings.HasPrefix(args[0], "-") {
		a0 := strings.TrimLeft(args[0], "-")
		return a0 == "h" || a0 == "help"
	}
	return false
}

type Func = func(context.Context, io.Reader, io.Writer, Path, ...string) error

type Map map[string]Func

var Root Map

func (m Map) Format(w fmt.State, verb rune) {
	for _, k := range m.Keys() {
		fmt.Fprint(w, "\n  ", k)
	}
}

func (m Map) Main() {
	r := io.Reader(os.Stdin)
	w := io.Writer(os.Stdout)
	Root = m
	ctx, stop := signal.NotifyContext(context.Background(),
		termination.Signals...)
	defer stop()
	path := Path{program.Base.String()}
	flag.CommandLine.Init(path[0], flag.ContinueOnError)
	flag.Usage = func() {
		Root.Usage(w, path, flag.CommandLine)
	}
	timeout := flag.Duration("timeout", 0,
		"Terminate command if incomplete by non-zero limit.")
	err := flag.CommandLine.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		return
	}
	if *timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}
	args := flag.Args()
	f, found := Root[program.Executable.String()] // e.g. /init
	if !found {
		f = Root.Select
	}
	isDaemon := len(args) > 0 && args[0] == "daemon"
	if isDaemon {
		style.System()
		r = io.LimitReader(nil, 0)
		w = style.Plain.Notice.Writer()
		fmt.Fprintln(w, args, "start")
	}
	err = f(ctx, r, w, path, args...)
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		Fatal(err)
	} else if isDaemon {
		fmt.Fprintln(w, args, "exit")
	}
}

func (m Map) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		if k != "daemon" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

func (m Map) Select(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path Path,
	args ...string,
) error {
	if len(args) > 0 && len(path) == 1 {
		if args[0] == "complete" || args[0] == "help" {
			path = append(path, args[0])
			args = args[1:]
		}
	}
	if len(args) == 0 {
		if f, found := m[""]; found {
			return f(ctx, r, w, path)
		}
		if path.HasComplete() {
			complete.Last(w, args, m.Keys())
			return nil
		}
		if path.HasHelp() {
			m.Usage(w, path)
			return nil
		}
		return ErrIncomplete
	}
	if f, found := m[args[0]]; found {
		err := f(ctx, r, w, append(path, args[0]), args[1:]...)
		if err != nil {
			return fmt.Errorf("%s: %w", args[0], err)
		}
		return nil
	}
	if f, found := m[""]; found {
		return f(ctx, r, w, path, args...)
	}
	if path.HasComplete() {
		complete.Last(w, args, m.Keys())
		return nil
	}
	if path.HasHelp() || HasHelp(args) {
		m.Usage(w, path)
		return nil
	}
	if f, found := m["command"]; found {
		return f(ctx, r, w, append(path, "command"), args...)
	}
	return fmt.Errorf("%q: %w", args[0], ErrNotFound)
}

func (m Map) Usage(
	w io.Writer,
	path Path,
	args ...any,
) {
	var usage string
	if len(path) == 1 {
		if _, ok := m[""]; ok {
			usage = "[<options>] [<command> [<args>]]\n"
		} else {
			usage = "[<options>] <command> [<args>]\n"
		}
	} else if _, ok := m[""]; ok {
		usage = "[<command> [<args>]]\n"
	} else {
		usage = "<command> [<args>]\n"
	}
	path.Usage(w, usage, args, m)
}
