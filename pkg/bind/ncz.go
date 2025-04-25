// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bind

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsdb"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

const (
	NCZ_C_Flag xflag.Xstring = `C Check mode: fail, or ignore.`
	NCZ_D_Flag xflag.Xbool   = `D Dumps zone file in canonical format.`
	NCZ_F_Flag xflag.Xstring = `F Output format: text, raw, or raw=N.`
	NCZ_L_Flag xflag.Xstring = `L
When compiling a zone to "raw" format, this option sets the "source
serial" value in the header to the specified serial number. This is
expected to be used primarily for testing purposes.`
	NCZ_M_Flag xflag.Xstring = `M
Check whether a MX records refer to a CNAME: fail, warn, ignore.`
	NCZ_S_Flag xflag.Xstring = `S
Checks whether an SRV record refers to a CNAME: fail, warn, ignore.`
	NCZ_T_Flag xflag.Xstring = `T
Checks whether Sender Policy Framework (SPF) records exist and
issues a warning if an SPF-formatted TXT record is not also present:
fail, warn, ignore.`
	NCZ_W_Flag xflag.Xstring = `W
Check non-terminal wildcards: warn or ignore`
	NCZ_c_Flag BindClassFlag = `c Zone class if unspecified`
	NCZ_d_Flag xflag.Xbool   = `d Enables debugging.`
	NCZ_f_Flag xflag.Xstring = `f Zone file format: text, or raw.`
	NCZ_i_Flag xflag.Xstring = `i
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

Mode "none" disables the checks.`
	NCZ_k_Flag xflag.Xstring = `k Failure mode: fail, warn, or ignore.`
	NCZ_l_Flag xflag.Xuint   = `l Maximum permissible TTL (0 is no max)`
	NCZ_m_Flag xflag.Xstring = `m MX record check: fail, warin, or ignore.`
	NCZ_n_Flag xflag.Xstring = `n
Check whether NS records are addresses: fail, warn, or ignore.`
	NCZ_o_Flag xflag.Xstring = `o
Writes the zone output to named file or  standard output if "-".`
	NCZ_q_Flag xflag.Xbool = `q
Quiet mode - only set an exit code to indicate
successful or failed verification.`
	NCZ_r_Flag xflag.Xstring = `r
Check for records that are treated as different by DNSSEC but are
semantically equal in plain DNS: fail, warn, or ignore.`
	NCZ_s_Flag xflag.Xstring = `s
Style of the dumped zone file: full or relative.`
	NCZ_t_Flag xflag.Xstring = `t
If not empty, chroot to named directory so that include directives are
are processed similar to chrooted "named.`
	NCZ_v_Flag xflag.Xbool   = `v Prints version then exits`
	NCZ_w_Flag xflag.Xstring = `w
If not empty, chdir to named directory for relative $INCLUDE directives.
This is similar to the directory clause in named.conf.`
)

func NCZ(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [-flags] {zone} {file | -}
Mimic BIND9's config verification tool.

{{flags .}}`)

	NCZ_C_Flag.Define("fail")
	NCZ_D_Flag.Define(false)
	NCZ_F_Flag.Define("text")
	NCZ_L_Flag.Define("")
	NCZ_M_Flag.Define("warn")
	NCZ_S_Flag.Define("warn")
	NCZ_T_Flag.Define("")
	NCZ_W_Flag.Define("warn")
	NCZ_c_Flag.Define(xdnsmessage.ClassINET)
	NCZ_d_Flag.Define(false)
	NCZ_f_Flag.Define("text")
	NCZ_i_Flag.Define("full")
	NCZ_k_Flag.Define("warn")
	NCZ_l_Flag.Define(0)
	NCZ_m_Flag.Define("warn")
	NCZ_n_Flag.Define("warn")
	NCZ_o_Flag.Define("-")
	NCZ_q_Flag.Define(false)
	NCZ_r_Flag.Define("")
	NCZ_s_Flag.Define("full")
	NCZ_t_Flag.Define("")
	NCZ_v_Flag.Define(false)
	NCZ_w_Flag.Define("")

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	args = flag.Args()

	wg := new(sync.WaitGroup)
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		wg.Wait()
	}()

	if NCZ_v_Flag.Value() {
		if mm := xprogram.MainModule(); mm != nil {
			fmt.Println(mm.Version)
		} else {
			fmt.Println("(unavailable)")
		}
		return nil
	}
	if NCZ_q_Flag.Value() {
		verbose = xlog.Mute(verbose)
	}
	switch len(args) {
	case 0:
		return xerrors.Incomplete("zone")
	case 1:
		return xerrors.Incomplete("file")
	}

	zone, fn := args[0], args[1]

	wg.Add(1)
	go xdnsdb.Routine(ctx, wg, verbose)

	if err = xdnsdb.Include(ctx, zone, fn); err != nil {
		return err
	}

	xdnsdb.Dump(os.Stdout)
	return nil
}
