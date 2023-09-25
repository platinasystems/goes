// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/platinasystems/goes/v2/pkg/context/help"
	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/frame"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/net/tuntap"
	"github.com/platinasystems/goes/v2/pkg/os/page"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/termination"
)

func main() {
	const usage = `usage: {{.Name}} [<options>]...
Create tun/tap device then print received packets/frames.
{{print .Flags}}`

	defer style.Recovery(context.Canceled)
	ha := netif.NewHardwareAddr()
	unit := flag.CommandLine.Uint("unit", 0, "Unit number suffix.")
	syslog := flag.CommandLine.Bool("syslog", false,
		"log to system instead of stdout")
	var isTap bool
	if tuntap.CanTAP {
		if err := ha.Rand(); err != nil {
			panic(err)
		}
		flag.CommandLine.BoolVar(&isTap, "tap", isTap, "(default tun)")
		flag.CommandLine.TextVar(ha, "link", ha,
			"override random link address")
	}
	var persist bool
	if tuntap.CanPersist {
		flag.CommandLine.BoolVar(&persist, "persist", persist, "")
	}
	owner := tuntap.Unset
	if tuntap.CanChangeOwner {
		flag.CommandLine.IntVar(&owner, "owner", owner, "unset w/ -1")
	}
	group := tuntap.Unset
	if tuntap.CanChangeGroup {
		flag.CommandLine.IntVar(&group, "group", group, "unset w/ -1")
	}
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}
	if help.Parameter.Value(ctx) {
		style.Usage(usage, struct {
			Name  string
			Flags fmt.Formatter
		}{program.Base(), flag.CommandLine})
		return
	}
	f, err := tuntap.New(*unit, isTap, persist, owner, group, ha)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	ctx, stop := signal.NotifyContext(context.Background(),
		termination.Signals...)
	defer stop()

	if *syslog {
		style.System()
	}

	pg := page.New()
	defer page.Free(pg)

	var (
		h   fmt.Formatter
		min int
	)

	if !isTap {
		pi := frame.Header[frame.TunPI](pg)
		h = pi
		ip := (*frame.IPv4)(frame.Data(pi))
		min = frame.Sizeof(pi) + frame.Sizeof(ip)
	} else if tuntap.HasPI {
		pi := frame.Header[frame.TapPI](pg)
		h = pi
		eth := (*frame.ETH)(frame.Data(pi))
		min = frame.Sizeof(pi) + frame.Sizeof(eth)
	} else {
		eth := frame.Header[frame.ETH](pg)
		h = eth
		min = frame.Sizeof(eth)
	}

	p := poll.WithReader(ctx, f)
	for {
		if n, err := p.Read(pg); err != nil {
			panic(err)
		} else if n < min {
			panic(tuntap.ErrUnderrun)
		}
		style.Noteln(h)
	}
}
