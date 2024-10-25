// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named_checkzone

import (
	"flag"

	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
)

var flags struct {
	d, q, v, D bool

	C, J, i, f, F, k, L, m, M, n, o, r, s, S, t, T, w, W string

	c xdnsmessage.Class
	l uint
}

func addCommandLineFlags() {
	flag.BoolVar(&flags.d, "d", false, `
Enables debugging.`[1:])
	flag.BoolVar(&flags.q, "q", false, `
Quiet mode - only set an exit code to indicate
successful or failed verification.`[1:])
	flag.BoolVar(&flags.v, "v", false, `
Prints version then exits`[1:])
	flag.TextVar(&flags.c, "c", xdnsmessage.DefaultClass, `
Zone class if unspecified`[1:])
	flag.StringVar(&flags.C, "C", "fail", `
Check mode: fail, or ignore.`[1:])
	flag.StringVar(&flags.i, "i", "full", `
Post-load zone integrity checks: full, full-sibling, local,
local-sibling, or none.

Mode "full" checks that MX records refer to A or AAAA records
(both in-zone and out-of-zone hostnames). Mode "local" only
checks MX records which refer to in-zone hostnames.

Mode "full" checks that SRV records refer to A or AAAA records
(both in-zone and out-of-zone hostnames). Mode "local" only
checks SRV records which refer to in-zone hostnames.

Mode "full" checks that delegation NS records refer to A or AAAA
records (both in-zone and out-of-zone hostnames). It also checks that
glue address records in the zone match those advertised by the child.
Mode "local" only checks NS records which refer to in-zone
hostnames or verifies that some required glue exists, i.e., when the
name server is in a child zone.

Modes "full-sibling" and "local-sibling" disable sibling glue
checks, but are otherwise the same as "full" and "local",
respectively.

Mode "none" disables the checks.`[1:])
	flag.StringVar(&flags.f, "f", "text", `
Zone file format: text, or raw.`[1:])
	flag.StringVar(&flags.F, "F", "text", `
Output format: text, raw, or raw=N.`[1:])
	flag.StringVar(&flags.k, "k", "warn", `
Failure mode: fail, warn, or ignore.`[1:])
	flag.UintVar(&flags.l, "l", 0, `
Maximum permissible TTL (0 is no max)`[1:])
	flag.StringVar(&flags.L, "L", "", `
When compiling a zone to "raw" format, this option sets the "source
serial" value in the header to the specified serial number. This is
expected to be used primarily for testing purposes.`[1:])
	flag.StringVar(&flags.m, "m", "warn", `
MX record check: fail, warin, or ignore.`[1:])
	flag.StringVar(&flags.M, "M", "warn", `
Check whether a MX records refer to a CNAME: fail, warn, ignore.`[1:])
	flag.StringVar(&flags.n, "n", "warn", `
Check whether NS records are addresses: fail, warn, or ignore.`[1:])
	flag.StringVar(&flags.o, "o", "-", `
Writes the zone output to named file or  standard output if "-".`[1:])
	flag.StringVar(&flags.r, "r", "", `
Check for records that are treated as different by DNSSEC but are
semantically equal in plain DNS: fail, warn, or ignore.`[1:])
	flag.StringVar(&flags.s, "s", "full", `
Style of the dumped zone file: full or relative.`[1:])
	flag.StringVar(&flags.S, "S", "warn", `
Checks whether an SRV record refers to a CNAME: fail, warn, ignore.`[1:])
	flag.StringVar(&flags.t, "t", "", `
If not empty, chroot to named directory so that include directives are
are processed similar to chrooted "named.`[1:])
	flag.StringVar(&flags.T, "T", "", `
Checks whether Sender Policy Framework (SPF) records exist and
issues a warning if an SPF-formatted TXT record is not also present:
fail, warn, ignore.`[1:])
	flag.StringVar(&flags.w, "w", "", `
If not empty, chdir to named directory for relative $INCLUDE directives.
This is similar to the directory clause in named.conf.`[1:])
	flag.BoolVar(&flags.D, "D", false, `
Dumps zone file in canonical format.`[1:])
	flag.StringVar(&flags.W, "W", "warn", `
Check non-terminal wildcards: warn or ignore`[1:])
}
