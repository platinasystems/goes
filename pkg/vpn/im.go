// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
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
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

const IMHelp = `
exit	Or EOF to quit.

text <name|addr> ...
	Change prompt; then forward the following lines to the referenced
	subscriber(s) until an empty line.
`

const IMMTU = TunMTU - netph.UDPSize - netph.IP6Size
const IMPrompt = "im> "

var im struct {
	sync.Mutex
	clio *clio.CLIO
	udp  *net.UDPConn

	addressed map[netip.Addr]string
	named     map[string]netip.AddrPort
}

var (
	imPort     = 8004
	imPortFlag = xflag.Label{"imp", "Instant Messaging Port.", &imPort}
)

var imFlags = append(restFlags, imPortFlag)

func InstantMessaging(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] <tunnel-interface>
Instant Messaging over the named interface.

{{flags .}}`)

	err := imFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	} else if args = flag.CommandLine.Args(); len(args) == 0 {
		return xerrors.Incomplete("tunnel-interface")
	}
	tun, err := net.InterfaceByName(args[0])
	if err != nil {
		return err
	}
	tunAddrs, err := tun.Addrs()
	if err != nil {
		return err
	} else if len(tunAddrs) == 0 {
		return fmt.Errorf("%s: has no address", tun.Name)
	}
	tunIPNet, ok := tunAddrs[0].(*net.IPNet)
	if !ok {
		return xerrors.Invalid(tun.Name, "address", tunAddrs[0])
	}
	tunAddr, ok := netip.AddrFromSlice(tunIPNet.IP)
	if !ok {
		return xerrors.Invalid(tun.Name, "IPNet", tunAddrs[0])
	}
	if err = restInit(); err != nil {
		return err
	}
	var zone string
	if tunAddr.Is6() {
		zone = tun.Name
	}
	im.udp, err = net.ListenUDP("udp", &net.UDPAddr{
		IP:   tunIPNet.IP,
		Port: imPort,
		Zone: zone,
	})
	if err != nil {
		return err
	}
	im.clio, err = clio.New(os.Stdin, IMPrompt, nil)
	if err != nil {
		im.udp.Close()
		return err
	}
	defer fmt.Println()
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
	sub, err := Whois(ctx, addr)
	if err == nil {
		name = sub.name()
		im.addressed[sub.Addr] = name
		im.named[sub.name()] = netip.
			AddrPortFrom(sub.Addr, uint16(imPort))
		return name
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
		fmt.Fprint(im.clio, IMHelp[1:])
	case "text":
		if len(args) < 2 {
			fmt.Fprintln(im.clio, "incomplete")
		} else {
			imText(ctx, args[1:])
		}
	default:
		fmt.Fprintf(im.clio, "invalid: %q, see help\n", args[0])
	}
	return ctx.Err()
}

func imRx(ctx context.Context) {
	b := make([]byte, IMMTU)
	for {
		n, ap, err := im.udp.ReadFromUDPAddrPort(b)
		if err != nil {
			return
		}
		im.clio.Interject(imFrom(ctx, ap.Addr()), ": ", b[:n])
	}
}

func imText(ctx context.Context, to []string) {
	aps, err := imTo(ctx, to)
	if err != nil {
		fmt.Fprintln(im.clio, err)
		return
	}
	prompt := to[0]
	if len(to) > 1 {
		prompt += "+"
	}
	im.clio.SetPrompt(fmt.Sprint(prompt, ", "))
	defer im.clio.SetPrompt(IMPrompt)
	for {
		s, err := im.clio.ReadLine()
		if err != nil || len(s) == 0 {
			break
		}
		for _, ap := range aps {
			_, err = im.udp.WriteToUDPAddrPort([]byte(s), ap)
			if err != nil {
				fmt.Fprintln(im.clio, err)
				return
			}
		}
	}
}

func imTo(ctx context.Context, to []string) (aps []netip.AddrPort, err error) {
	var sub *Subscriber
	im.Lock()
	defer im.Unlock()
	for _, s := range to {
		if ap, ok := im.named[s]; ok {
			aps = append(aps, ap)
		} else if sub, err = Whois(ctx, s); err == nil {
			ap := netip.
				AddrPortFrom(sub.Addr, uint16(imPort))
			aps = append(aps, ap)
			im.addressed[sub.Addr] = sub.name()
			im.named[sub.name()] = ap
		} else {
			break
		}
	}
	return
}
