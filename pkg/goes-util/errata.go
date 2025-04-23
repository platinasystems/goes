// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes_util

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xos"
)

func Errata(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [message]
Log error message, if given, or stdin.
`)
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	args = flag.Args()

	w, err := xos.OpenErrorLog()
	if err != nil {
		return err
	}
	defer w.Close()

	if len(args) > 0 {
		fmt.Fprintln(w, strings.Join(args, " "))
		return nil
	}

	data, err := io.ReadAll(os.Stdin)
	if err == nil {
		_, err = w.Write(data)
	}

	return err
}
