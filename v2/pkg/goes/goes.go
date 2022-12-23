// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"encoding"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strings"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/container/slice"
	"github.com/platinasystems/goes/v2/pkg/flag/flags"
	"github.com/platinasystems/goes/v2/pkg/goes/complete"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/termination"
)

var cut = slice.Cut[string]

var (
	Input = flags.CommandLine.String("input", "",
		"Input from named file instead of stdin.")
	Output = flags.CommandLine.String("output", "",
		"Output to named file instead of stdout.")
	Json = flags.CommandLine.Bool("json", false,
		"JSON instead of text input/output.")
	Timeout = flags.CommandLine.Duration("timeout", 0,
		"Terminate if incomplete by non-zero limit.")
)

// Load Root before calling Main.
var Root = map[string]any{}

// Return sorted map keys.
func Keys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		if k != "daemon" && !strings.HasPrefix(k, "_") {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// Parse command-line flags from os.Args[1:].
// Returns flags.ErrHelp on help request.
func Parse() error {
	err := flags.CommandLine.Parse(os.Args[1:])
	if err == flags.ErrHelp {
		template.Must(template.New("usage").Parse(`
usage: {{.Prog}} [<options>] <command> [<args>]
       {{.Prog}} [<options>] <object> [<value>]
{{print .Flags}}`[1:])).Execute(os.Stderr, struct {
			Prog string
			flags.Flags
		}{
			program.Base(),
			flags.Flags{flag.CommandLine},
		})
	}
	return err
}

// If not yet done, Parse command line flags, then set i/o and create context
// before Select of Root map.
func Main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		termination.Signals...)
	defer stop()

	r := io.Reader(os.Stdin)
	w := io.Writer(os.Stdout)
	path := []string{program.Base()}

	if !flags.CommandLine.Parsed() {
		if err := Parse(); err == flags.ErrHelp {
			return
		} else if err != nil {
			style.Fatal(Error{path, err})
		}
		style.Verbosity()
	}

	if len(*Input) > 0 {
		if f, err := os.Open(*Input); err != nil {
			style.Fatal(err)
		} else {
			defer f.Close()
			r = f
		}
	}
	if len(*Output) > 0 {
		if f, err := os.Create(*Output); err != nil {
			style.Fatal(err)
		} else {
			defer f.Close()
			w = f
		}
	}
	if *Timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *Timeout)
		defer cancel()
	}

	args := flag.CommandLine.Args()
	exe := program.Executable()
	v, ok := Root[exe]
	if ok { // an executable link, e.g. /init
		path[0] = exe
	} else if len(args) == 0 {
		style.Fatal(Error{path, ErrIncomplete})
	} else {
		v = Select
	}

	switch args[0] {
	case "complete":
		path = append(path, args[0])
		if args = args[1:]; len(args) == 0 {
			complete.Last(w, args, Keys(Root),
				flags.CommandLine.FlagSet)
			return
		} else if args[0] == "help" {
			args = args[1:]
		}
	case "help":
		path = append(path, args[0])
		args = args[1:]
	case "daemon":
		style.System()
		r = io.LimitReader(nil, 0)
		w = style.Plain.Notice.Writer()
		fmt.Fprintln(w, "start", args)
		defer func() { fmt.Fprintln(w, "exit", args) }()
	}

	err := do(v, ctx, r, w, path, Root, args...)
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		style.Fatal(err)
	}
}

func Select(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	m map[string]any,
	args ...string,
) (err error) {
	const usage = `
usage:{{range .Path}} {{.}}{{end}} <command> [<args>]
      {{range .Path}} {{.}}{{end}} <object> [<value>]
{{range .Keys}}
  {{.}}{{end}}
`
	if len(args) == 0 {
		if len(path) > 1 {
			switch path[1] {
			case "complete":
				complete.Last(w, args, Keys(m))
			case "help":
				path = cut(path, 1, 1)
				err = template.Must(template.New("usage").
					Parse(usage[1:])).
					Execute(w, struct {
						Path, Keys []string
					}{
						path, Keys(m),
					})
			default:
				err = ErrIncomplete
			}
		}
	} else if v, ok := m[args[0]]; ok {
		path = append(path, args[0])
		args = args[1:]
		err = do(v, ctx, r, w, path, m, args...)
	} else if len(path) > 1 && path[1] == "complete" && len(args) == 1 {
		complete.Last(w, args, Keys(m))
	} else if v, ok := m["command"]; ok {
		path = append(path, "command")
		err = do(v, ctx, r, w, path, m, args...)
	} else {
		path = append(path, args[0])
		err = ErrNotFound
	}
	if err != nil {
		if _, wrapped := err.(Error); !wrapped {
			err = Error{path, err}
		}
	}
	return
}

func do(
	v any,
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	m map[string]any,
	args ...string,
) error {
	const marshalUsage = `
usage:{{range .}} {{.}}{{end}}
Format named object.
`
	const unmarshalUsage = `
usage:{{range .}} {{.}}{{end}} <value>
Unmarshal or scan object from text value.
`
	switch t := v.(type) {
	case map[string]any:
		return Select(ctx, r, w, path, t, args...)
	case func(
		context.Context,
		io.Reader,
		io.Writer,
		[]string,
		map[string]any,
		...string,
	) error:
		return t(ctx, r, w, path, m, args...)
	case func(
		context.Context,
		io.Reader,
		io.Writer,
		[]string,
		...string,
	) error:
		return t(ctx, r, w, path, args...)
	case func(
		context.Context,
		io.Writer,
		[]string,
		...string,
	) error:
		return t(ctx, w, path, args...)
	}
	if len(args) == 0 {
		var (
			text []byte
			err  error
		)
		if path[1] == "help" {
			return template.Must(template.New("usage").
				Parse(marshalUsage[1:])).
				Execute(w, path)
		} else if *Json {
			text, err = json.MarshalIndent(v, "", "  ")
		} else if method, ok := v.(encoding.TextMarshaler); ok {
			text, err = method.MarshalText()
		} else {
			text = []byte(fmt.Sprint(v))
		}
		if n := len(text); err == nil && n > 0 {
			w.Write(text)
			if text[n-1] != '\n' {
				w.Write([]byte{'\n'})
			}
		}
		return err
	}
	text := []byte(strings.Join(args, " "))
	if path[1] == "help" {
		return template.Must(template.New("usage").
			Parse(unmarshalUsage[1:])).
			Execute(w, cut(path, 1, 1))
	}
	if *Json {
		return json.Unmarshal(text, v)
	} else if method, ok := v.(encoding.TextUnmarshaler); ok {
		return method.UnmarshalText(text)
	}
	_, err := fmt.Sscan(args[0], v)
	return err
}
