// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package coreutils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/help"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

func Env(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `{{/*
*/}}usage: {{join . " "}} [<var>=<val>]... [<command> <args>]
Print or set environment variables.
`
	if complete.Parameter.Value(ctx) {
		style.Completions(args, "*")
		return nil
	}
	if help.Wanted(ctx) {
		return style.Usage(usage, path)
	}
	environ := os.Environ()
	if len(args) == 0 {
		for _, env := range environ {
			fmt.Fprintln(w, env)
		}
		return nil
	}
	for len(args) > 0 {
		eq := strings.Index(args[0], "=")
		if eq < 0 {
			break
		}
		for i, env := range environ {
			if strings.HasPrefix(env, args[0][:eq+1]) {
				environ[i] = args[0]
				eq = -1
				break
			}
		}
		if eq > 0 {
			environ = append(environ, args[0])
		}
		args = args[1:]
	}
	if len(args) == 0 {
		for _, env := range environ {
			fmt.Fprintln(w, env)
		}
		return nil
	}
	stderr := new(strings.Builder)
	cmd := exec.CommandContext(ctx, program.Executable(), args...)
	cmd.Env = environ
	cmd.Stdin = r
	cmd.Stdout = w
	cmd.Stderr = stderr
	err := cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok && stderr.Len() > 0 {
			err = errors.New(stderr.String())
		}
	}
	return err
}
