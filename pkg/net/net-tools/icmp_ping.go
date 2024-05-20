// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"fmt"
	"time"
	"unicode"

	"github.com/platinasystems/goes/v2/pkg/goes"
	probing "github.com/prometheus-community/pro-bing"
)

func ICMPPing(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} [<options>] [<host>]
Send ICMP ECHO_REQUEST packets to network <host> (default 127.0.0.1).

Options{{flags .}}`
	flags := goes.ContextFlags(ctx)
	cFlag := flags.Uint("c", 3, "count")
	iFlag := flags.Duration("i", time.Second, "interval")

	if goes.ContextComplete(ctx) {
		return nil
	}
	ctx, err := goes.ParseFlagsContext(ctx, args)
	if err != nil {
		return err
	}
	if goes.ContextHelp(ctx) {
		return goes.Usage(ctx, usage)
	}

	host := "127.0.0.1"
	if args = flags.Args(); len(args) > 0 {
		host = args[0]
	}

	pinger, err := probing.NewPinger(host)
	if err != nil {
		return err
	}

	pinger.Count = int(*cFlag)
	pinger.Interval = *iFlag
	pinger.OnFinish = func(stats *probing.Statistics) {
		w := goes.ContextStdout(ctx)
		fmt.Fprint(w, stats.PacketsRecv, "/", stats.PacketsSent,
			" replies/requests to ", stats.Addr)
		if unicode.IsLetter(rune(stats.Addr[0])) {
			fmt.Fprint(w, "(", stats.IPAddr, ")")
		}
		fmt.Fprintln(w, " in avg.", stats.AvgRtt)
	}

	return pinger.RunWithContext(ctx)
}
