// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ip

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/cli"
	"github.com/platinasystems/go/goes/cmd/helpers"
	"github.com/platinasystems/go/goes/cmd/ip/address"
	"github.com/platinasystems/go/goes/cmd/ip/batch"
	"github.com/platinasystems/go/goes/cmd/ip/link"
	"github.com/platinasystems/go/goes/cmd/ip/monitor"
	"github.com/platinasystems/go/goes/cmd/ip/n"
	"github.com/platinasystems/go/goes/cmd/ip/neighbor"
	"github.com/platinasystems/go/goes/cmd/ip/netns"
	"github.com/platinasystems/go/goes/cmd/ip/route"
	"github.com/platinasystems/go/goes/cmd/show_packages"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "ip"
	Apropos = "show / manipulate routing, etc."
	Usage   = `
	ip [ -n[etns] NAME | -a[ll] ] [ OPTION... ] OBJECT COMMAND [ ARG ]...
	ip [ -n[ntns] NAME | -a[ll] ] -batch [ -x | -f ] [ - | FILE ]
	
OBJECT := { address | link | monitor | neighbor | netns | route }

OPTION := { -s[tat[isti]cs] | -d[etails] | -r[esolve] |
	-human[-readable] | -iec |
	-f[amily] { inet | inet6 | mpls | bridge | link } |
	-4 | -6 | -B | -0 |
	-l[oops] { maximum-addr-flush-attempts } | -br[ief] |
	-o[neline] | -t[imestamp] | -ts[hort] | -b[atch] [filename] |
	-rc[vbuf] [size] | -c[olor] }`
)

func New() *goes.Goes {
	g := goes.New(Name, Usage,
		lang.Alt{
			lang.EnUS: Apropos,
		},
		lang.Alt{
			lang.EnUS: Man,
		})
	g.Plot(helpers.New()...)
	g.Plot(cli.New()...)
	g.Plot(address.New(),
		batch.New(),
		link.New(),
		monitor.New(),
		n.New("-n"),
		n.New("-netns"),
		neighbor.New(),
		netns.New(),
		route.New(),
		show_packages.New("license"),
		show_packages.New("version"),
	)
	return g
}
