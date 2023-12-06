// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"fmt"
	"unicode"

	"github.com/platinasystems/goes/v2/pkg/context/ctxparm"
	probing "github.com/prometheus-community/pro-bing"
)

func ICMPPing(ctx context.Context, args ...string) error {
	w := ctxparm.Writer.In(ctx)
	host := "127.0.0.1"
	if len(args) > 0 {
		host = args[0]
	}
	pinger, err := probing.NewPinger(host)
	if err != nil {
		return err
	}
	pinger.Count = 3
	err = pinger.Run()
	if err != nil {
		return err
	}
	res := pinger.Statistics()
	fmt.Fprint(w, res.PacketsRecv, "/", res.PacketsSent,
		" replies/requests to ", res.Addr)
	if unicode.IsLetter(rune(res.Addr[0])) {
		fmt.Fprint(w, "(", res.IPAddr, ")")
	}
	fmt.Fprintln(w, " in avg.", res.AvgRtt)
	return nil
}
