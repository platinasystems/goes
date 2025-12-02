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
Type EOF to quit or enter the following to send a message to the
last origin or the comma separated destinations;
or without argument(s), send subsequent, non-blank entries.

	im> r[eply] MESSAGE...
	im> HOST[:PORT][,...] MESSAGE...
	im> HOST
	HOST> MESSAGE
	HOST> ...
	HOST> <enter>
	im>
`

const Port = 8004
const Prompt = "im> "

var im struct {
	sync.Mutex
	clio *clio.CLIO
	port uint16
	udp  *net.UDPConn

	last struct {
		name string
		ap   netip.AddrPort
	}

	name  map[netip.AddrPort]string
	named map[string]netip.AddrPort
}

var resolver *net.Resolver

func InstantMessaging(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [interface]
Instant Messaging over named or all interface(s).

{{flags .}}`)

	im.port = Port
	err := xflag.Labels{
		xmain.ConfigFlag,
		xdnsdoh.ConfigFlag,
		{"p", "Instant Messaging Port.", &im.port},
	}.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}

	if resolver, err = xdnsdoh.Resolver(); err != nil {
		return err
	}

	lnudp := &net.UDPAddr{Port: int(im.port)}
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
	im.clio.SetPrompt(Prompt)
	defer fmt.Println()

	var wg sync.WaitGroup
	defer wg.Wait()

	defer im.udp.Close()
	defer im.clio.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	im.name = make(map[netip.AddrPort]string)
	im.named = make(map[string]netip.AddrPort)

	wg.Go(func() { imRx(ctx) })

	var aps []netip.AddrPort
	for err == nil && ctx.Err() == nil {
		var s string
		if s, err = im.clio.ReadLine(); err != nil {
		} else if len(s) == 0 {
			if len(aps) == 0 {
				fmt.Fprint(im.clio, Help[1:])
			} else {
				im.clio.SetPrompt(Prompt)
				aps = aps[:0]
			}
		} else if s == "?" || s == "h" || s == "help" {
			fmt.Fprint(im.clio, Help[1:])
		} else if len(aps) > 0 {
			err = imTx(s, aps)
		} else if aps, s, err = imTo(ctx, s); err != nil {
			fmt.Fprintln(im.clio, err)
			err = nil
		} else if len(s) > 0 {
			err = imTx(s, aps)
			aps = aps[:0]
			im.clio.SetPrompt(Prompt)
		}
	}
	return xerrors.Suppress(err, io.EOF)
}

func imName(ctx context.Context, ap netip.AddrPort) string {
	im.Lock()
	defer im.Unlock()
	if name, ok := im.name[ap]; ok {
		return name
	}
	names, err := resolver.LookupAddr(ctx, ap.Addr().String())
	if err != nil {
		return ap.String()
	}
	s := imShortName(names[0])
	im.name[ap] = s
	im.named[s] = ap
	im.last.name = s
	im.last.ap = ap
	return s
}

func imRx(ctx context.Context) {
	b := make([]byte, 4<<10)
	for ctx.Err() == nil {
		n, ap, err := im.udp.ReadFromUDPAddrPort(b)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				im.clio.Interject("rx:", err)
			}
			break
		}
		im.clio.Interject(imName(ctx, ap), ": ", b[:n])
	}
}

func imShortName(name string) string {
	if i := strings.Index(name, "."); i > 0 {
		name = name[:i]
	}
	return name
}

func imTo(ctx context.Context, s string) (
	aps []netip.AddrPort, rem string, err error,
) {
	var prompt strings.Builder
	im.Lock()
	defer im.Unlock()

	if len(s) == 0 {
		err = xerrors.Incomplete("destination")
		return
	}

	for _, prefix := range []string{"r ", "reply "} {
		if strings.HasPrefix(s, prefix) {
			rem = strings.TrimSpace(strings.TrimPrefix(s, prefix))
			if len(im.last.name) > 0 && im.last.ap.IsValid() {
				aps = append(aps, im.last.ap)
				im.clio.SetPrompt(im.last.name + ", ")
			} else {
				err = xerrors.NotFound("conversation")
			}
			return
		}
	}

	if i := strings.IndexAny(s, " \t"); i > 0 {
		rem = strings.TrimSpace(s[i+1:])
		s = s[:i]
	}

	for i, dst := range strings.Split(s, ",") {
		if i > 0 {
			fmt.Fprint(&prompt, ",")
		}
		if ap, ok := im.named[dst]; ok {
			aps = append(aps, ap)
			fmt.Fprint(&prompt, dst)
			continue
		}
		p := im.port
		hs, ps, se := net.SplitHostPort(dst)
		if se != nil {
			hs = dst
		} else {
			hs = xdnsdoh.FQDN(hs)
			pi, pe := net.DefaultResolver.LookupPort(ctx, "udp", ps)
			if pe != nil {
				if _, err = fmt.Sscan(ps, &pi); pe != nil {
					return
				}
			}
			p = uint16(pi)
		}
		var found []netip.Addr
		found, err = resolver.LookupNetIP(ctx, "udp", hs)
		if err != nil {
			return
		}
		ap := netip.AddrPortFrom(found[0], p)
		aps = append(aps, ap)
		hs = imShortName(hs)
		im.name[ap] = hs
		im.named[hs] = ap
		fmt.Fprint(&prompt, hs)
	}
	fmt.Fprint(&prompt, ", ")
	im.clio.SetPrompt(prompt.String())
	return
}

func imTx(s string, aps []netip.AddrPort) error {
	for _, ap := range aps {
		_, err := im.udp.WriteToUDPAddrPort([]byte(s), ap)
		if err != nil {
			return err
		}
	}
	return nil
}
