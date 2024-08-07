// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes_util

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

func Env(ctx context.Context, complete bool, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [<var>=<val>]... [<feature> <args>]
Set environment and perform named feature, or print environment.
`)
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	} else if args = flag.Args(); complete {
		for i, arg := range args {
			if strings.Index(arg, "=") < 1 {
				return goes.IntrinsicComplete(ctx, args[i:])
			}
		}
		return nil
	}

	environ := os.Environ()

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
			fmt.Println(env)
		}
		return nil
	}

	cmd := exec.CommandContext(ctx, xprogram.Path(), args...)
	cmd.Args[0] = xprogram.MainName()
	cmd.Env = environ
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// DaemonEnv returns select variables from the current environment that are
// daemon appropriate.
//
//   - AppData
//   - LocalAppData
//   - home
//   - HOME
//   - TMPDIR
//   - USERPROFILE
//   - XDG_CACHE_HOME
//   - XDG_CONFIG_DIRS
//   - XDG_CONFIG_HOME
//   - XDG_DATA_DIRS
//   - XDG_DATA_HOME
//   - XDG_RUNTIME_DIR
//   - XDG_STATE_HOME
func DaemonEnv() []string {
	env := []string{
		Path(),
	}
	for _, name := range []string{
		"AppData",
		"LocalAppData",
		"home",
		"HOME",
		"TMPDIR",
		"USERPROFILE",
		"XDG_CACHE_HOME",
		"XDG_CONFIG_DIRS",
		"XDG_CONFIG_HOME",
		"XDG_DATA_DIRS",
		"XDG_DATA_HOME",
		"XDG_RUNTIME_DIR",
		"XDG_STATE_HOME",
	} {
		if val, ok := os.LookupEnv(name); ok {
			env = append(env, fmt.Sprint(name, "=", val))
		}
	}
	return env
}
