// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package icmp

import (
	"context"
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsdoh"
	probing "github.com/prometheus-community/pro-bing"
)

const PingUsage = `
usage: {{.Name}} [flags] [host]
Send ICMP ECHO_REQUEST packets to network “host”. (default localhost)
{{flags .}}`

const (
	PingDefaultCount    = -1
	PingDefaultInterval = time.Second
	PingDefaultTimeout  = 100000 * time.Second
)

var (
	Ping_c = PingDefaultCount
	Ping_i = PingDefaultInterval
	Ping_l = 64
	Ping_s = 24
	Ping_t = PingDefaultTimeout
	Ping_p,
	Ping_q,
	Ping_v bool
)

var PingFlags = xflag.Labels{
	xmain.ConfigFlag,
	xdnsdoh.ConfigFlag,
	{"c", "Count. (<1 infinite)", &Ping_c},
	{"i", "Interval.", &Ping_i},
	{"l", "Request Time To Live.", &Ping_l},
	{"p", "Privileged, raw ICMP.", &Ping_p},
	{"q", "Quiet.", &Ping_q},
	{"s", "Size.", &Ping_s},
	{"t", "Timeout regardless of how many received packets.", &Ping_t},
	{"v", "Verbose.", &Ping_v},
}

func Ping(ctx context.Context, args []string) error {
	xflag.TemplateUsage(PingUsage)
	err := PingFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}

	if Ping_c != PingDefaultCount &&
		Ping_t == PingDefaultTimeout &&
		Ping_i == PingDefaultInterval {
		Ping_t = time.Duration(Ping_c+1) * Ping_i
	}

	args = flag.Args()

	host := "localhost"
	resolver, err := xdnsdoh.Resolver()
	if err != nil {
		return err
	}
	if args = flag.Args(); len(args) > 0 {
		host = xdnsdoh.FQDN(args[0])
	}

	nw := "ip"
	addrs, err := resolver.LookupNetIP(ctx, nw, host)
	if err != nil {
		return err
	}
	ipaddr := &net.IPAddr{
		IP: net.IP(addrs[0].AsSlice()),
	}
	if addrs[0].Is4() {
		nw = "ip4"
	} else if addrs[0].Is6() {
		nw = "ip6"
	}

	pinger := probing.New(host)
	pinger.Count = Ping_c
	pinger.Interval = Ping_i
	pinger.Size = Ping_s
	pinger.Timeout = Ping_t
	pinger.TTL = Ping_l

	pinger.SetNetwork(nw)
	pinger.SetIPAddr(ipaddr)
	pinger.SetPrivileged(Ping_p)

	/* e.g.
	PING localhost (127.0.0.1): 56 data bytes
	64 bytes from 127.0.0.1: icmp_seq=0 ttl=64 time=0.082 ms
	64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=0.151 ms
	64 bytes from 127.0.0.1: icmp_seq=2 ttl=64 time=0.134 ms

	--- localhost ping statistics ---
	3 packets transmitted, 3 packets received, 0.0% packet loss
	round-trip min/avg/max/stddev = 0.082/0.122/0.151/0.029 ms
	*/
	if Ping_q {
		pinger.OnDuplicateRecv = onDuplicateQuiet
		pinger.OnFinish = onFinishQuiet
		pinger.OnRecv = onRecvQuiet
		pinger.OnSetup = func() {}
	} else {
		pinger.OnDuplicateRecv = onDuplicate
		pinger.OnRecv = onRecv
		pinger.OnFinish = onFinish
		pinger.OnSetup = func() { onSetup(host, pinger) }
	}
	if Ping_v {
		pinger.OnSend = onSendVerbose
	} else {
		pinger.OnSend = onSend
	}

	return pinger.RunWithContext(ctx)
}

func onDuplicate(pkt *probing.Packet) {
	fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v (DUP!)\n",
		pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.TTL)
}

func onDuplicateQuiet(pkt *probing.Packet) {}

func onFinish(stats *probing.Statistics) {
	fmt.Printf("\n--- %s ping statistics ---\n", stats.Addr)
	fmt.Printf("%d packets transmitted, %d packets received, %.1f%s\n",
		stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss,
		"% packet loss")
	fmt.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
		stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
}

func onFinishQuiet(stats *probing.Statistics) {}

func onRecv(pkt *probing.Packet) {
	fmt.Printf("%d bytes from %v; icmp_seq=%d ttl=%d time=%v\n",
		pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.TTL, pkt.Rtt)
}

func onRecvQuiet(pkt *probing.Packet) {}

func onSend(pkt *probing.Packet) {}

func onSendVerbose(pkt *probing.Packet) {
	fmt.Printf("%d bytes to %v; icmp_seq=%d\n",
		pkt.Nbytes, pkt.IPAddr, pkt.Seq)
}

func onSetup(host string, pinger *probing.Pinger) {
	fmt.Printf("PING %s (%v); %d data bytes\n",
		host, pinger.IPAddr(), pinger.Size)
}
