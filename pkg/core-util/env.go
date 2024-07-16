// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package core_util

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xexec"
	"github.com/platinasystems/goes/v2/pkg/xflag"
)

func Env(ctx context.Context, complete bool, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [<var>=<val>]... [<command> <args>]
Set environment and execute command, or print environment.
`)
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	} else if args = flag.Args(); complete {
		for _, arg := range args {
			if strings.Index(arg, "=") < 1 {
				for _, s := range xexec.Match(arg) {
					fmt.Println(s)
				}
				break
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

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Env = environ
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
