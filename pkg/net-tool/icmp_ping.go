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

const (
	ICMPPing_c_Flag xflag.Xuint     = "c Count."
	ICMPPing_i_Flag xflag.Xduration = "i Interval."
	ICMPPing_m_Flag xflag.Xuint     = "m Request Time To Live."
	ICMPPing_q_Flag xflag.Xbool     = "q Quiet."
	ICMPPing_t_Flag xflag.Xduration = "t Timeout regardless of how many received packets."
	ICMPPing_v_Flag xflag.Xbool     = "v Verbose."
)

func ICMPPing(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [host]
Send ICMP ECHO_REQUEST packets to network “host”, default 127.0.0.1.
{{flags .}}`)

	cFlag := ICMPPing_c_Flag.Define(0)
	iFlag := ICMPPing_i_Flag.Define(time.Second)
	mFlag := ICMPPing_m_Flag.Define(0)
	qFlag := ICMPPing_q_Flag.Define(false)
	tFlag := ICMPPing_t_Flag.Define(3 * time.Second)
	vFlag := ICMPPing_v_Flag.Define(false)

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
