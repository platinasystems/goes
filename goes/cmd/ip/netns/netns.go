// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netns

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/helpers"
	"github.com/platinasystems/go/goes/cmd/ip/netns/exec"
	"github.com/platinasystems/go/goes/cmd/ip/netns/mod"
	"github.com/platinasystems/go/goes/cmd/ip/netns/mon"
	"github.com/platinasystems/go/goes/cmd/ip/netns/show"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "netns"
	Apropos = "network namespace management"
	Usage   = `
	ip [ OPTIONS ] netns  [ COMMAND [ ARGS ]... ]

	ip netns add NETNSNAME
	ip [-all] netns delete [ NETNSNAME ]
	ip netns set NETNSNAME NETNSID
	ip netns [ list ]
	ip netns list-id
	ip netns list-pids NETNSNAME
	ip netns identify [ PID ]
	ip [-all] netns exec [ NETNSNAME ] command...
	ip netns monitor
	`
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
	g.Plot(mod.New("add"),
		mod.New("delete"),
		mod.New("set"),
		show.New("identify"),
		show.New(""),
		show.New("list"),
		show.New("list-ids"),
		show.New("pids"),
		exec.New(),
		mon.New(),
	)
	return g
}
