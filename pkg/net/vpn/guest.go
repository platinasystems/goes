// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/box"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binpdu"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binph"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/net/tuntap"
	"golang.org/x/exp/maps"
)

type guest struct {
	client
}

func (g *guest) daemon(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} [<options>] ` + vpnRegistrySyntax + `
Start VPN tunnel.
{{flags .}}`
	var wg sync.WaitGroup

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

	if err = g.register(ctx, args[0]); err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(ctx)

	iguest := IdIndex(g.id)
	verbose.Printf("start (%d, %v)", iguest, svc)
	defer verbose.Printf("stopped (%d, %v) %v", iguest, svc, err)
	defer wg.Wait()
	defer cancel()
	defer verbose.Printf("stopping (%d, %v)", iguest, svc)

	udp, err := egress.MarkResult(net.ListenUDP("udp", &net.UDPAddr{
		IP:   svc.Addr().AsSlice(),
		Port: int(svc.Port()),
		// Zone: FIXME,
	}))
	if err != nil {
		return err
	}

	defer udp.Close()

	via := g.via[iguest]

	viaBlk, err := httpWhoIsIdentified(cctx, g.registry, via)
	if err != nil {
		return err
	} else if err = g.peer(viaBlk); err != nil {
		return err
	}

	dst, err := egress.MarkResult(addressHeader(viaBlk))
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
	tun, err := egress.MarkResult(tuntap.
		New(*uFlag, istap, persist, owner, group, ha))
	if err != nil {
		return err
	}
	defer tun.Close()

	nif := netif.Named(tun.Name())
	if nif == nil {
		return egress.Markf("%s: %w", tun.Name(), ErrNotFound)
	}

	err = egress.Mark(nif.Add(cctx, g.hostPrefix, dst,
		"up", "mtu", fmt.Sprintf("%d", box.ContentMTU)))
	if err != nil {
		return err
	}

	err = egress.Mark(routeAdd(cctx, g.vpnPrefix, dst),
		"route add", g.vpnPrefix, "via", dst, "through", nif.Name)
	if err != nil {
		return err
	}
	defer routeDelete(cctx, g.vpnPrefix, dst)

	pktRxCh := make(chan *box.Box, 4)
	pktTxCh := make(chan *box.Box, 4)
	tunReadCh := make(chan *box.Box, 4)
	tunWriteCh := make(chan *box.Box, 4)

	maps.Copy(binpdu.TunPIprotos, TunPIprotos)

	wg.Add(1)
	go pktRxRoutine(cctx, &wg, udp, pktRxCh)
	wg.Add(1)
	go pktTxRoutine(cctx, &wg, udp, pktTxCh)
	defer close(pktTxCh)
	wg.Add(1)
	go tunReadRoutine(cctx, &wg, tun.Name(), tun, tunReadCh)
	wg.Add(1)
	go tunWriteRoutine(cctx, &wg, tun.Name(), tun, tunWriteCh)
	defer close(tunWriteCh)

	if err = g.hello(pktTxCh, via, time.Now()); err != nil {
		return err
	}

	kat := time.NewTicker(30 * time.Second)
	defer kat.Stop()

	verbose.Printf("start (%s, %v)", nif.Name, g.hostPrefix)
	defer verbose.Printf("stopped (%s, %v)", nif.Name, g.hostPrefix)

guestLoop:
	for {
		select {
		case <-cctx.Done():
			verbose.Println("done")
			break guestLoop
		case t := <-kat.C:
			if x, ok := g.via[iguest]; ok && x != via {
				via = x
			}
			if err = g.hello(pktTxCh, via, t); err != nil {
				errata.Println(err)
			}
		case bx, ok := <-tunReadCh:
			if !ok {
				verbose.Println("tun read ch closed")
				break guestLoop
			}
			ato := g.toWhom(bx)
			if !ato.IsValid() {
				bx.Return()
				continue guestLoop
			}
			to, ok := g.addressed[ato]
			if !ok {
				g.whoisAddressed(pktTxCh, ato)
				bx.Return()
				continue guestLoop
			}
			ito := IdIndex(to)
			cto, ok := g.gcm[ito]
			if !ok {
				verbose.Printf("no guest cipher %d, %v",
					ito, ato)
				bx.Return()
				continue guestLoop
			}
			via, ok := g.via[ito]
			if !ok {
				errata.Println("no guest exchange %d, %v",
					ito, ato)
				bx.Return()
				continue guestLoop
			}
			ivia := IdIndex(via)
			cvia, ok := g.gcm[ivia]
			if !ok {
				g.whoisId(pktTxCh, via)
				bx.Return()
				continue guestLoop
			}
			ap, ok := g.service[ivia]
			if !ok {
				errata.Println("no exchange service %d", ivia)
				bx.Return()
				continue guestLoop
			}
			bx.AddrPort = ap
			bx.From(g.id)
			bx.Via(via)
			bx.To(to)
			bx.CloseWith(cto)
			bx.SealWith(cvia)
			bx.NonBlockingPut(pktTxCh)
		case bx, ok := <-pktRxCh:
			if !ok {
				verbose.Println("pkt tx ch closed")
				break guestLoop
			}
			avia := bx.AddrPort
			ex, ok := g.via[iguest]
			ivia := IdIndex(ex)
			if !ok {
				errata.Printf("no cipher for exchange %d", ivia)
				bx.Return()
				continue guestLoop
			}
			cex, ok := g.gcm[ivia]
			if !ok {
				errata.Printf("no cipher for exchange %d", ivia)
				bx.Return()
				continue guestLoop
			}
			if svc := g.service[ivia]; avia != svc {
				verbose.Println(avia, "!=", svc)
				bx.Return()
				continue guestLoop
			}
			from := bx.FromWhom()
			ifrom, vfrom := IdIndex(from), IdVersion(from)
			ver, ok := g.ver[ifrom]
			if !ok || ver != vfrom {
				g.whoisId(pktTxCh, from)
				bx.Return()
				continue guestLoop
			}
			cfrom, ok := g.gcm[ifrom]
			if !ok {
				errata.Printf("no cipher for guest %d", ifrom)
				bx.Return()
				continue guestLoop
			}
			if err = bx.UnsealWith(cex); err != nil {
				verbose.Println("unseal:", err)
				bx.Return()
				continue guestLoop
			}
			to := bx.ToWhom()
			if to != g.id {
				err = g.hello(pktTxCh, from, time.Now())
				if err != nil {
					errata.Println(err)
				}
				bx.Return()
				continue guestLoop
			}
			if err = bx.OpenWith(cfrom); err != nil {
				verbose.Print(err)
				bx.Return()
				continue guestLoop
			}
			var pi binph.TunPI
			if _, err = pi.ReadFrom(bx); err != nil {
				verbose.Print(ErrUnderrun)
				bx.Return()
				continue guestLoop
			}
			switch pi.Proto {
			case VPN_P_HELLO:
				verbose.Print(FIXME)
				// re-checkin if registry era mismatch
				// else re-query exchange from registry
				// if that era is mismatched
			case VPN_P_PUBLIC_KEY:
				blk, _ := pem.Decode(bx.Contents)
				if blk == nil {
					errata.Println(ErrNotPEM)
				} else if err = g.peer(blk); err != nil {
					verbose.Println(err)
				}
				bx.Return()
			case binpdu.ETH_P_IP, binpdu.ETH_P_IPV6:
				bx.Rewind()
				tunWriteCh <- bx
			default:
				verbose.Printf("unknown %#x\n", pi.Proto)
				bx.Return()
			}
		}
	}
	return err
}

func (*guest) toWhom(bx *box.Box) netip.Addr {
	var pi binph.TunPI
	defer bx.Rewind()
	if _, err := pi.ReadFrom(bx); err != nil {
		return zaddr
	}
	switch pi.Proto {
	case binpdu.ETH_P_IP:
		var ip binph.IP
		if _, err := ip.ReadFrom(bx); err == nil {
			return netip.AddrFrom4([4]byte(ip.DA))
		}
	case binpdu.ETH_P_IPV6:
		var ip6 binph.IP6
		if _, err := ip6.ReadFrom(bx); err == nil {
			return netip.AddrFrom16([16]byte(ip6.DA))
		}
	}
	return zaddr
}

func (g *guest) whoisAddressed(ch chan<- *box.Box, addr netip.Addr) {
	bx := box.New()
	binph.TunPI{
		Proto: VPN_P_WHOIS_ADDRESSED,
	}.WriteTo(bx)
	WriteAddrTo(bx, addr)
	g.whois(ch, bx)
}

func (g *guest) whoisId(ch chan<- *box.Box, id box.Id) {
	bx := box.New()
	binph.TunPI{
		Proto: VPN_P_WHOIS_IDENTIFIED,
	}.WriteTo(bx)
	binint.BigEndianValue(id).WriteTo(bx)
	g.whois(ch, bx)
}

func (g *guest) whois(ch chan<- *box.Box, bx *box.Box) {
	via, ok := g.via[IdIndex(g.id)]
	if !ok {
		errata.Print("no assigned exchange")
		bx.Return()
		return
	}
	ivia := IdIndex(via)
	cvia, ok := g.gcm[ivia]
	if !ok {
		errata.Print("no shared cipher")
		bx.Return()
		return
	}
	svc, ok := g.service[ivia]
	if !ok {
		errata.Print("no assigned service")
		bx.Return()
		return
	}
	bx.AddrPort = svc
	bx.From(g.id)
	bx.Via(via)
	bx.To(via)
	bx.CloseWith(cvia)
	bx.SealWith(cvia)
	bx.NonBlockingPut(ch)
}

func tunReadRoutine(
	ctx context.Context,
	wg *sync.WaitGroup,
	name string,
	r io.Reader,
	ch chan<- *box.Box,
) {
	defer wg.Done()
	verbose.Println("start", name, "read routine")
	defer verbose.Println("stopped", name, "read routine")
	for ctx.Err() == nil {
		bx, err := box.NewReadContents(r)
		if err != nil {
			if !errors.Is(err, os.ErrClosed) {
				verbose.Print(err)
			}
			return
		}
		verbose.Print(binpdu.TunPI(bx.Contents))
		ch <- bx
	}
}

func tunWriteRoutine(
	ctx context.Context,
	wg *sync.WaitGroup,
	name string,
	w io.Writer,
	ch <-chan *box.Box,
) {
	defer wg.Done()
	verbose.Println("start", name, "write routine")
	defer verbose.Println("stopped", name, "write routine")
	for {
		select {
		case <-ctx.Done():
			verbose.Println("done")
			return
		case bx, ok := <-ch:
			if !ok {
				verbose.Println("tun write ch closed")
				return
			}
			if len(bx.Contents) == 0 {
				verbose.Print("no content")
			} else if _, err := bx.WriteTo(w); err != nil {
				verbose.Print(err)
			} else {
				verbose.Print(binpdu.TunPI(bx.Contents))
			}
			bx.Return()
		}
	}
}
