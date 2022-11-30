// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os/signal"

	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/frame"
	"github.com/platinasystems/goes/v2/pkg/net/tuntap"
	"github.com/platinasystems/goes/v2/pkg/os/page"
	"github.com/platinasystems/goes/v2/pkg/os/termination"
)

func main() {
	var link tuntap.Link
	flag.TextVar(&link, "link", link, "tap link address (default autogen)")
	group := flag.Int("group", tuntap.Unset, "unset w/ -1")
	owner := flag.Int("owner", tuntap.Unset, "unset w/ -1")
	persist := flag.Bool("persist", false, "")
	tap := flag.Bool("tap", false, "(default tun)")
	unit := flag.Uint("unit", 0, "")
	flag.Parse()

	cfg := &tuntap.Configuration{
		Unit:    *unit,
		IsTap:   *tap,
		Persist: *persist,
		Owner:   *owner,
		Group:   *group,
		Link:    link,
	}

	min := tuntap.TunMin
	if *tap {
		min = tuntap.TapMin
	}

	f, err := tuntap.New(cfg)
	if err != nil {
		style.Fatal(err)
	}
	defer f.Close()

	pg := page.New()
	defer page.Free(pg)

	ctx, stop := signal.NotifyContext(context.Background(),
		termination.Signals...)
	defer stop()
	p := poll.With(ctx, f)
	for {
		n, err := p.Read(pg)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				style.Error(err)
			}
			break
		}
		if n < min {
			style.Error("too short")
			break
		}
		if tuntap.HasPI {
			fmt.Println(frame.TunPI(pg[:n]))
		} else if *tap {
			fmt.Println(frame.Eth(pg[:n]))
		} else {
			fmt.Println(frame.IP(pg[:n]))
		}
	}
}
