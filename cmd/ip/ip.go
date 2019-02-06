// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ip

import (
	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/cmd/cli"
	"github.com/platinasystems/goes/cmd/ip/address"
	"github.com/platinasystems/goes/cmd/ip/all"
	"github.com/platinasystems/goes/cmd/ip/batch"
	"github.com/platinasystems/goes/cmd/ip/fou"
	"github.com/platinasystems/goes/cmd/ip/link"
	"github.com/platinasystems/goes/cmd/ip/monitor"
	"github.com/platinasystems/goes/cmd/ip/n"
	"github.com/platinasystems/goes/cmd/ip/neighbor"
	"github.com/platinasystems/goes/cmd/ip/netns"
	"github.com/platinasystems/goes/cmd/ip/route"
	"github.com/platinasystems/goes/lang"
)

var Goes = &goes.Goes{
	NAME: "ip",
	USAGE: `
	ip [ NETNS ] OBJECT [ COMMAND [ FAMILY ] [ OPTIONS ]... [ ARG ]... ]
	ip [ NETNS ] -batch [ -x | -f ] [ - | FILE ]
	
NETNS := { -a[ll] | -n[etns] NAME }

OBJECT := { address | fou | link | monitor | neighbor | netns | route }

FAMILY := { -f[amily] { inet | inet6 | mpls | bridge | link } |
	{ -4 | -6 | -B | -0 } }

OPTION := { -s[tat[isti]cs] | -d[etails] | -r[esolve] |
	-human[-readable] | -iec |
	-l[oops] { maximum-addr-flush-attempts } | -br[ief] |
	-o[neline] | -t[imestamp] | -ts[hort] |
	-rc[vbuf] [size] | -c[olor] }`,
	APROPOS: lang.Alt{
		lang.EnUS: "show / manipulate routing, etc.",
	},
	MAN: lang.Alt{
		lang.EnUS: Man,
	},
	ByName: map[string]cmd.Cmd{
		"-a":       &all.Command{Name: "-a"},
		"-all":     &all.Command{Name: "-all"},
		"-batch":   &batch.Command{},
		"-n":       &n.Command{Name: "-n"},
		"-netns":   &n.Command{Name: "-netns"},
		"address":  address.Goes,
		"cli":      &cli.Command{Prompt: "ip> "},
		"fou":      fou.Goes,
		"link":     link.Goes,
		"netns":    netns.Goes,
		"monitor":  monitor.Command{},
		"neighbor": neighbor.Goes,
		"route":    route.Goes,
	},
}
