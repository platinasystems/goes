// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"embed"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/context/help"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/termination"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

var (
	// Reload is called on SIGHUP.
	Reload = func() {}
	// Load Root before calling Main.
	Root = map[string]any{}
)

var cancel context.CancelFunc

// Walk embedded FS tree to add path references to map.
func EmbedFS(m map[string]any, efs embed.FS, root string) {
	walk := func(path string, entry fs.DirEntry, err error) error {
		if err == nil && entry.Type().IsRegular() {
			m[path] = efs
		}
		return err
	}
	fs.WalkDir(efs, root, walk)
}

// Execute subsystem in an interruptible context with stdin and and stdout.
func Exec(subsys any) {
	ctx, stop := signal.NotifyContext(context.Background(),
		termination.Signals...)
	defer stop()

	r := io.Reader(os.Stdin)
	w := io.Writer(os.Stdout)
	path := []string{program.Base()}
	args := os.Args[1:]

	hupch := make(chan os.Signal, 4)
	signal.Notify(hupch, os.Signal(syscall.SIGHUP))
	defer signal.Stop(hupch)
	var wg sync.WaitGroup
	wg.Add(1)
	go reload(ctx, &wg, hupch)

	ctx, cancel = context.WithCancel(ctx)

	Root["integral"] = Integral
	Merge(Root, Integral)
	if show, ok := Root["show"].(map[string]any); ok {
		if _, ok = show["build"]; !ok {
			show["build"] = program.Build
		}
		if _, ok = show["main"]; !ok {
			show["main"] = program.Main
		}
		if _, ok = show["completion"]; !ok {
			show["completion"] = IntegralShowCompletion
		}
	}

	err := do(subsys, ctx, r, w, path, Root, args...)
	if err != nil {
		if _, wrapped := err.(Error); !wrapped {
			err = Error{path, err}
		}
		style.Fatal(err)
	}
}

func Main() { Exec(Select) }

func Merge(to, from map[string]any) {
	for k, v := range from {
		if _, ok := to[k]; !ok {
			to[k] = v
		}
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
	if len(args) == 0 {
		if complete.Parameter.Value(ctx) {
			style.Completions(args, m)
		} else if help.Parameter.Value(ctx) {
			path = append(path, "help")
			return IntegralHelp(ctx, r, w, path, m)
		} else {
			err = ErrIncomplete
		}
	} else if v, ok := m[args[0]]; ok {
		path = append(path, args[0])
		err = do(v, ctx, r, w, path, m, args[1:]...)
	} else if len(args) == 1 && complete.Parameter.Value(ctx) {
		style.Completions(args, m)
	} else if len(path) == 1 {
		err = IntegralCommand(ctx, r, w, path, args...)
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
) (err error) {
	const marshalUsage = `{{/*
*/}}usage: {{join . " "}} [-json]
Format named object.
`
	const unmarshalUsage = `{{/*
*/}}usage: {{join . " "}} [-json] <value>
Unmarshal or scan object from text value.
`
	var text []byte
	switch t := v.(type) {
	case map[string]any:
		return Select(ctx, r, w, path, t, args...)
	case embed.FS:
		return IntegralShowFS(ctx, w, path, t, args...)
	case func(
		context.Context,
		io.Reader,
		io.Writer,
		[]string,
		map[string]any,
		...string,
	) error:
		err = t(ctx, r, w, path, m, args...)
	case func(
		context.Context,
		io.Reader,
		io.Writer,
		[]string,
		...string,
	) error:
		err = t(ctx, r, w, path, args...)
	case func(
		context.Context,
		io.Writer,
		[]string,
		...string,
	) error:
		err = t(ctx, w, path, args...)
	case func(
		context.Context,
		[]string,
		...string,
	) error:
		err = t(ctx, path, args...)
	case func(
		context.Context,
		...string,
	) error:
		err = t(ctx, args...)
	case func() ([]byte, error):
		text, err = t()
		if n := len(text); err == nil && n > 0 {
			fwriteln(w, text)
		}
	default:
		// Show or set objects
		nargs := len(args)
		if help.Parameter.Value(ctx) {
			if nargs == 0 || (nargs == 1 && args[0] == "-json") {
				err = style.Usage(marshalUsage, path)
			} else {
				err = style.Usage(unmarshalUsage, path)
			}
		} else if nargs == 0 {
			if method, ok := v.(encoding.TextMarshaler); ok {
				text, err = method.MarshalText()
			} else if f, ok := v.(func() string); ok {
				text = []byte(f())
			} else if f, ok := v.(func() fmt.Stringer); ok {
				text = []byte(f().String())
			} else {
				text = []byte(fmt.Sprint(v))
			}
			if n := len(text); err == nil && n > 0 {
				fwriteln(w, text)
			}
		} else if args[0] != "-json" {
			text = []byte(strings.Join(args, " "))
			if method, ok := v.(encoding.TextUnmarshaler); ok {
				err = method.UnmarshalText(text)
			} else {
				_, err = fmt.Sscan(args[0], v)
			}
		} else if nargs == 1 {
			text, err = json.MarshalIndent(v, "", "  ")
			if n := len(text); err == nil && n > 0 {
				fwriteln(w, text)
			}
		} else {
			text := []byte(strings.Join(args[1:], " "))
			err = json.Unmarshal(text, v)
		}
	}
	return
}

func fwriteln(w io.Writer, text []byte) {
	w.Write(text)
	if text[len(text)-1] != '\n' {
		w.Write([]byte{'\n'})
	}
}

func reload(ctx context.Context, wg *sync.WaitGroup, hupch <-chan os.Signal) {
	wg.Done()
	for {
		select {
		case _, ok := <-hupch:
			if !ok {
				return
			}
			Reload()
		case <-ctx.Done():
			return
		}
	}
}
