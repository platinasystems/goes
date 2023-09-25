// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package udpecho

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/platinasystems/goes/v2/pkg/context/help"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/resolve"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

const (
	Packets = 1024
	Window  = 4
)

var ErrIncomplete = errors.New("incomplete")

func Ping(
	ctx context.Context,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `{{$path := join .Path " "}}{{/*
*/}}usage: {{$path}} [<options>] <host> [<interface>]
Ping echo host with UDP sequenced packets.

<host>
    [<ip6>]:<port>
    <ip6>
    <ip4>:<port>
    <ip4>
    <name>:<port>
    <name>

The default <port> is {{.Port}}.
`
	if complete.Parameter.Value(ctx) {
		return nil
	}
	if help.Parameter.Value(ctx) {
		return style.Usage(usage, struct {
			Path []string
			Port int
		}{path, Port})
	}
	if len(args) == 0 {
		return ErrIncomplete
	}

	aps, err := resolve.AddrPort(ctx, args[0], Port)
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
				if acked = seq; acked == Packets {
					break
				}
				win = Window - (next - acked) + 1
			} else {
				retx += 1
				next = acked + 1
			}
		}
		for i := 0; i < win && next <= Packets; i++ {
			binary.BigEndian.PutUint64(pg, uint64(next))
			if _, err = c.Write(pg[:8]); err != nil {
				return err
			}
			next += 1
		}
	}
	fmt.Print(retx, " retransmits, ",
		Packets/time.Now().Sub(start).Seconds(), "pps\n")
	return ctx.Err()
}
