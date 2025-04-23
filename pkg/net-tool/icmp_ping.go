// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tool

import (
	"context"
	"flag"
	"fmt"
	"time"
	"unicode"

	"github.com/platinasystems/goes/v2/pkg/xflag"
	probing "github.com/prometheus-community/pro-bing"
)

func ICMPPing(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [host]
Send ICMP ECHO_REQUEST packets to network “host”, default 127.0.0.1.
{{flags .}}`)

	cFlag := flag.Uint("c", 0, "Count.")

	iFlag := flag.Duration("i", time.Second, "Interval.")
	mFlag := flag.Uint("m", 0, "IP Time To Live for outgoing packets.")
	tFlag := flag.Duration("t", 3*time.Second, `
Timeout before ping exits, regardless of how many received packets.`[1:])
	qFlag := flag.Bool("q", false, "Quiet.")
	vFlag := flag.Bool("v", false, "Verbose.")

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	args = flag.Args()

	host := "127.0.0.1"
	if args = flag.Args(); len(args) > 0 {
		host = args[0]
	}

	isNumericHost := unicode.IsLetter(rune(host[0]))

	pinger, err := probing.NewPinger(host)
	if err != nil {
		return err
	}

	pinger.Count = int(*cFlag)
	pinger.Interval = *iFlag

	if *tFlag != 0 {
		pinger.Timeout = *tFlag
	}
	if *mFlag != 0 {
		pinger.TTL = int(*mFlag)
	}
	if !isNumericHost {
		if err = pinger.Resolve(); err != nil {
			return err
		}
	}

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
		fmt.Printf("PING %s (%v); %d data bytes\n",
			pinger.Addr(), pinger.IPAddr(), pinger.Size)
	}
	if *vFlag {
		pinger.OnSend = func(pkt *probing.Packet) {
			fmt.Printf("%d bytes to %v; icmp_seq=%d\n",
				pkt.Nbytes,
				pkt.IPAddr,
				pkt.Seq)
		}
	}
	if !*qFlag {
		pinger.OnRecv = func(pkt *probing.Packet) {
			fmt.Printf("%d bytes from %v; "+
				"icmp_seq=%d ttl=%d time=%v\n",
				pkt.Nbytes,
				pkt.IPAddr,
				pkt.Seq,
				pkt.TTL,
				pkt.Rtt)
		}
		pinger.OnFinish = func(stats *probing.Statistics) {
			fmt.Printf("\n--- %s ping statistics ---\n",
				stats.Addr)
			fmt.Printf("%d packets transmitted, "+
				"%d packets received, %.1f%% packet loss\n",
				stats.PacketsSent,
				stats.PacketsRecv,
				stats.PacketLoss)
			fmt.Printf("round-trip min/avg/max/stddev = "+
				"%v/%v/%v/%v\n",
				stats.MinRtt,
				stats.AvgRtt,
				stats.MaxRtt,
				stats.StdDevRtt)
		}
	}
	return pinger.RunWithContext(ctx)
}
