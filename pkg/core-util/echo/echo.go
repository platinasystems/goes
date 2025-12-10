// Copyright © 2022-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package echo

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/platinasystems/goes/v2/pkg/xflag"
)

const EchoUsage = `
usage: {{.Name}} [flags] [message]
Print message to stdout.

{{flags .}}`

var (
	Echo_e,
	Echo_n bool
)

var EchoFlags = xflag.Labels{
	{"e", "Interpret escapes.", &Echo_e},
	{"n", "Print without trailing newline.", &Echo_n},
}

func Echo(ctx context.Context, args []string) error {
	xflag.TemplateUsage(EchoUsage)
	err := EchoFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}
	for i, arg := range flag.Args() {
		if i > 0 {
			os.Stdout.WriteString(" ")
		}
		if Echo_e {
			text := []byte(fmt.Sprintf(`"%s"`, arg))
			err = json.Unmarshal(text, &arg)
			if err != nil {
				break
			}
		}
		os.Stdout.WriteString(arg)
	}
	if !Echo_n {
		os.Stdout.WriteString("\n")
	}
	return err
}
