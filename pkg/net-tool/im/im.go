// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package im

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/clio"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsdoh"
)

const Help = `
exit	Or EOF to quit.

text <name|addr> ...
	Change prompt; then forward the following lines to the referenced
	subscriber(s) until an empty line.
`

const Prompt = "im> "

var im struct {
	sync.Mutex
	clio *clio.CLIO
	udp  *net.UDPConn

	addressed map[netip.Addr]string
	named     map[string]netip.AddrPort
}

var port = 8004

func InstantMessaging(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [interface]
Instant Messaging over named or all interface(s).

{{flags .}}`)

	err := xflag.Labels{
		xmain.ConfigFlag,
		xflag.Label{"p", "Instant Messaging Port.", &port},
	}.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}

	lnudp := &net.UDPAddr{Port: port}
	if args = flag.CommandLine.Args(); len(args) > 0 {
		nif, err := net.InterfaceByName(args[0])
		if err != nil {
			return err
		}
		nifAddrs, err := nif.Addrs()
		if err != nil {
			return err
		} else if len(nifAddrs) == 0 {
			return fmt.Errorf("%s: has no address", nif.Name)
		}
		nifIPNet, ok := nifAddrs[0].(*net.IPNet)
		if !ok {
			return xerrors.Invalid(nif.Name, "address", nifAddrs[0])
		}
		lnudp.IP = nifIPNet.IP
		lnudp.Zone = nif.Name
	}

	im.udp, err = net.ListenUDP("udp", lnudp)
	if err != nil {
		return err
	}

	im.clio, err = clio.New(os.Stdin, Prompt, nil)
	if err != nil {
		im.udp.Close()
		return err
	}
	defer fmt.Println()

	var wg sync.WaitGroup
	defer wg.Wait()

	defer im.udp.Close()
	defer im.clio.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	im.addressed = make(map[netip.Addr]string)
	im.named = make(map[string]netip.AddrPort)

	wg.Go(func() { imRx(ctx) })
	for err == nil {
		err = imMain(ctx)
	}

	return xerrors.Suppress(err, io.EOF)
}

func imFrom(ctx context.Context, addr netip.Addr) string {
	im.Lock()
	defer im.Unlock()
	name, ok := im.addressed[addr]
	if ok {
		return name
	}
	names, err := xdnsdoh.LookupName(ctx, addr)
	if err == nil {
		if len(names) > 0 {
			name := names[0]
			im.addressed[addr] = name
			im.named[name] = netip.
				AddrPortFrom(addr, uint16(port))
			return name
		}
	}
	return addr.String()
}

func imMain(ctx context.Context) error {
	s, err := im.clio.ReadLine()
	if err != nil {
		return err
	}
	cmd := "help"
	args := strings.Fields(s)
	if len(args) > 0 {
		cmd = args[0]
	}
	switch cmd {
	case "exit":
		return io.EOF
	case "?", "help":
		fmt.Fprint(im.clio, Help[1:])
	case "text":
		if len(args) < 2 {
			fmt.Fprintln(im.clio, "incomplete")
		} else {
			err = imText(ctx, args[1:])
		}
	default:
		fmt.Fprintf(im.clio, "invalid: %q, see help\n", args[0])
	}
	return ctx.Err()
}

func imRx(ctx context.Context) {
	b := make([]byte, 4<<10)
	for {
		n, ap, err := im.udp.ReadFromUDPAddrPort(b)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				im.clio.Interject("rx:", err)
			}
			break
		}
		im.clio.Interject(imFrom(ctx, ap.Addr()), ": ", b[:n])
	}
}

func imText(ctx context.Context, to []string) error {
	aps, err := imTo(ctx, to)
	if err != nil {
		return err
	}
	prompt := to[0]
	if len(to) > 1 {
		prompt += "+"
	}
	im.clio.SetPrompt(fmt.Sprint(prompt, ", "))
	defer im.clio.SetPrompt(Prompt)
textLoop:
	for {
		s, err := im.clio.ReadLine()
		if err != nil || len(s) == 0 {
			break
		}
		for _, ap := range aps {
			_, err = im.udp.WriteToUDPAddrPort([]byte(s), ap)
			if err != nil {
				break textLoop
			}
		}
	}
	return err
}

func imTo(ctx context.Context, to []string) (aps []netip.AddrPort, err error) {
	im.Lock()
	defer im.Unlock()
	for _, s := range to {
		var addrs []netip.Addr
		if ap, ok := im.named[s]; ok {
			aps = append(aps, ap)
		} else if addrs, err = xdnsdoh.LookupNetIP(ctx, s); err == nil {
			addr := addrs[0]
			ap := netip.AddrPortFrom(addr, uint16(port))
			aps = append(aps, ap)
			im.addressed[addr] = s
			im.named[s] = ap
		} else {
			break
		}
	}
	return
}
