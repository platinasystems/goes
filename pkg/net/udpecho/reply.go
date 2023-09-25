// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package udpecho

import (
	"context"
	"net"
	"net/netip"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/context/help"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

func Reply(
	ctx context.Context,
	path []string,
	args ...string,
) error {
	const usage = `{{$path := join .Path " "}}{{/*
*/}}usage: {{$path}} [<address>:<port>]
Echo UDP received packets (default <:{{.Port}}>)
`
	if complete.Parameter.Value(ctx) {
		return nil
	}
	if help.Parameter.Value(ctx) {
		if len(path) > 1 && path[1] == "daemon" {
			path[1] = "start"
		}
		return style.Usage(usage, struct {
			Path []string
			Port int
		}{path, Port})
	}
	if !program.IsKoApp() {
		style.System()
	}

	var udpa *net.UDPAddr
	if len(args) == 0 {
		udpa = &net.UDPAddr{Port: 7}
	} else if ap, err := netip.ParseAddrPort(args[0]); err != nil {
		return err
	} else {
		udpa = net.UDPAddrFromAddrPort(ap)
	}

	c, err := net.ListenUDP("udp", udpa)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go routine(ctx, &wg, c)
	<-ctx.Done()
	c.Close()
	wg.Wait()
	return nil
}

func routine(
	ctx context.Context,
	wg *sync.WaitGroup,
	c *net.UDPConn,
) {
	defer wg.Done()
	defer style.Recovery()
	pg := make([]byte, 4<<10)
	for {
		n, from, err := c.ReadFromUDP(pg)
		if err != nil {
			if ctx.Err() != context.Canceled {
				panic(err)
			}
			break
		}
		style.ShortFile.Notice.Println("rx", n, "bytes")
		if _, err = c.WriteTo(pg[:n], from); err != nil {
			panic(err)
		}
	}
}
