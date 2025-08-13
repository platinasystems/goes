// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	probing "github.com/prometheus-community/pro-bing"
)

func Ping(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] <guest>
ICMP with named guest.

{{flags .}}`)

	cFlag := 0
	iFlag := time.Second
	mFlag := 0
	qFlag := false
	tFlag := 3 * time.Second
	vFlag := false

	xflag.Define(&cFlag, "c", "Count.")
	xflag.Define(&iFlag, "i", "Interval.")
	xflag.Define(&mFlag, "m", "Request Time To Live.")
	xflag.Define(&qFlag, "q", "Quiet.")
	xflag.Define(&tFlag, "t",
		"Timeout regardless of how many received packets.")
	xflag.Define(&vFlag, "v", "Verbose.")

	defineRestFlags()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	if args = flag.Args(); len(args) == 0 {
		return xerrors.Incomplete("guest")
	}

	if err = restInit(); err != nil {
		return err
	}
	sub, err := restWhois(ctx, args[0])
	if err != nil {
		return xerrors.Label(err, "guest")
	}
	rest.CloseIdleConnections()

	ipaddr := &net.IPAddr{IP: net.IP(sub.Addr.AsSlice())}

	pinger := probing.New(args[0])
	pinger.SetIPAddr(ipaddr)

	pinger.Count = cFlag
	pinger.Interval = iFlag
	if tFlag != 0 {
		pinger.Timeout = tFlag
	}
	if mFlag != 0 {
		pinger.TTL = mFlag
	}
	if vFlag {
		pinger.OnSend = func(pkt *probing.Packet) {
			fmt.Printf("%d bytes to %v; icmp_seq=%d\n",
				pkt.Nbytes,
				pkt.IPAddr,
				pkt.Seq)
		}
	}
	if !qFlag {
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
	pinger.OnSetup = func() {
		fmt.Printf("PING %s (%v); %d data bytes\n",
			pinger.Addr(), pinger.IPAddr(), pinger.Size)
	}
	return pinger.RunWithContext(ctx)
}
