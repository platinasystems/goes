// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package command

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"golang.org/x/term"
)

var (
	ErrMissingArg = errors.New("missing argument")
	ErrNotFound   = errors.New("command not found")
)

func Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path selection.Path,
	args ...string,
) error {
	var full string
	fs := flag.NewFlagSet("command", flag.ContinueOnError)
	p := fs.Bool("p", false, "restricted path search")
	v := fs.Bool("v", false, "report path found")
	vv := fs.Bool("V", false, "more verbose report")
	fs.Usage = func() {
		path.Usage(w, "[<options>] <command> [<args>]\n",
			"Run external command.\n",
			fs,
		)
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	args = fs.Args()
	if path.HasComplete() {
		return nil
	}
	if path.HasHelp() || selection.HasHelp(args) {
		fs.Usage()
		return nil
	}
	if len(args) == 0 {
		return ErrMissingArg
	}
	if *p {
		full, err = LookRestrictedPath(args[0])
	} else {
		full, err = exec.LookPath(args[0])
	}
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
		if fd := method.Fd(); term.IsTerminal(int(fd)) {
			cmd.Stderr = w
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Setsid:  true,
				Setctty: true,
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
