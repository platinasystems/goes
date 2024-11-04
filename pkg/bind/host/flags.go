// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package host

import (
	"flag"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
)

var flags struct {
	a, A, C, d, i, ip4, ip6, l, m, r, s, T, U, v, V, w bool

	c xdnsmessage.Class
	N,
	p,
	R uint
	t xdnsmessage.Type
	W time.Duration
}

func addCommandLineFlags() {
	flag.BoolVar(&flags.a, "a", false,
		"is equivalent to -v -t ANY")
	flag.BoolVar(&flags.A, "A", false,
		"is like -a but omits RRSIG, NSEC, NSEC3")
	flag.TextVar(&flags.c, "c", xdnsmessage.ClassINET,
		"specifies query class for non-IN data")
	flag.BoolVar(&flags.C, "C", false,
		"compares SOA records on authoritative nameservers")
	flag.BoolVar(&flags.d, "d", false,
		"is equivalent to -v")
	flag.BoolVar(&flags.i, "i", false, "FIXME?")
	flag.BoolVar(&flags.ip4, "4", false,
		"use IPv4 query transport only")
	flag.BoolVar(&flags.ip6, "6", false,
		"use IPv6 query transport only")
	flag.BoolVar(&flags.l, "l", false,
		"lists all hosts in a domain, using AXFR")
	flag.BoolVar(&flags.m, "m", false,
		"set memory debugging flag (trace|record|usage)")
	flag.UintVar(&flags.N, "N", 0,
		"changes the number of dots allowed before root lookup is done")
	flag.UintVar(&flags.p, "p", 53,
		"specifies the port on the server to query")
	flag.BoolVar(&flags.r, "r", false,
		"disables recursive processing")
	flag.UintVar(&flags.R, "R", 3,
		"specifies number of retries for UDP packets")
	flag.BoolVar(&flags.s, "s", false,
		"a SERVFAIL response should stop query")
	flag.TextVar(&flags.t, "t", xdnsmessage.TypeA,
		"specifies the query type")
	flag.BoolVar(&flags.T, "T", false,
		"enables TCP/IP mode")
	flag.BoolVar(&flags.U, "U", true,
		"enables UDP mode")
	flag.BoolVar(&flags.v, "v", false,
		"enables verbose output")
	flag.BoolVar(&flags.V, "V", false,
		"print version number and exit")
	flag.BoolVar(&flags.w, "w", false,
		"wait forever for a reply")
	flag.DurationVar(&flags.W, "W", 30*time.Second,
		"how long to wait for a reply")
}
