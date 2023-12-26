// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || freebsd || netbsd || openbsd

package netrt

import (
	"flag"

	"github.com/platinasystems/goes/v2/pkg/flag/flagset"
	"github.com/platinasystems/goes/v2/pkg/integer"
)

func AddFlags(opts *flag.FlagSet) {
	if RTF_BLACKHOLE != 0 {
		opts.Bool("blackhole", false, "Silently discard packets.")
	}
	if RTF_CLONING != 0 {
		opts.Bool("cloning", false, "Generates a new route.")
	}
	if RTF_GATEWAY != 0 {
		iface := opts.Bool("interface", false,
			"<gateway> is a point-to-point interface name")
		opts.BoolVar(iface, "iface", *iface, "aka -interface")
	}
	if RTF_PROTO1 != 0 {
		opts.Bool("proto1", false, "FIXME")
	}
	if RTF_PROTO2 != 0 {
		opts.Bool("proto2", false, "FIXME")
	}
	if RTF_ANNOUNCE != 0 {
		opts.Bool("proxy", false, "FIXME")
	}
	if RTF_REJECT != 0 {
		opts.Bool("reject", false, "Emit ICMP unreachable.")
	}
	if RTF_STATIC != 0 {
		opts.Bool("static", true, "Manually added route.")
	}
	if RTF_STICKY != 0 {
		opts.Bool("sticky", false, "FIXME")
	}
	if RTF_XRESOLVE != 0 {
		opts.Bool("xresolve", false, "Emit mesg on use.")
	}
}

func SetFlags(rtm *RtMsghdr, opts *flag.FlagSet) {
	if RTF_BLACKHOLE != 0 && flagset.Search[bool](opts, "blackhole") {
		integer.Set(&rtm.Flags, RTF_BLACKHOLE)
	}
	if RTF_CLONING != 0 && flagset.Search[bool](opts, "cloning") {
		integer.Set(&rtm.Flags, RTF_CLONING)
	}
	if RTF_PROTO1 != 0 && flagset.Search[bool](opts, "proto1") {
		integer.Set(&rtm.Flags, RTF_PROTO1)
	}
	if RTF_PROTO2 != 0 && flagset.Search[bool](opts, "proto2") {
		integer.Set(&rtm.Flags, RTF_PROTO2)
	}
	if RTF_ANNOUNCE != 0 && flagset.Search[bool](opts, "proxy") {
		integer.Set(&rtm.Flags, RTF_ANNOUNCE)
	}
	if RTF_REJECT != 0 && flagset.Search[bool](opts, "reject") {
		integer.Set(&rtm.Flags, RTF_REJECT)
	}
	if RTF_XRESOLVE != 0 && flagset.Search[bool](opts, "xresolve") {
		integer.Set(&rtm.Flags, RTF_XRESOLVE)
	}
	if RTF_STATIC != 0 {
		if flagset.Search[bool](opts, "static") {
			integer.Set(&rtm.Flags, RTF_STATIC)
		} else {
			integer.Reset(&rtm.Flags, RTF_STATIC)
		}
	}
	if RTF_STICKY != 0 {
		if flagset.Search[bool](opts, "sticky") {
			integer.Set(&rtm.Flags, RTF_STICKY)
		} else {
			integer.Reset(&rtm.Flags, RTF_STICKY)
		}
	}
}
