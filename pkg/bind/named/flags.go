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
	z,
	Z string
	p,
	P uint
}

const defaultNamedConfig = "/etc/named.conf"

func addCommandLineFlags() {
	flag.BoolVar(&flags.ip4only, "4", false, `
Only service IPv4 host addresses.`[1:])
	flag.BoolVar(&flags.ip6only, "6", false, `
Only service IPv6 host addresses.`[1:])
	flag.StringVar(&flags.c, "c", defaultNamedConfig, `
Absolute path name of configuration file.`[1:])
	flag.BoolVar(&flags.C, "C", false, `
Print configuration and exit.`[1:])
	flag.UintVar(&flags.p, "p", 53, `
UDP port number (0 disable).`[1:])
	flag.UintVar(&flags.P, "P", 443, `
HTTPS port number (0 disable).`[1:])
	flag.BoolVar(&flags.q, "q", false, `
Quiet logging.`[1:])
	flag.BoolVar(&flags.v, "v", false, `
Verbose logging.`[1:])
	flag.BoolVar(&flags.V, "V", false, `
Print version and exit.`[1:])
	flag.StringVar(&flags.z, "z", ".", `
Default zone.`[1:])
	flag.StringVar(&flags.Z, "Z", "", `
Comma separated zone files instead of or in addition to configuation.`[1:])
}
