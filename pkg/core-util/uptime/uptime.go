// Copyright © 2019-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package uptime

import (
	"context"
	"flag"

	"github.com/platinasystems/goes/v2/pkg/xflag"
)

const UptimeUsage = `
usage: {{.Name}}
Print system uptime.
`

func Uptime(ctx context.Context, args []string) error {
	xflag.TemplateUsage(UptimeUsage)
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	return uptime()
}
