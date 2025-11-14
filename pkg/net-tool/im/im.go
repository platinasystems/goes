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
Enter:
  e[xit]
	Or EOF to quit.

  t[ext] host[:port] ...
	Change prompt; then forward the following lines to the named or
	numbered host/port(s).
`

const Port = 8004
const Prompt = "im> "

var im struct {
	sync.Mutex
	clio *clio.CLIO
	port uint16
	udp  *net.UDPConn

	name  map[netip.AddrPort]string
	named map[string]netip.AddrPort
}

func InstantMessaging(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [interface]
Instant Messaging over named or all interface(s).

{{flags .}}`)

	im.port = Port
	err := xflag.Labels{
		xmain.ConfigFlag,
		xflag.Label{"p", "Instant Messaging Port.", &im.port},
	}.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
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

	for err == nil && ctx.Err() == nil {
		var s string
		if s, err = im.clio.ReadLine(); err != nil {
			break
		}
		if args = strings.Fields(s); len(args) == 0 {
			imHelp()
			continue
		}
		switch args[0] {
		case "?", "h", "help":
			imHelp()
		case "e", "exit":
			return nil
		case "t", "text":
			if len(args) < 2 {
				fmt.Fprintln(im.clio, "incomplete")
			} else if err = imText(ctx, args[1:]); err == nil {
			} else if errors.Is(err, xerrors.ErrNotFound) {
				fmt.Fprintln(im.clio, err)
				err = nil
			}
		default:
			fmt.Fprintf(im.clio, "invalid: %q, see help.\n", args[0])
		}
	}
	return xerrors.Suppress(err, io.EOF)
}

func imFrom(ctx context.Context, ap netip.AddrPort) string {
	im.Lock()
	defer im.Unlock()
	if name, ok := im.name[ap]; ok {
		return name
	}
	if names, err := xdnsdoh.LookupName(ctx, ap.Addr()); err == nil {
		return imName(ap, names[0])
	}
	return ap.String()
}

func imHelp() { fmt.Fprint(im.clio, Help[1:]) }

func imName(ap netip.AddrPort, name string) string {
	if i := strings.Index(name, "."); i > 0 {
		name = name[:i]
	}
	im.name[ap] = name
	im.named[name] = ap
	return name
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
		im.clio.Interject(imFrom(ctx, ap), ": ", b[:n])
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
	for err == nil {
		var s string
		if s, err = im.clio.ReadLine(); err != nil || len(s) == 0 {
			break
		}
		for _, ap := range aps {
			_, err = im.udp.WriteToUDPAddrPort([]byte(s), ap)
			if err != nil {
				break
			}
		}
	}
	return err
}

func imTo(ctx context.Context, to []string) (aps []netip.AddrPort, err error) {
	im.Lock()
	defer im.Unlock()
	for _, s := range to {
		var found []netip.AddrPort
		if ap, ok := im.named[s]; ok {
			aps = append(aps, ap)
			continue
		}
		found, err = xdnsdoh.LookupAddrPort(ctx, "udp", s)
		if err != nil {
			return
		}
		ap := found[0]
		if ap.Port() == 0 {
			ap = netip.AddrPortFrom(ap.Addr(), im.port)
		}
		aps = append(aps, ap)
		if n := strings.Count(s, ":"); n > 1 {
			continue
		} else if n == 1 {
			s = s[:strings.Index(s, ":")]
		}
		imName(ap, s)
	}
	return
}
