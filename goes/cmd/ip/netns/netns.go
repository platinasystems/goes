// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netns

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/helpers"
	"github.com/platinasystems/go/goes/cmd/ip/netns/add"
	"github.com/platinasystems/go/goes/cmd/ip/netns/delete"
	"github.com/platinasystems/go/goes/cmd/ip/netns/exec"
	"github.com/platinasystems/go/goes/cmd/ip/netns/identify"
	"github.com/platinasystems/go/goes/cmd/ip/netns/list"
	"github.com/platinasystems/go/goes/cmd/ip/netns/listid"
	"github.com/platinasystems/go/goes/cmd/ip/netns/mon"
	"github.com/platinasystems/go/goes/cmd/ip/netns/pids"
	"github.com/platinasystems/go/goes/cmd/ip/netns/set"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "netns"
	Apropos = "network namespace management"
	Usage   = `
	ip netns add NETNSNAME
	ip [-all] netns delete [ NETNSNAME ]
	ip [-all] netns exec [ NETNSNAME ] command...
	ip netns [ list ]
	ip netns list-id
	ip netns identify [ PID ]
	ip netns pids NETNSNAME
	ip netns monitor
	ip netns set NETNSNAME NETNSID`
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
	g.Plot(
		add.New(),
		delete.New(),
		exec.New(),
		identify.New(),
		list.New(""),
		list.New("list"),
		listid.New(),
		mon.New(),
		pids.New(),
		set.New(),
	)
	return g
}
