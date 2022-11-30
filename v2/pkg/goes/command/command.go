// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/flag/flags"
	"github.com/platinasystems/goes/v2/pkg/goes/complete"
	"golang.org/x/term"
)

var (
	ErrMissingArg = errors.New("missing <command>")
	ErrNotFound   = errors.New("<command> not found")
)

func Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	fs := flags.New()
	p := fs.Bool("p", false, "Restricted path search.")
	v := fs.Bool("v", false, "Report path found.")
	vv := fs.Bool("V", false, "More verbose report.")
	usage := func() error {
		return template.Must(template.New("usage").Parse(`
usage: {{.Command}} [<options>] <command> [<args>]
Run external command.
{{print .Flags}}`[1:])).Execute(w, struct {
			Command string
			Flags   flags.Flags
		}{
			strings.Join(path, " "),
			fs,
		})
	}
	switch path[1] {
	case "complete":
		complete.Last(w, args, fs.FlagSet)
		return nil
	case "help":
		copy(path[1:], path[2:])
		path = path[:len(path)-1]
		return usage()
	}
	err := fs.Parse(args)
	if err == flags.ErrHelp {
		return usage()
	} else if err != nil {
		return err
	}
	if args = fs.Args(); len(args) == 0 {
		return ErrMissingArg
	}
	lookpath := exec.LookPath
	if *p {
		lookpath = LookRestrictedPath
	}
	full, err := lookpath(args[0])
	if err != nil {
		return err
	}
	if *v {
		fmt.Fprintln(w, full)
		return nil
	}
	if *vv {
		fmt.Fprintln(w, args[0], "is", full)
		return nil
	}
	stderr := new(strings.Builder)
	cmd := exec.CommandContext(ctx, full, args[1:]...)
	cmd.Stdin = r
	cmd.Stdout = w
	cmd.Stderr = stderr
	if method, ok := r.(interface{ Fd() uintptr }); ok {
		if fd := int(method.Fd()); term.IsTerminal(fd) {
			cmd.Stderr = w
			switch runtime.GOOS {
			case "linux":
				cmd.SysProcAttr = &syscall.SysProcAttr{
					Setsid:  true,
					Setctty: true,
				}
			case "darwin":
				cmd.SysProcAttr = &syscall.SysProcAttr{
					Setsid: true,
					// FIXME can't Setctty on darwin
					// Setctty: true,
					// Ctty:    0,
				}
			}
		}
	}
	if err = cmd.Start(); err == nil {
		err = cmd.Wait()
	}
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok && stderr.Len() > 0 {
			err = errors.New(stderr.String())
		}
	}
	return err
}

func LookRestrictedPath(name string) (string, error) {
	if strings.ContainsRune(name, filepath.Separator) {
		if IsExecutable(name) {
			return name, nil
		}
		return "", ErrNotFound
	}
	rpath, err := RestrictedPath.ValErr()
	if err != nil {
		return "", err
	}
	for _, d := range rpath {
		if full := filepath.Join(d, name); IsExecutable(full) {
			return full, nil
		}
	}
	return "", ErrNotFound
}
