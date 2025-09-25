// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bind

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsdb"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
	"github.com/platinasystems/goes/v2/pkg/xsync"
)

var NCZ_C = xflag.New[string]("C", "Check mode: fail, or ignore.",
	func() string { return "fail" })
var NCZ_D = xflag.New[bool]("D", `Dump zone file in canonical format.`, nil)
var NCZ_F = xflag.New[string]("F", "Output format: text, raw, or raw=N.",
	func() string { return "text" })
var NCZ_L = xflag.New[string]("L", `
When compiling a zone to "raw" format, this option sets the "source
serial" value in the header to the specified serial number. This is
expected to be used primarily for testing purposes.`[1:],
	nil)
var NCZ_M = xflag.New[string]("M", `
Print whether a MX records refer to a CNAME: fail, warn, or ignore.`[1:],
	func() string { return "warn" })
var NCZ_S = xflag.New[string]("S", `
Print whether an SRV record refers to a CNAME: fail, warn, or ignore.`[1:],
	func() string { return "warn" })
var NCZ_T = xflag.New[string]("T", `
Checks whether Sender Policy Framework (SPF) records exist and
issues a warning if an SPF-formatted TXT record is not also present:
fail, warn, ignore.`[1:],
	nil)
var NCZ_W = xflag.New[string]("W", `
Print non-terminal wildcards: warn or ignore`[1:],
	func() string { return "warn" })
var NCZ_c = xflag.New[xdnsmessage.Class]("c", "Zone class if unspecified.",
	func() xdnsmessage.Class { return xdnsmessage.ClassINET })
var NCZ_d = xflag.New[bool]("d", `Enables debugging.`, nil)
var NCZ_f = xflag.New[string]("f", "Zone file format: text, or raw.",
	func() string { return "text" })
var NCZ_i = xflag.New[string]("i", `
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

Mode "none" disables the checks.`[1:],
	func() string { return "full" })
var NCZ_k = xflag.New[string]("k", "Failure mode: fail, warn, or ignore.",
	func() string { return "warn" })
var NCZ_l = xflag.New[int]("l", "Maximum permissible TTL (0 is no max)", nil)
var NCZ_m = xflag.New[string]("m", "MX record check: fail, warin, or ignore.",
	func() string { return "warn" })
var NCZ_n = xflag.New[string]("n", `
Print whether NS records are addresses: fail, warn, or ignore.`[1:],
	func() string { return "warn" })
var NCZ_o = xflag.New[string]("o", `
Writes the zone output to named file or standard output if "-".`[1:],
	func() string { return "-" })
var NCZ_q = xflag.New[bool]("q", `
Quiet mode - only set an exit code to indicate
successful or failed verification.`[1:],
	nil)
var NCZ_r = xflag.New[string]("r", `
Check for records that are treated as different by DNSSEC but are
semantically equal in plain DNS: fail, warn, or ignore.`[1:],
	nil)
var NCZ_s = xflag.New[string]("s", `
Style of the dumped zone file: full or relative.`[1:],
	func() string { return "full" })
var NCZ_t = xflag.New[string]("t", `
If not empty, chroot to named directory so that include directives are
are processed similar to chrooted "named.`[1:],
	nil)
var NCZ_v = xflag.New[bool]("v", `Prints version then exits`, nil)
var NCZ_w = xflag.New[string]("w", `
If not empty, chdir to named directory for relative $INCLUDE directives.
This is similar to the directory clause in named.conf.`[1:],
	nil)

var NCZFlags = []xflag.Definer{
	NCZ_C,
	NCZ_D,
	NCZ_F,
	NCZ_L,
	NCZ_M,
	NCZ_S,
	NCZ_T,
	NCZ_W,
	NCZ_c,
	NCZ_d,
	NCZ_f,
	NCZ_i,
	NCZ_k,
	NCZ_m,
	NCZ_n,
	NCZ_o,
	NCZ_q,
	NCZ_r,
	NCZ_s,
	NCZ_t,
	NCZ_v,
	NCZ_w,
}

func NCZ(ctx context.Context, args []string) error {
	var wg xsync.WaitGroup

	xflag.TemplateUsage(`
usage: {{.Name}} [-flags] {zone} {file | -}
Mimic BIND9's config verification tool.

{{flags .}}`)

	for _, f := range NCZFlags {
		f.Define()
	}

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	args = flag.Args()

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		wg.Wait()
	}()

	if NCZ_v.Value() {
		fmt.Println(Version())
		return nil
	}
	if NCZ_q.Value() {
		verbose = xlog.Mute(verbose)
	}
	switch len(args) {
	case 0:
		return xerrors.Incomplete("zone")
	case 1:
		return xerrors.Incomplete("file")
	}

	zone, fn := args[0], args[1]

	wg.Go(func() { xdnsdb.Server(ctx, verbose) })

	if err = xdnsdb.Include(ctx, zone, fn); err != nil {
		return err
	}

	xdnsdb.Dump(os.Stdout)
	return nil
}
