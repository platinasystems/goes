// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import "flag"

func routeFlagsGOOS(flags *flag.FlagSet) {
	flags.Bool("blackhole", false, "Silently discard pkts (during updates).")
	flags.Bool("cloning", false, "Generates a new route.")
	flags.Bool("d", false, "Debug mode, don't modify route table.")
	flags.Bool("host", false, "Host <destination>.")
	flags.Bool("link", false, "Address hint.")
	flags.Bool("llinfo", false, "Translates protocol to link address")
	flags.Bool("n", false, "Suppress name lookup, just print numerics.")
	flags.Bool("net", false, "Network <destination>.")
	flags.Bool("proto1", false, "")
	flags.Bool("proto2", false, "")
	flags.Bool("proxy", false, "")
	flags.Bool("q", false, "Quiet mode, suppress all output.")
	flags.Bool("reject", false, "Emit an ICMP unreachable when matched.")
	flags.Bool("static", true, "Manually added route.")
	flags.Bool("t", false, "Test mode, use /dev/null socket.")
	flags.Bool("v", false, "Verbose mode, print additional details.")
	flags.Bool("xresolve", false, "Emit mesg on use (for external lookup).")
	flags.Int("expire", 0, "Seconds from now.")
	flags.Int("fib", -1, "(-1 current)")
	flags.String("genmask", "", "")
	flags.String("ifa", "", "")
	flags.String("ifp", "", "")
	flags.Uint("hopcount", 0, "")
	flags.Uint("mtu", 1500, "")
	flags.Uint("recvpipe", 0, "")
	flags.Uint("rtt", 0, "")
	flags.Uint("rttvar", 0, "")
	flags.Uint("sendpipe", 0, "")
	flags.Uint("ssthresh", 0, "")
	flags.Uint("weight", 0, "")
}
