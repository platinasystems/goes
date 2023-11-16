// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netrt

import (
	"syscall"
	"time"

	"github.com/platinasystems/goes/v2/pkg/flag"
)

func FlagSet(name string) *flag.FlagSet {
	fs := flag.New(name)
	afinet := fs.Bool("4", false, "Address hint.")
	fs.BoolVar(afinet, "inet", false, "aka -4.")
	afinet6 := fs.Bool("6", false, "Address hint.")
	fs.BoolVar(afinet6, "inet6", false, "aka -6.")
	iface := fs.Bool("interface", false, "Instead of next-hop.")
	fs.BoolVar(iface, "iface", false, "aka. -interface")
	fs.Bool("blackhole", false, "Silently discard pkts (during updates).")
	fs.Bool("cloning", false, "Generates a new route.")
	fs.Bool("d", false, "Debug mode, don't modify route table.")
	fs.Bool("host", false, "Host <destination>.")
	fs.Bool("link", false, "Address hint.")
	fs.Bool("llinfo", false, "Translates protocol to link address")
	fs.Bool("n", false, "Suppress name lookup, just print numerics.")
	fs.Bool("net", false, "Network <destination>.")
	fs.Bool("proto1", false, "")
	fs.Bool("proto2", false, "")
	fs.Bool("proxy", false, "")
	fs.Bool("q", false, "Quiet mode, suppress all output.")
	fs.Bool("reject", false, "Emit an ICMP unreachable when matched.")
	fs.Bool("static", true, "Manually added route.")
	fs.Bool("t", false, "Test mode, use /dev/null socket.")
	fs.Bool("v", false, "Verbose mode, print additional details.")
	fs.Bool("xresolve", false, "Emit mesg on use (for external lookup).")
	fs.Int("expire", 0, "Seconds from now.")
	fs.Int("fib", -1, "(-1 current)")
	fs.Int("prefixlen", -1, "Instead of 1st arg /<suffix>.")
	fs.String("dst", "", "Instead 1st position arg.")
	fs.String("gateway", "", "Instead 2nd position arg.")
	fs.String("genmask", "", "")
	fs.String("ifa", "", "")
	fs.String("ifp", "", "")
	fs.String("mask", "", "Instead 3rd position arg or 1st /<suffix>.")
	fs.Uint("hopcount", 0, "")
	fs.Uint("mtu", 1500, "")
	fs.Uint("recvpipe", 0, "")
	fs.Uint("rtt", 0, "")
	fs.Uint("rttvar", 0, "")
	fs.Uint("sendpipe", 0, "")
	fs.Uint("ssthresh", 0, "")
	fs.Uint("weight", 0, "")
	return fs
}

func setFlags(rtm *syscall.RtMsghdr, fs *flag.FlagSet, cmd uint8) {
	switch cmd {
	case syscall.RTM_ADD, syscall.RTM_CHANGE:
		rtm.Flags = syscall.RTF_UP
	case syscall.RTM_DELETE:
		rtm.Flags |= syscall.RTF_PINNED
	}
	for _, x := range []struct {
		name string
		flag int32
	}{
		{"static", syscall.RTF_STATIC},
		{"reject", syscall.RTF_REJECT},
		{"blackhole", syscall.RTF_BLACKHOLE},
		{"proto1", syscall.RTF_PROTO1},
		{"proto2", syscall.RTF_PROTO2},
		{"proxy", syscall.RTF_PROXY},
		{"xresolve", syscall.RTF_XRESOLVE},
	} {
		if flag.Eval[bool](fs, x.name) {
			rtm.Flags |= x.flag
		}
	}
	if !flag.Eval[bool](fs, "interface") {
		rtm.Flags |= syscall.RTF_GATEWAY
	}
	if mtu := flag.Eval[int](fs, "mtu"); mtu != 1500 {
		rtm.Rmx.Mtu = uint32(mtu)
		rtm.Inits |= syscall.RTV_MTU
	}
	if secs := flag.Eval[int](fs, "expire"); secs != 0 {
		elapse := time.Second * time.Duration(secs)
		expire := time.Now().Add(elapse).Unix()
		rtm.Rmx.Expire = int32(expire)
		rtm.Inits |= syscall.RTV_EXPIRE
	}
	for _, x := range []struct {
		name string
		rtv  uint32
		rmx  *uint32
	}{
		{"hopcount", syscall.RTV_HOPCOUNT, &rtm.Rmx.Hopcount},
		{"recvpipe", syscall.RTV_RPIPE, &rtm.Rmx.Recvpipe},
		{"sendpipe", syscall.RTV_SPIPE, &rtm.Rmx.Sendpipe},
		{"ssthresh", syscall.RTV_SSTHRESH, &rtm.Rmx.Ssthresh},
		{"rtt", syscall.RTV_RTT, &rtm.Rmx.Rtt},
		{"rttvar", syscall.RTV_RTTVAR, &rtm.Rmx.Rttvar},
	} {
		if v := flag.Eval[int](fs, x.name); v != 0 {
			*(x.rmx) = uint32(v)
			rtm.Inits |= x.rtv
		}
	}
}
