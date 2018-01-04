// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netns

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
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

var Goes = &goes.Goes{
	NAME: "netns",
	USAGE: `
	ip netns add NETNSNAME
	ip [-all] netns delete [ NETNSNAME ]
	ip [-all] netns exec [ NETNSNAME ] command...
	ip netns [ list ]
	ip netns list-id
	ip netns identify [ PID ]
	ip netns pids NETNSNAME
	ip netns monitor
	ip netns set NETNSNAME NETNSID`,
	APROPOS: lang.Alt{
		lang.EnUS: "network namespace management",
	},
	MAN: lang.Alt{
		lang.EnUS: Man,
	},
	ByName: map[string]cmd.Cmd{
		"add":      add.Command{},
		"delete":   delete.Command{},
		"exec":     &exec.Command{},
		"identify": identify.Command{},
		"":         list.Command(""),
		"list":     list.Command("list"),
		"list-id":  listid.Command{},
		"mon":      mon.Command{},
		"pids":     pids.Command{},
		"set":      set.Command{},
	},
}
