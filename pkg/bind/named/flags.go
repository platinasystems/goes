// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named

import "flag"

var flags struct {
	ip4only,
	ip6only,
	C,
	q,
	v,
	V bool
	c,
	p,
	T,
	z,
	Z string
}

const (
	defaultNamedConf = "/etc/named.conf"
	defaultNamedCrt  = "/etc/named.crt"
	defaultNamedKey  = "/etc/named.key"
)

// This is “quoted”.
func addCommandLineFlags() {
	flag.BoolVar(&flags.ip4only, "4", false, `
Only service IPv4 host addresses.`[1:])
	flag.BoolVar(&flags.ip6only, "6", false, `
Only service IPv6 host addresses.`[1:])
	flag.StringVar(&flags.c, "c", defaultNamedConf, `
Absolute path name of configuration file.`[1:])
	flag.BoolVar(&flags.C, "C", false, `
Print configuration and exit.`[1:])
	flag.StringVar(&flags.p, "p", "53", `
Comma separated ports on which the server will listen for queries.
If value is of the form “<portnum> or “dns=<portnum>”, the server will
listen for DNS queries on the numbered port. If value is of the form
“tls=<portnum>”, the server will listen for TLS queries on portnum;
the default is 853.  If value is of the form “https=<portnum>”,
the server will listen for HTTPS queries on portnum; the default is 443.
If value is of the form “http=<portnum>”, the server will listen for
HTTP queries on portnum; the default is 80.`[1:])
	flag.BoolVar(&flags.q, "q", false, `
Quiet logging.`[1:])
	flag.StringVar(&flags.T, "T", "", `
Commas separated “<key>[=<value>]” options.  e.g.
    -T notcp,key=/etc/named.key,cert=/etc/named.crt`[1:])
	flag.BoolVar(&flags.v, "v", false, `
Verbose logging.`[1:])
	flag.BoolVar(&flags.V, "V", false, `
Print version and exit.`[1:])
	flag.StringVar(&flags.z, "z", ".", `
Default zone.`[1:])
	flag.StringVar(&flags.Z, "Z", "", `
Comma separated zone files instead of or in addition to configuation.`[1:])
}
