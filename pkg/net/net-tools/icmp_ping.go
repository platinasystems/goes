// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"errors"
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
	cFlag := flags.Uint("c", 0, "Count.")
	iFlag := flags.Duration("i", time.Second, "Interval.")
	mFlag := flags.Uint("m", 0, "IP Time To Live for outgoing packets.")
	oFlag := flags.Bool("o", false,
		"Exit successfully after receiving one reply packet.")
	qFlag := flags.Bool("q", false, "Quiet.")
	vFlag := flags.Bool("v", false, "Verbose.")
	w := goes.ContextStdout(ctx)

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

	isNumericHost := unicode.IsLetter(rune(host[0]))

	pinger, err := probing.NewPinger(host)
	if err != nil {
		return err
	}

	pinger.Count = int(*cFlag)
	pinger.Interval = *iFlag
	if *mFlag != 0 {
		pinger.TTL = int(*mFlag)
	}
	if !isNumericHost {
		if err = pinger.Resolve(); err != nil {
			return err
		}
	}
	cctx, cancel := context.WithCancel(ctx)
	/* e.g.
	PING localhost (127.0.0.1): 56 data bytes
	64 bytes from 127.0.0.1: icmp_seq=0 ttl=64 time=0.082 ms
	64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=0.151 ms
	64 bytes from 127.0.0.1: icmp_seq=2 ttl=64 time=0.134 ms

	--- localhost ping statistics ---
	3 packets transmitted, 3 packets received, 0.0% packet loss
	round-trip min/avg/max/stddev = 0.082/0.122/0.151/0.029 ms
	*/
	pinger.OnSetup = func() {
		fmt.Fprintf(w, "PING %s (%v); %d data bytes\n",
			pinger.Addr(), pinger.IPAddr(), pinger.Size)
	}
	pinger.OnSend = func(pkt *probing.Packet) {
		if *vFlag {
			fmt.Fprintf(w, "%d bytes to %v; icmp_seq=%d\n",
				pkt.Nbytes,
				pkt.IPAddr,
				pkt.Seq)
		}
	}
	pinger.OnRecv = func(pkt *probing.Packet) {
		if !*qFlag {
			fmt.Fprintf(w, "%d bytes from %v; "+
				"icmp_seq=%d ttl=%d time=%v\n",
				pkt.Nbytes,
				pkt.IPAddr,
				pkt.Seq,
				pkt.TTL,
				pkt.Rtt)
		}
		if *oFlag {
			cancel()
		}
	}
	pinger.OnFinish = func(stats *probing.Statistics) {
		if !*qFlag {
			fmt.Fprintf(w, "\n--- %s ping statistics ---\n",
				stats.Addr)
			fmt.Fprintf(w, "%d packets transmitted, "+
				"%d packets received, %.1f%% packet loss\n",
				stats.PacketsSent,
				stats.PacketsRecv,
				stats.PacketLoss)
			fmt.Fprintf(w, "round-trip min/avg/max/stddev = "+
				"%v/%v/%v/%v\n",
				stats.MinRtt,
				stats.AvgRtt,
				stats.MaxRtt,
				stats.StdDevRtt)
		}
	}
	err = pinger.RunWithContext(cctx)
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	return err

}
