// Copyright © 2015-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mkdir

import (
	"context"
	"flag"
	"fmt"
	"os"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/xflag"
)

const MkdirUsage = `
usage: {{.Name}} [flags] <directory>...
Create directory(ies).

{{flags .}}`

const MkdirDefaultMode os.FileMode = 0777

var (
	Mkdir_m xflag.FileMode
	Mkdir_p,
	Mkdir_v bool
)

var MkdirFlags = xflag.Labels{
	{"m", `
If non-zero, set mode (as in chmod) to the given octal value;
othwise, use umasked 0777.`[1:], &Mkdir_m},
	{"p", `
Ignore off exists and make parent directories as needed.`[1:], &Mkdir_p},
	{"v", "Print each directory created.", &Mkdir_v},
}

func Mkdir(ctx context.Context, args []string) error {
	xflag.TemplateUsage(MkdirUsage)
	err := MkdirFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}

	args = flag.Args()

	mkdir := os.Mkdir
	if Mkdir_p {
		mkdir = os.MkdirAll
	}

	mode := os.FileMode(0777)
	if Mkdir_m != 0 {
		old := syscall.Umask(0)
		defer syscall.Umask(old)
		mode = Mkdir_m.Mode()
	}

	for _, dn := range args {
		if err = mkdir(dn, mode); err != nil {
			break
		}
		if Mkdir_v {
			fmt.Printf("mkdir: created directory ‘%s’\n", dn)
		}
	}
	return err
}
