// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"net/netip"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/context/ctxparm"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/net/resolve"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

const (
	UDPEchoPackets = 1024
	UDPEchoPort    = 7
	UDPEchoWindow  = 4
)

const UDPEchoUsageTemplate = `
usage: {{.Path}} [<address>:<port>]
Echo UDP received packets (default <:{{.Port}}>)`

func UDPEchoUsageData(ctx context.Context) any {
	return struct {
		Path string
		Port int
	}{
		Path: strings.Join(ctxparm.Strings.In(ctx), " "),
		Port: UDPEchoPort,
	}
}

func UDPEcho(ctx context.Context, args ...string) error {
	path := ctxparm.Strings.In(ctx)
	if *complete.Help {
		return nil
	}
	if *usage.Help {
		if len(path) > 1 && path[1] == "daemon" {
			path[1] = "start"
		}
		return usage.Error(UDPEchoUsageTemplate[1:],
			UDPEchoUsageData(ctx))
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

	w := ctxparm.Writer.In(ctx)

	fmt.Fprintln(w, "start", udpa, "service")
	var wg sync.WaitGroup
	wg.Add(1)
	go udpEchoReply(ctx, &wg, c)
	<-ctx.Done()
	c.Close()
	wg.Wait()
	fmt.Fprintln(w, "stopped", udpa, "service")
	return nil
}

const UDPPingUsageTemplate = `
usage: {{.Path}} [<options>] [<host>]
Ping echo host with UDP sequenced packets.

<host>
    [<ip6>]:<port>
    <ip6>
    <ip4>:<port>
    <ip4>
    <name>:<port>
    <name>

The default <host> is 127.0.0.1:{{.Port}}. `

var UDPPingUsageData = UDPEchoUsageData

func UDPPing(ctx context.Context, args ...string) error {
	if *complete.Help {
		return nil
	}
	if *usage.Help {
		return usage.Error(UDPPingUsageTemplate[1:],
			UDPPingUsageData(ctx))
	}
	addr := "127.0.0.1"
	if len(args) > 0 {
		addr = args[0]
	}
	aps, err := resolve.AddrPort(ctx, addr, UDPEchoPort)
	if err != nil {
		return err
	}
	raddr := net.UDPAddrFromAddrPort(aps[0])

	var laddr *net.UDPAddr
	if len(args) > 1 {
		itf, err := net.InterfaceByName(args[1])
		if err != nil {
			return err
		}
		addrs, err := itf.Addrs()
		if err != nil {
			return err
		}
		laddr = &net.UDPAddr{IP: net.ParseIP(addrs[0].String())}
	}

	c, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		return err
	}
	defer c.Close()

	pg := make([]byte, 4<<10)

	if _, err = c.Write(pg[:8]); err != nil {
		return err
	}

	var retx int
	var start time.Time
	for acked, next, tries, win := -1, 1, 0, 1; ctx.Err() == nil; win = 1 {
		c.SetReadDeadline(time.Now().Add(time.Second))
		if n, err := c.Read(pg); err != nil {
			if !errors.Is(err, os.ErrDeadlineExceeded) ||
				tries == 3 {
				return err
			}
			next = acked + 1
			tries += 1
		} else if n != 8 {
			return fmt.Errorf("returned unexpected length: %d", n)
		} else {
			tries = 0
			seq := int(binary.BigEndian.Uint64(pg))
			if seq == 0 {
				start = time.Now()
			}
			if seq <= acked {
				continue
			} else if seq == acked+1 {
				if acked = seq; acked == UDPEchoPackets {
					break
				}
				win = UDPEchoWindow - (next - acked) + 1
			} else {
				retx += 1
				next = acked + 1
			}
		}
		for i := 0; i < win && next <= UDPEchoPackets; i++ {
			binary.BigEndian.PutUint64(pg, uint64(next))
			if _, err = c.Write(pg[:8]); err != nil {
				return err
			}
			next += 1
		}
	}
	fmt.Print(retx, " retransmits, ",
		UDPEchoPackets/time.Now().Sub(start).Seconds(), "pps\n")
	return ctx.Err()
}

func udpEchoReply(
	ctx context.Context,
	wg *sync.WaitGroup,
	c *net.UDPConn,
) {
	defer wg.Done()
	pg := make([]byte, 4<<10)
	for {
		n, from, err := c.ReadFromUDP(pg)
		if err != nil {
			if ctx.Err() != context.Canceled {
				log.Print(err)
			}
			break
		}
		if _, err = c.WriteTo(pg[:n], from); err != nil {
			log.Print(err)
			break
		}
	}
}
