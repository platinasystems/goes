// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/context/help"
	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/frame"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/net/tuntap"
	"github.com/platinasystems/goes/v2/pkg/sync/chunk"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

func TunTapper(
	ctx context.Context,
	path []string,
	args ...string,
) error {
	const usage = `{{/*
*/}}usage: {{join .Path " "}} [<option>]... [<addr> <dest> [up]]
Create a tun/tap device then log received packets/frames.
{{print .Flags}}`
	fs, h := flag.New()
	unit := fs.Uint("u", 0, "Unit number suffix.")
	ha := netif.NewHardwareAddr()
	var isTap bool
	if tuntap.CanTAP {
		if err := ha.Rand(); err != nil {
			panic(err)
		}
		fs.BoolVar(&isTap, "tap", isTap, "(default tun)")
		fs.TextVar(&ha, "link", ha, "override random link address")
	}
	var persist bool
	if tuntap.CanPersist {
		fs.BoolVar(&persist, "persist", persist, "")
	}
	owner := tuntap.Unset
	if tuntap.CanChangeOwner {
		fs.IntVar(&owner, "owner", owner, "unset w/ -1")
	}
	group := tuntap.Unset
	if tuntap.CanChangeGroup {
		fs.IntVar(&group, "group", group, "unset w/ -1")
	}
	if complete.Parameter.Value(ctx) {
		style.Completions(args, fs.FlagSet)
		return nil
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	args = fs.Args()
	if help.Parameter.Value(ctx) || *h {
		return style.Usage(usage, struct {
			Path  []string
			Flags fmt.Formatter
		}{path, fs})
	}

	f, err := tuntap.New(*unit, isTap, persist, owner, group, ha)
	if err != nil {
		return err
	}
	defer f.Close()

	if len(args) > 1 {
		nif, err := netif.ByName(f.Name())
		if err != nil {
			return err
		}
		local, err := netip.ParseAddr(args[0])
		if err != nil {
			return fmt.Errorf("%q %w", args[0], err)
		}
		remote, err := netip.ParseAddr(args[1])
		if err != nil {
			return fmt.Errorf("%q %w", args[1], err)
		}
		bits := 32
		if local.Is6() {
			bits = 128
		}
		err = nif.Add(local, remote, bits, args[2:])
		if err != nil {
			return err
		}
	}

	buf := chunk.NewJumbo()
	defer chunk.FreeJumbo(buf)

	var (
		hdr fmt.Formatter
		min int
	)

	if !isTap {
		pi := frame.Header[frame.TunPI](buf)
		hdr = pi
		ip := (*frame.IPv4)(frame.Data(pi))
		min = frame.Sizeof(pi) + frame.Sizeof(ip)
	} else if tuntap.HasPI {
		pi := frame.Header[frame.TapPI](buf)
		hdr = pi
		eth := (*frame.ETH)(frame.Data(pi))
		min = frame.Sizeof(pi) + frame.Sizeof(eth)
	} else {
		eth := frame.Header[frame.ETH](buf)
		hdr = eth
		min = frame.Sizeof(eth)
	}

	p := poll.WithReader(ctx, f)
	for {
		if n, err := p.Read(buf); err != nil {
			panic(err)
		} else if n < min {
			panic(tuntap.ErrUnderrun)
		}
		style.Noteln(hdr)
	}
	return nil
}
