// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"flag"
	"fmt"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binpdu"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/net/tuntap"
	"github.com/platinasystems/goes/v2/pkg/sync/chunk"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

var HdrDump = func(...any) {}

const TunTapperUsage = `
usage: {{branch .}} [<option>]... [<addr> <dest> [up]]
Create a tun/tap device then log received packets/frames.
{{flags .}}`

func TunTapper(ctx context.Context, args []string) error {
	var flags flag.FlagSet
	ctx = goes.FlagsContext(ctx, &flags)
	unit := flags.Uint("u", 0, "Unit number suffix.")
	ha := netif.NewHardwareAddr()
	var isTap bool
	if tuntap.CanTAP {
		if err := ha.Rand(); err != nil {
			return egress.Mark(err)
		}
		flags.BoolVar(&isTap, "tap", isTap, "(default tun)")
		flags.TextVar(&ha, "link", ha, "override random link address")
	}
	var persist bool
	if tuntap.CanPersist {
		flags.BoolVar(&persist, "persist", persist, "")
	}
	owner := tuntap.Unset
	if tuntap.CanChangeOwner {
		flags.IntVar(&owner, "owner", owner, "unset w/ -1")
	}
	group := tuntap.Unset
	if tuntap.CanChangeGroup {
		flags.IntVar(&group, "group", group, "unset w/ -1")
	}
	if goes.ContextComplete(ctx) {
		return complete.Last(args, flags)
	}
	ctx, err := goes.ParseFlagsContext(ctx, args)
	if err != nil {
		return err
	}
	if goes.ContextHelp(ctx) {
		return goes.Usage(ctx, TunTapperUsage)
	}
	args = flags.Args()

	f, err := tuntap.New(*unit, isTap, persist, owner, group, ha)
	if err != nil {
		return egress.Mark(err)
	}
	defer f.Close()

	if len(args) > 1 {
		nif := netif.Named(f.Name())
		if nif == nil {
			return egress.Markf("%q %w", f.Name(), ErrNotFound)
		}
		local, err := netip.ParseAddr(args[0])
		if err != nil {
			return egress.Markf("%q %w", args[0], err)
		}
		remote, err := netip.ParseAddr(args[1])
		if err != nil {
			return egress.Markf("%q %w", args[1], err)
		}
		var prefix netip.Prefix
		if local.Is4() {
			prefix = netip.PrefixFrom(local, 32)
		} else {
			prefix = netip.PrefixFrom(local, 128)
		}
		err = nif.Add(ctx, prefix, remote, args[2:]...)
		if err != nil {
			return egress.Mark(err)
		}
	}

	buf := chunk.NewJumbo()
	defer chunk.FreeJumbo(buf)

	var (
		hdr fmt.Formatter
		min int
	)

	const (
		sizeof_pi  = 4
		sizeof_ip  = 20
		sizeof_eth = 14
	)
	if !isTap {
		hdr = binpdu.TunPI(buf)
		min = sizeof_pi + sizeof_ip
	} else if tuntap.HasPI {
		hdr = binpdu.TunPI(buf)
		min = sizeof_pi + sizeof_eth
	} else {
		hdr = binpdu.Eth(buf)
		min = sizeof_eth
	}

	p := poll.WithReader(ctx, f)
	for {
		if n, err := p.Read(buf); err != nil {
			return egress.Mark(err)
		} else if n < min {
			return egress.Mark(tuntap.ErrUnderrun)
		}
		HdrDump(hdr)
	}
	return nil
}
