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

func NCZ(ctx context.Context, args []string) error {
	var wg xsync.WaitGroup

	xflag.TemplateUsage(`
usage: {{.Name}} [-flags] {zone} {file | -}
Mimic BIND9's config verification tool.

{{flags .}}`)

	defineNCZFlags()

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

	if ncz_v {
		fmt.Println(Version())
		return nil
	}
	if ncz_q {
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

var (
	// NCX Flags
	ncz_C = "fail"
	ncz_D = false
	ncz_F = "text"
	ncz_L = ""
	ncz_M = "warn"
	ncz_S = "warn"
	ncz_T = ""
	ncz_W = "warn"
	ncz_c = xdnsmessage.ClassINET
	ncz_d = false
	ncz_f = "text"
	ncz_i = "full"
	ncz_k = "warn"
	ncz_l = 0
	ncz_m = "warn"
	ncz_n = "warn"
	ncz_o = "-"
	ncz_q = false
	ncz_r = ""
	ncz_s = "full"
	ncz_t = ""
	ncz_v = false
	ncz_w = ""
)

func defineNCZFlags() {
	xflag.Define(&ncz_C, "C", `Check mode: fail, or ignore.`)
	xflag.Define(&ncz_D, "D", `Dumps zone file in canonical format.`)
	xflag.Define(&ncz_F, "F", `Output format: text, raw, or raw=N.`)
	xflag.Define(&ncz_L, "L", `
When compiling a zone to "raw" format, this option sets the "source
serial" value in the header to the specified serial number. This is
expected to be used primarily for testing purposes.`[1:])
	xflag.Define(&ncz_M, "M", `
Check whether a MX records refer to a CNAME: fail, warn, ignore.`[1:])
	xflag.Define(&ncz_S, "S", `
Checks whether an SRV record refers to a CNAME: fail, warn, ignore.`[1:])
	xflag.Define(&ncz_T, "T", `
Checks whether Sender Policy Framework (SPF) records exist and
issues a warning if an SPF-formatted TXT record is not also present:
fail, warn, ignore.`[1:])
	xflag.Define(&ncz_W, "W", `
Check non-terminal wildcards: warn or ignore`[1:])
	xflag.Define(&ncz_c, "c", `Zone class if unspecified`)
	xflag.Define(&ncz_d, "d", `Enables debugging.`)
	xflag.Define(&ncz_f, "f", `Zone file format: text, or raw.`)
	xflag.Define(&ncz_i, "i", `
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
	xflag.Define(&ncz_k, "k", `Failure mode: fail, warn, or ignore.`)
	xflag.Define(&ncz_l, "l", `Maximum permissible TTL (0 is no max)`)
	xflag.Define(&ncz_m, "m", `MX record check: fail, warin, or ignore.`)
	xflag.Define(&ncz_n, "n", `
Check whether NS records are addresses: fail, warn, or ignore.`[1:])
	xflag.Define(&ncz_o, "o", `
Writes the zone output to named file or  standard output if "-".`[1:])
	xflag.Define(&ncz_q, "q", `
Quiet mode - only set an exit code to indicate
successful or failed verification.`[1:])
	xflag.Define(&ncz_r, "r", `
Check for records that are treated as different by DNSSEC but are
semantically equal in plain DNS: fail, warn, or ignore.`[1:])
	xflag.Define(&ncz_s, "s", `
Style of the dumped zone file: full or relative.`)
	xflag.Define(&ncz_t, "t", `
If not empty, chroot to named directory so that include directives are
are processed similar to chrooted "named.`[1:])
	xflag.Define(&ncz_v, "v", `Prints version then exits`)
	xflag.Define(&ncz_w, "w", `
If not empty, chdir to named directory for relative $INCLUDE directives.
This is similar to the directory clause in named.conf.`[1:])
}
