// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"flag"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xflag"
)

func IntrinsicComplete(ctx context.Context, args []string) error {
	const tmpl = `
usage: {{.Name}} <feature> [args]
Print last arg prefix matches.
`
	if len(args) > 0 && args[0] == "-h" {
		xflag.TemplateUsage(tmpl)
		flag.CommandLine.Usage()
		return nil
	}
	return Do(ctx, Complete, Features, args)
}

func IntrinsicHelp(ctx context.Context, complete bool, args []string) error {
	const tmpl = `
usage: {{.Name}} <feature> [args]
Print feature usage.
`
	if len(args) > 0 && args[0] == "-h" {
		xflag.TemplateUsage(tmpl)
		flag.CommandLine.Usage()
		return nil
	}
	name := flag.CommandLine.Name()
	name = strings.TrimSuffix(name, " help")
	xflag.Rename(flag.CommandLine, name)
	preempt := Help
	if complete {
		preempt = Complete
	}
	return Do(ctx, preempt, Features, args)
}
