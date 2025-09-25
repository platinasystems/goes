// Copyright © 2022-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package core_util

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/platinasystems/goes/v2/pkg/xflag"
)

var Echo_e = xflag.New[bool]("e", "Interpret escapes.", nil)
var Echo_n = xflag.New[bool]("n", "Print without trailing newline.", nil)

var EchoFlags = []xflag.Definer{
	Echo_e,
	Echo_n,
}

func Echo(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [message]
Print message to stdout.

{{flags .}}`)

	for _, f := range EchoFlags {
		f.Define()
	}

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	for i, arg := range flag.Args() {
		if i > 0 {
			os.Stdout.WriteString(" ")
		}
		if Echo_e.Value() {
			text := []byte(fmt.Sprintf(`"%s"`, arg))
			err = json.Unmarshal(text, &arg)
			if err != nil {
				break
			}
		}
		os.Stdout.WriteString(arg)
	}
	if !Echo_n.Value() {
		os.Stdout.WriteString("\n")
	}
	return err
}
