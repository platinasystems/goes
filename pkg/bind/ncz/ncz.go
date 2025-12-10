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

const NamedCheckzoneUsage = `
usage: {{.Name}} [-flags] {zone} {file | -}
Mimic BIND9's config verification tool.

{{flags .}}`

var NCZ_D, NCZ_d, NCZ_v bool
var NCZ_L, NCZ_T, NCZ_r, NCZ_t, NCZ_w string
var NCZ_C = "fail"
var NCZ_F = "text"
var NCZ_M = "warn"
var NCZ_S = "warn"
var NCZ_W = "warn"
var NCZ_c = xdnsmessage.ClassINET
var NCZ_f = "text"
var NCZ_i = "full"
var NCZ_k = "warn"
var NCZ_l = 0
var NCZ_m = "warn"
var NCZ_n = "warn"
var NCZ_o = "-"
var NCZ_s = "full"

var NamedCheckzoneFlags = xflag.Labels{
	xlog.VerboseFlag,
	{"C", "Check mode: fail, or ignore.", &NCZ_C},
	{"D", `Dump zone file in canonical format.`, &NCZ_D},
	{"F", "Output format: text, raw, or raw=N.", &NCZ_F},
	{"L", `
When compiling a zone to "raw" format, this option sets the "source
serial" value in the header to the specified serial number. This is
expected to be used primarily for testing purposes.`[1:], &NCZ_L},
	{"M", `
Print whether a MX records refer to a CNAME:
	fail, warn, or ignore.`[1:], &NCZ_M},
	{"S", `
Print whether an SRV record refers to a CNAME:
	fail, warn, or ignore.`[1:], &NCZ_S},
	{"T", `
Checks whether Sender Policy Framework (SPF) records exist and
issues a warning if an SPF-formatted TXT record is not also present:
	fail, warn, or ignore.`[1:], &NCZ_T},
	{"W", "Print non-terminal wildcards: warn or ignore", &NCZ_W},
	{"c", "Zone class if unspecified.", &NCZ_c},
	{"d", `Enables debugging.`, &NCZ_d},
	{"f", "Zone file format: text, or raw.", &NCZ_f},
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

Mode "none" disables the checks.`[1:], &NCZ_i},
	{"k", "Failure mode: fail, warn, or ignore.", &NCZ_k},
	{"l", "Maximum permissible TTL (0 is no max)", &NCZ_l},
	{"m", "MX record check: fail, warin, or ignore.", &NCZ_m},
	{"n", `
Print whether NS records are addresses:
	fail, warn, or ignore.`[1:], &NCZ_n},
	{"o", `
Writes the zone output to named file or
standard output if "-".`[1:], &NCZ_o},
	{"r", `
Check for records that are treated as different by DNSSEC but are
semantically equal in plain DNS:
	fail, warn, or ignore.`[1:], &NCZ_r},
	{"s", `
Style of the dumped zone file:
	full or relative.`[1:], &NCZ_s},
	{"t", `
If not empty, chroot to named directory so that include directives are
are processed similar to chrooted "named.`[1:], &NCZ_t},
	{"v", `Prints version then exits`, &NCZ_v},
	{"w", `
If not empty, chdir to named directory for relative $INCLUDE directives.
This is similar to the directory clause in named.conf.`[1:], &NCZ_w},
}

func NamedCheckZone(ctx context.Context, args []string) error {
	var wg xsync.WaitGroup

	xflag.TemplateUsage(NamedCheckzoneUsage)
	err := NamedCheckzoneFlags.Define()
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

	if NCZ_v {
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
