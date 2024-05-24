// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/box/label"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/endian"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/net/frame"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/net/tuntap"
)

type guest struct {
	client
	dst    netip.Addr
	tunnel *os.File
	nif    *netif.NetIf
	ch     struct {
		pkt struct {
			rx, tx chan *crate
		}
		tun struct {
			read, write chan *crate
		}
	}
	wg sync.WaitGroup
}

func (g *guest) daemon(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} [<options>] ` + vpnRegistrySyntax + `
Start VPN tunnel.
{{flags .}}`

	flags := goes.ContextFlags(ctx)
	uFlag := flags.Uint("u", 0, "Unit number.")

	if goes.ContextComplete(ctx) {
		return nil
	}
	svc, err := vpnDaemonFlags(ctx, usage, args)
	if err != nil {
		return err
	} else if args = flags.Args(); len(args) < 1 {
		return ErrIncomplete
	}

	defer g.wg.Wait()

	if err = g.register(ctx, args[0]); err != nil {
		return err
	}

	udp, err := egress.MarkResult(net.ListenUDP("udp", &net.UDPAddr{
		IP:   svc.Addr().AsSlice(),
		Port: int(svc.Port()),
		// Zone: FIXME,
	}))
	if err != nil {
		return err
	}

	defer udp.Close()

	verbose.Printf("start (%d, %v)", g.label, svc)
	defer verbose.Printf("stopped (%d, %v)", g.label, svc)

	via := g.via[g.label.Index()]

	viaBlk, err := httpWhoIsLabelled(ctx, g.registry, via)
	if err != nil {
		return err
	} else if err = g.peer(viaBlk); err != nil {
		return err
	}

	g.dst, err = egress.MarkResult(addressHeader(viaBlk))
	if err != nil {
		return err
	}

	ha := netif.NewHardwareAddr()
	if err = egress.Mark(ha.Rand()); err != nil {
		return err
	}
	const (
		istap   = false
		persist = false
		owner   = -1
		group   = -1
	)
	g.tunnel, err = egress.MarkResult(tuntap.
		New(*uFlag, istap, persist, owner, group, ha))
	if err != nil {
		return err
	}
	defer g.tunnel.Close()
	if g.nif = netif.Named(g.tunnel.Name()); g.nif == nil {
		return egress.Markf("%s: %w", g.tunnel.Name(), ErrNotFound)
	}
	err = egress.Mark(g.nif.Add(ctx, g.hostPrefix, g.dst, "up"))
	if err != nil {
		return err
	}
	err = egress.Mark(routeAdd(ctx, g.vpnPrefix, g.dst),
		"route add", g.vpnPrefix, "via", g.dst, "through", g.nif.Name)
	if err != nil {
		return err
	}
	defer routeDelete(ctx, g.vpnPrefix, g.dst)

	g.ch.pkt.rx = make(chan *crate, 4)
	g.ch.pkt.tx = make(chan *crate, 4)
	g.ch.tun.read = make(chan *crate, 4)
	g.ch.tun.write = make(chan *crate, 4)

	g.wg.Add(1)
	go pktRxRoutine(ctx, &g.wg, udp, g.ch.pkt.rx)
	g.wg.Add(1)
	go pktTxRoutine(ctx, &g.wg, udp, g.ch.pkt.tx)
	g.wg.Add(1)
	go g.tunReadRoutine(ctx)
	g.wg.Add(1)
	go g.tunWriteRoutine(ctx)

	g.hello(via, time.Now())

	kat := time.NewTicker(30 * time.Second)
	defer kat.Stop()

	verbose.Printf("start (%s, %v)", g.nif.Name, g.hostPrefix)
	defer verbose.Printf("stopped (%s, %v)", g.nif.Name, g.hostPrefix)

guestLoop:
	for {
		select {
		case <-ctx.Done():
			return nil
		case t := <-kat.C:
			if x, ok := g.via[g.label.Index()]; ok && x != via {
				via = x
			}
			g.hello(via, t)
		case c := <-g.ch.tun.read:
			ato := g.toWhom(c.box.Contents())
			if !ato.IsValid() {
				inventory.Put(c)
				continue guestLoop
			}
			to, ok := g.addressed[ato]
			if !ok {
				g.whoisAddressed(ato)
				inventory.Put(c)
				continue guestLoop
			}
			ito := to.Index()
			togcm, ok := g.gcm[ito]
			if !ok {
				verbose.Printf("no guest cipher %d, %v",
					ito, ato)
				inventory.Put(c)
				continue guestLoop
			}
			via, ok := g.via[ito]
			if !ok {
				errata.Println("no guest exchange %d, %v",
					ito, ato)
				inventory.Put(c)
				continue guestLoop
			}
			ivia := via.Index()
			viagcm, ok := g.gcm[ivia]
			if !ok {
				g.whoisLabelled(via)
				inventory.Put(c)
				continue guestLoop
			}
			c.ap, ok = g.service[ivia]
			if !ok {
				errata.Println("no exchange service %d", ivia)
				inventory.Put(c)
				continue guestLoop
			}
			c.box = c.box.From(g.label)
			c.box = c.box.Via(via)
			c.box = c.box.To(to)
			c.box = c.box.CloseWith(togcm)
			c.box.SealWith(viagcm)
			c.Put(g.ch.pkt.tx)
		case c := <-g.ch.pkt.rx:
			ex, ok := g.via[g.label.Index()]
			ivia := ex.Index()
			if !ok {
				errata.Printf("no cipher for exchange %d", ivia)
				inventory.Put(c)
				continue guestLoop
			}
			exc, ok := g.gcm[ivia]
			if !ok {
				errata.Printf("no cipher for exchange %d", ivia)
				inventory.Put(c)
				continue guestLoop
			}
			if svc := g.service[ivia]; c.ap != svc {
				verbose.Println(c.ap, "!=", svc)
				inventory.Put(c)
				continue guestLoop
			}
			from := c.box.FromWhom()
			ifrom, fromv := from.Index(), from.Version()
			ver, ok := g.ver[ifrom]
			if !ok || ver != fromv {
				g.whoisLabelled(from)
				inventory.Put(c)
				continue guestLoop
			}
			fromc, ok := g.gcm[ifrom]
			if !ok {
				errata.Printf("no cipher for guest %d", ifrom)
				inventory.Put(c)
				continue guestLoop
			}
			if err = c.box.UnsealWith(exc); err != nil {
				verbose.Println("unseal:", err)
				inventory.Put(c)
				continue guestLoop
			}
			to := c.box.ToWhom()
			if to != g.label {
				g.hello(from, time.Now())
				inventory.Put(c)
				continue guestLoop
			}
			c.box, err = c.box.OpenWith(fromc)
			if err != nil {
				verbose.Print(err)
				inventory.Put(c)
				continue guestLoop
			}
			pi := NewTunPI(c.box.Contents())
			switch pi.Proto {
			case TUNPI_P_VPN_HELLO:
				verbose.Print(FIXME)
				// re-checkin if registry era mismatch
				// else re-query exchange from registry
				// if that era is mismatched
			case TUNPI_P_VPN_WHOIS:
				if pi.Flags != VPN_WHOIS_F_RESPONSE {
					errata.Printf("unknown %#x", pi.Flags)
				} else if blk, _ := pem.
					Decode(pi.Data); blk == nil {
					errata.Print(ErrNotPEM)
				} else if err = g.peer(blk); err != nil {
					fmt.Fprint(verbose, err)
				}
				inventory.Put(c)
			case TUNPI_P_IP, TUNPI_P_IPV6:
				g.ch.tun.write <- c
			default:
				verbose.Printf("unknown %#x", pi.Proto)
				inventory.Put(c)
			}
		}
	}
}

/*FIXME
func (*guest) getLnIP(want4 bool) (net.IP, error) {
	ifas, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, ifa := range ifas {
		prefix, err := netip.ParsePrefix(ifa.String())
		if err != nil {
			continue
		}
		addr := prefix.Addr()
		if addr.IsLoopback() {
			continue
		}
		if (want4 && addr.Is4()) || (!want4 && addr.Is6()) {
			return net.IP(addr.AsSlice()), nil
		}
	}
	return nil, ErrNoServiceIP
}
*/

func (g *guest) hello(to label.Label, now time.Time) {
	ifrom := g.label.Index()
	ito := to.Index()
	c := newCrate()
	c.box = c.box.Empty()
	c.box = TunPI{
		Flags: VPN_HELLO_F_UNIX_MICRO,
		Proto: TUNPI_P_VPN_HELLO,
		Data:  endian.NewBigInteger(now.UnixMicro()),
	}.Append(c.box)
	c.box = c.box.From(g.label)
	c.box = c.box.To(to)
	via, ok := g.via[ito]
	if !ok {
		via = to
	}
	ivia := via.Index()
	c.box = c.box.Via(via)
	c.ap, ok = g.service[ivia]
	if !ok {
		errata.Println("no exchange service for", ivia)
		inventory.Put(c)
		return
	}
	gcm, ok := g.gcm[ivia]
	if !ok {
		errata.Println("no exchange cipher for", ivia)
		inventory.Put(c)
		return
	}
	c.box = c.box.CloseWith(gcm)
	c.box.SealWith(gcm)
	verbose.Println("hello", ito, "from", ifrom, "via", ivia)
	g.ch.pkt.tx <- c
}

func (*guest) toWhom(contents []byte) netip.Addr {
	pi := frame.Header[frame.TunPI](contents)
	switch pi.Proto.Value() {
	case frame.PI_P_IP:
		return netip.AddrFrom4((*frame.IPv4)(frame.Data(pi)).DA)
	case frame.PI_P_IPV6:
		return netip.AddrFrom16((*frame.IPv6)(frame.Data(pi)).DA)
	}
	return netip.Addr{}
}

func (g *guest) tunReadRoutine(ctx context.Context) {
	defer g.wg.Done()
	tname := g.tunnel.Name()
	verbose.Println("start", tname, "read routine")
	defer verbose.Println("stopped", tname, "read routine")
	for ctx.Err() == nil {
		c := newCrate()
		n, err := g.tunnel.Read(c.box.Contents())
		if err != nil {
			if !errors.Is(err, os.ErrClosed) {
				verbose.Print(err)
			}
			return
		}
		c.box = c.box.Shrink(n)
		// FIXME rewrite frame package to use binary/endian
		verbose.Print(frame.Header[frame.TunPI](c.box.Contents()))
		g.ch.tun.read <- c
	}
}

func (g *guest) tunWriteRoutine(ctx context.Context) {
	defer g.wg.Done()
	tname := g.tunnel.Name()
	verbose.Println("start", tname, "write routine")
	defer verbose.Println("stopped", tname, "write routine")
	for {
		select {
		case <-ctx.Done():
			return
		case c := <-g.ch.tun.write:
			data := c.box.Contents()
			if len(data) == 0 {
				verbose.Print("no content")
			} else if _, err := g.tunnel.Write(data); err != nil {
				verbose.Print(err)
			} else {
				verbose.Print(frame.Header[frame.TunPI](data))
			}
			inventory.Put(c)
		}
	}
}

func (g *guest) whoisAddressed(addr netip.Addr) {
	a16 := addr.As16()
	c := newCrate()
	c.box = c.box.Empty()
	c.box = TunPI{
		Flags: VPN_WHOIS_F_ADDRESSED,
		Proto: TUNPI_P_VPN_WHOIS,
		Data:  a16[:],
	}.Append(c.box)
	g.whois(c)
}

func (g *guest) whoisLabelled(lbl label.Label) {
	c := newCrate()
	c.box = c.box.Empty()
	c.box = TunPI{
		Flags: VPN_WHOIS_F_LABELLED,
		Proto: TUNPI_P_VPN_WHOIS,
		Data:  endian.NewBigInteger(lbl),
	}.Append(c.box)
	g.whois(c)
}

func (g *guest) whois(c *crate) {
	verbose.Print(string(c.box.Contents()))
	via, ok := g.via[g.label.Index()]
	if !ok {
		errata.Print("no assigned exchange")
		inventory.Put(c)
	}
	ivia := via.Index()
	gcm, ok := g.gcm[ivia]
	if !ok {
		errata.Print("no shared cipher")
		inventory.Put(c)
	}
	if c.ap, ok = g.service[ivia]; !ok {
		errata.Print("no assigned service")
		inventory.Put(c)
	}
	c.box = c.box.From(g.label)
	c.box = c.box.Via(via)
	c.box = c.box.To(via)
	c.box = c.box.CloseWith(gcm)
	c.box.SealWith(gcm)
	c.Put(g.ch.pkt.tx)
}
