// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
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

func Echo(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] [message]
Print message to stdout.

{{flags .}}`)

	esc := flag.Bool("e", false, "Interpret escapes.")
	nonl := flag.Bool("n", false, "Without trailing newline.")

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	for i, arg := range flag.Args() {
		if i > 0 {
			os.Stdout.WriteString(" ")
		}
		if *esc {
			text := []byte(fmt.Sprintf(`"%s"`, arg))
			err = json.Unmarshal(text, &arg)
			if err != nil {
				break
			}
		}
		os.Stdout.WriteString(arg)
	}
	if !*nonl {
		os.Stdout.WriteString("\n")
	}
	return err
}
