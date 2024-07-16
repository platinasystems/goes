// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes_util

import (
	"context"
	"flag"

	"github.com/platinasystems/goes/v2/pkg/xflag"
)

func Standby(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}}
Wait for a context interrupt or termination signal.
This will hold container open for interactive “kubectl exec ...”.
`)
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	<-ctx.Done()
	return ctx.Err()
}
