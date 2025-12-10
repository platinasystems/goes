// Copyright © 2015-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build unix

package pwd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xflag"
	"golang.org/x/sys/unix"
)

const PwdUsage = `
usage: {{.Name}} [flags]

{{flags .}}`

var Pwd_L bool

var PwdFlags = xflag.Labels{
	{"L", "Display the logical current working directory.", &Pwd_L},
}

func Pwd(ctx context.Context, args []string) error {
	xflag.TemplateUsage(PwdUsage)
	err := PwdFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}
	if Pwd_L {
		if wd := os.Getenv("PWD"); strings.HasPrefix(wd, "/") {
			if wdi, wde := os.Stat(wd); wde == nil {
				if doti, dote := os.Stat("."); dote == nil {
					if os.SameFile(wdi, doti) {
						fmt.Println(wd)
						return nil
					}
				}
			}
		}
	}
	wd, err := unix.Getwd()
	if err == nil {
		fmt.Println(wd)
	}
	return err
}
