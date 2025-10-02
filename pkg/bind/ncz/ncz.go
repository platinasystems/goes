// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ncz

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsdb"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
	"github.com/platinasystems/goes/v2/pkg/xsync"
)

var ncz_D, ncz_d, ncz_v bool
var ncz_L, ncz_T, ncz_r, ncz_t, ncz_w string
var ncz_C = "fail"
var ncz_F = "text"
var ncz_M = "warn"
var ncz_S = "warn"
var ncz_W = "warn"
var ncz_c = xdnsmessage.ClassINET
var ncz_f = "text"
var ncz_i = "full"
var ncz_k = "warn"
var ncz_l = 0
var ncz_m = "warn"
var ncz_n = "warn"
var ncz_o = "-"
var ncz_s = "full"

var nczFlags = xflag.Labels{
	xlog.VerboseFlag,
	{"C", "Check mode: fail, or ignore.", &ncz_C},
	{"D", `Dump zone file in canonical format.`, &ncz_D},
	{"F", "Output format: text, raw, or raw=N.", &ncz_F},
	{"L", `
When compiling a zone to "raw" format, this option sets the "source
serial" value in the header to the specified serial number. This is
expected to be used primarily for testing purposes.`[1:], &ncz_L},
	{"M", `
Print whether a MX records refer to a CNAME:
	fail, warn, or ignore.`[1:], &ncz_M},
	{"S", `
Print whether an SRV record refers to a CNAME:
	fail, warn, or ignore.`[1:], &ncz_S},
	{"T", `
Checks whether Sender Policy Framework (SPF) records exist and
issues a warning if an SPF-formatted TXT record is not also present:
	fail, warn, or ignore.`[1:], &ncz_T},
	{"W", `
Print non-terminal wildcards:
	warn or ignore`[1:], &ncz_W},
	{"c", "Zone class if unspecified.", &ncz_c},
	{"d", `Enables debugging.`, &ncz_d},
	{"f", "Zone file format: text, or raw.", &ncz_f},
	{"i", `
Post-load zone integrity checks:
	full, full-sibling, local, local-sibling, or none.

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

Mode "none" disables the checks.`[1:], &ncz_i},
	{"k", "Failure mode: fail, warn, or ignore.", &ncz_k},
	{"l", "Maximum permissible TTL (0 is no max)", &ncz_l},
	{"m", "MX record check: fail, warin, or ignore.", &ncz_m},
	{"n", `
Print whether NS records are addresses:
	fail, warn, or ignore.`[1:], &ncz_n},
	{"o", `
Writes the zone output to named file or
standard output if "-".`[1:], &ncz_o},
	{"r", `
Check for records that are treated as different by DNSSEC but are
semantically equal in plain DNS:
	fail, warn, or ignore.`[1:], &ncz_r},
	{"s", `
Style of the dumped zone file:
	full or relative.`[1:], &ncz_s},
	{"t", `
If not empty, chroot to named directory so that include directives are
are processed similar to chrooted "named.`[1:], &ncz_t},
	{"v", `Prints version then exits`, &ncz_v},
	{"w", `
If not empty, chdir to named directory for relative $INCLUDE directives.
This is similar to the directory clause in named.conf.`[1:], &ncz_w},
}

func NamedCheckZone(ctx context.Context, args []string) error {
	var wg xsync.WaitGroup

	xflag.TemplateUsage(`
usage: {{.Name}} [-flags] {zone} {file | -}
Mimic BIND9's config verification tool.

{{flags .}}`)

	err := nczFlags.Define()
	if err != nil {
		return err
	} else if err := flag.CommandLine.Parse(args); err != nil {
		return err
	}
	args = flag.Args()

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		wg.Wait()
	}()

	if ncz_v {
		fmt.Println(xmain.Version())
		return nil
	}
	switch len(args) {
	case 0:
		return xerrors.Incomplete("zone")
	case 1:
		return xerrors.Incomplete("file")
	}

	zone, fn := args[0], args[1]

	wg.Go(func() { xdnsdb.Server(ctx, xlog.Info) })

	if err = xdnsdb.Include(ctx, zone, fn); err != nil {
		return err
	}

	xdnsdb.Dump(os.Stdout)
	return nil
}
