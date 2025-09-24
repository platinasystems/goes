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

var (
	PingCountFlag    = xflag.New[int]("c", "Count.", nil)
	PingIntervalFlag = xflag.New[time.Duration]("i", `
Interval.`[1:], func() time.Duration {
		return time.Second
	})
	PingTTLFlag     = xflag.New[int]("m", "Request Time To Live.", nil)
	PingQuietFlag   = xflag.New[bool]("q", "Quiet.", nil)
	PingTimeoutFlag = xflag.New[time.Duration]("t", `
Timeout regardless of how many received packets.
`[1:], func() time.Duration {
		return 3 * time.Second
	})
	PingVerboseFlag = xflag.New[bool]("v", "Verbose.", nil)
)

var PingFlags = append(RestFlags,
	PingCountFlag,
	PingIntervalFlag,
	PingTTLFlag,
	PingQuietFlag,
	PingTimeoutFlag,
	PingVerboseFlag)

func Ping(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] <guest>
ICMP with named guest.

{{flags .}}`)

	for _, f := range PingFlags {
		f.Define()
	}

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
	sub, err := RestWhois(ctx, args[0])
	if err != nil {
		return xerrors.Label(err, "guest")
	}
	rest.CloseIdleConnections()

	ipaddr := &net.IPAddr{IP: net.IP(sub.Addr.AsSlice())}

	pinger := probing.New(args[0])
	pinger.SetIPAddr(ipaddr)

	pinger.Count = PingCountFlag.Value()
	pinger.Interval = PingIntervalFlag.Value()
	if to := PingTimeoutFlag.Value(); to != 0 {
		pinger.Timeout = to
	}
	if ttl := PingTTLFlag.Value(); ttl != 0 {
		pinger.TTL = ttl
	}
	if PingVerboseFlag.Value() {
		pinger.OnSend = func(pkt *probing.Packet) {
			fmt.Printf("%d bytes to %v; icmp_seq=%d\n",
				pkt.Nbytes,
				pkt.IPAddr,
				pkt.Seq)
		}
	}
	if !PingQuietFlag.Value() {
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
