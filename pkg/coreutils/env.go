// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package coreutils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/pathctx"
	"github.com/platinasystems/goes/v2/pkg/context/rctx"
	"github.com/platinasystems/goes/v2/pkg/context/selctx"
	"github.com/platinasystems/goes/v2/pkg/context/wctx"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

const EnvUsageTemplate = `
usage: {{.Path}} [<var>=<val>]... [<command> <args>]
Print or set environment variables.

Commands/Objects
{{.Commands}}`

func EnvUsageData(ctx context.Context) any {
	return struct{ Path, Commands string }{
		Path:     pathctx.StringIn(ctx),
		Commands: selctx.StringIn(ctx),
	}
}

func Env(ctx context.Context, args ...string) error {
	if flag.Search[bool]("complete") {
		return complete.Last(args, "*")
	}
	if flag.Search[bool]("help") {
		return usage.Error(EnvUsageTemplate[1:], EnvUsageData(ctx))
	}
	environ := os.Environ()
	w := wctx.Parameter.In(ctx)
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
	cmd.Stdin = rctx.Parameter.In(ctx)
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
