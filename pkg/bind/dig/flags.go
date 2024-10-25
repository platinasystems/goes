// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package dig

import (
	"flag"

	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
)

func newFlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("dig", flag.ContinueOnError)
	xflag.UsageTemplate(fs, "[-flags] [name [TYPE] [CLASS] [+options]]")
	return fs
}

var flags struct {
	ip4, ip6, m, O, T, u, v bool

	b, f, i, k, x, y string

	c xdnsmessage.Class
	q xdnsmessage.Name
	t xdnsmessage.Type
	p uint
}

func addCommandLineFlags() {
	flag.BoolVar(&flags.ip4, "4", false, "Use IPv4 only.")
	flag.BoolVar(&flags.ip6, "6", false, "Use IPv6 only.")
	flag.BoolVar(&flags.m, "m", false, "Enable memory usage debugging.")
	flag.BoolVar(&flags.O, "O", false, `
Print plus (+) prefaced options and exit.`[1:])
	flag.BoolVar(&flags.T, "T", false, `
Print types and exit.`[1:])
	flag.BoolVar(&flags.v, "v", false, `
Print the version number and exit.`[1:])
	flag.UintVar(&flags.p, "p", 53, `
Send the query to a non-standard port on the server, instead
of the defaut port 53. This option would be used to test a
name server that has been configured to listen for queries
on a non-standard port number.`[1:])
	flag.StringVar(&flags.b, "b", "", `
Set the source IP address of the query. The address must be a
valid address on one of the host's network interfaces, or
"0.0.0.0" or "::". An optional port may be specified by
appending "#<port>"`[1:])
	flag.StringVar(&flags.f, "f", "", `
Batch mode: dig reads a list of lookup requests to process from
the given file. Each line in the file should be organized in the
same way they would be presented as queries to dig using the
command-line interface.`[1:])
	flag.StringVar(&flags.k, "k", "", `
Sign queries using TSIG using a key read from the given file.
Key files can be generated using tsig-keygen(8). When using TSIG
authentication with dig, the name server that is queried needs
to know the key and algorithm that is being used. In BIND, this
is done by providing appropriate key and server statements in
named.conf.`[1:])
	flag.BoolVar(&flags.u, "u", false, `
This option indicates that print query times should be provided in microseconds
instead of milliseconds.`[1:])
	flag.StringVar(&flags.y, "y", "", `
Sign queries using TSIG with the given authentication key.
keyname is the name of the key, and secret is the base64 encoded
shared secret.  hmac is the name of the key algorithm; valid
choices are hmac-md5, hmac-sha1, hmac-sha224, hmac-sha256,
hmac-sha384, or hmac-sha512. If hmac is not specified, the
default is hmac-md5 or if MD5 was disabled hmac-sha256.

NOTE: You should use the -k option and avoid the -y option,
because with -y the shared secret is supplied as a command line
argument in clear text. This may be visible in the output from
ps(1) or in a history file maintained by the user's shell.`[1:])
}

func addQueryFlags(fs *flag.FlagSet) {
	fs.TextVar(&flags.c, "c", xdnsmessage.Class0, `
Set the query class. { ANY, CH, CS, HS, IN }`[1:])
	fs.StringVar(&flags.i, "i", "", `
Do reverse IPv6 lookups using the obsolete RFC1886 IP6.INT
domain, which is no longer in use. Obsolete bit string label
queries (RFC2874) are not attempted.`[1:])
	flags.q.Length = 0
	fs.TextVar(&flags.q, "q", flags.q, `
Query the flagged name instead of positional argument.`[1:])
	fs.TextVar(&flags.t, "t", xdnsmessage.Type0, `
The resource record type to query. It can be any valid query
type which is supported in BIND 9. The default query type is
"A", unless the -x option is supplied to indicate a reverse
lookup. A zone transfer can be requested by specifying a type of
AXFR. When an incremental zone transfer (IXFR) is required, set
the type to ixfr=N. The incremental zone transfer will contain
the changes made to the zone since the serial number in the
zone's SOA record was N.`[1:])
	fs.StringVar(&flags.x, "x", "", `
Simplified reverse lookups, for mapping addresses to names. The
addr is an IPv4 address in dotted-decimal notation, or a
colon-delimited IPv6 address. When the -x is used, there is no
need to provide the name, class and type arguments.  dig
automatically performs a lookup for a name like
94.2.0.192.in-addr.arpa and sets the query type and class to PTR
and IN respectively. IPv6 addresses are looked up using nibble
format under the IP6.ARPA domain (but see also the -i option).`[1:])
}
