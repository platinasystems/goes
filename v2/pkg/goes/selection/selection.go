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
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/flag/flags"
	"github.com/platinasystems/goes/v2/pkg/goes/complete"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/termination"
)

var (
	ErrIncomplete = errors.New("incomplete")
	ErrNotFound   = errors.New("not found")

	Timeout = flags.CommandLine.Duration("timeout", 0,
		"Terminate if incomplete by non-zero limit.")
)

var Usage = `
usage: {{.Prog}} [<options>] <command> [<args>]
{{print .Flags}}
{{print .Selection}}`

type UsageData struct {
	Prog string
	flags.Flags
	Selection Map
}

type Map map[string]func(
	context.Context,
	io.Reader,
	io.Writer,
	[]string,
	...string,
) error

func (m Map) Format(w fmt.State, verb rune) {
	for _, k := range m.Keys() {
		fmt.Fprintln(w, " ", k)
	}
}

func (m Map) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		if k != "daemon" && !strings.HasPrefix(k, "_") {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// The hooks are run after parse of command line flags.
// Use these to extend the main Map per flag input.
func (m Map) Main(hooks ...func(Map) error) {
	r := io.Reader(os.Stdin)
	w := io.Writer(os.Stdout)
	path := []string{program.Base()}

	usage := func() {
		template.Must(template.New("usage").Parse(Usage[1:])).
			Execute(w, UsageData{
				path[0],
				flags.CommandLine,
				m,
			})
	}

	if !flags.CommandLine.Parsed() {
		err := flags.CommandLine.Parse(os.Args[1:])
		if err == flags.ErrHelp {
			usage()
			return
		} else if err != nil {
			style.Fatal(Error{path, err})
		}
		style.Verbosity()
		for _, hook := range hooks {
			if err := hook(m); err != nil {
				style.Fatal(Error{path, err})
			}
		}
	}

	args := flag.CommandLine.Args()
	if len(args) == 0 {
		style.Fatal(Error{path, ErrIncomplete})
	}

	switch args[0] {
	case "complete":
		path = append(path, args[0])
		if args = args[1:]; len(args) == 0 {
			complete.Last(w, args, m.Keys(),
				flags.CommandLine.FlagSet)
			return
		} else if args[0] == "help" {
			args = args[1:]
		}
	case "help":
		path = append(path, args[0])
		if args = args[1:]; len(args) == 0 {
			usage()
			return
		}
	case "daemon":
		style.System()
		r = io.LimitReader(nil, 0)
		w = style.Plain.Notice.Writer()
		fmt.Fprintln(w, "start", args)
		defer func() { fmt.Fprintln(w, "exit", args) }()
	}

	exe := program.Executable()
	f, found := m[exe]
	if found { // is an executable link, e.g. /init
		path[0] = exe
	} else {
		f = m.Select
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		termination.Signals...)
	defer stop()

	if *Timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *Timeout)
		defer cancel()
	}

	if err := f(ctx, r, w, path, args...); err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			style.Fatal(err)
		}
	}
}

func (m Map) Select(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) (err error) {
	if len(args) == 0 {
		switch path[1] {
		case "complete":
			complete.Last(w, args, m.Keys())
		case "help":
			copy(path[1:], path[2:])
			path = path[:len(path)-1]
			err = template.Must(template.New("usage").Parse(`
usage: {{.Command}} <command|object> [<args>]

{{print .Selection}}`[1:])).Execute(w, struct {
				Command   string
				Selection Map
			}{strings.Join(path, " "), m})
			if err != nil {
				err = Error{path, err}
			}
		default:
			err = Error{path, ErrIncomplete}
		}
	} else if f, found := m[args[0]]; found {
		err = f(ctx, r, w, append(path, args[0]), args[1:]...)
		if err != nil {
			if _, wrapped := err.(Error); !wrapped {
				err = Error{path, err}
			}
		}
	} else if f, found := m["command"]; found {
		err = f(ctx, r, w, append(path, "command"), args...)
		if err != nil {
			err = Error{append(path, args[0]), err}
		}
	} else if path[1] == "complete" && len(args) == 1 {
		complete.Last(w, args, m.Keys())
		return
	} else {
		err = Error{append(path, args[0]), ErrNotFound}
	}
	return
}
