// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"maps"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/box"
	"github.com/platinasystems/goes/v2/pkg/netif"
	"github.com/platinasystems/goes/v2/pkg/nettun"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netpdu"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

// Guest is a UDP server that forwards ciphered packets between an exchange
// and a network tunnel interface.
func Guest(ctx context.Context, args []string) error {
	const defport = 0
	var wg sync.WaitGroup
	var g guest

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] [vpn]
Forward ciphered packets between exchange and tunnel interface.

{{flags .}}`)

	tflag := TunnelFlag()

	g.xFlag = GuestFlag()
	err := g.flags(ctx, defport, args)
	if err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(ctx)

	udp, err := net.ListenUDP(g.udpv, &net.UDPAddr{
		IP:   g.sap.Addr().AsSlice(),
		Port: int(g.sap.Port()),
	})
	if err != nil {
		return xerrors.Label(err, "ListenUDP")
	}

	defer udp.Close()

	lap, err := netip.ParseAddrPort(udp.LocalAddr().String())
	if err != nil {
		return xerrors.Label(err, "LocalAddr")
	}
	if err = g.register(ctx, ap0); err != nil {
		return err
	}

	iguest := IdIndex(g.id)
	via := g.via[iguest]

	svc := fmt.Sprintf("(%d via %d @ %v)", iguest, via, lap)
	goRoutineTrace.Println("start", svc)
	defer goRoutineTrace.Println("stopped", svc)
	defer wg.Wait()
	defer cancel()
	defer goRoutineTrace.Println("stopping", svc, "...")

	viaBlk, err := g.whoisIdentified(cctx, via)
	if err != nil {
		return err
	} else if err = g.peer(viaBlk); err != nil {
		return err
	}

	dst, err := addressHeader(viaBlk)
	if err != nil {
		return xerrors.Label(err, "address")
	}

	ha := netif.NewHardwareAddr()
	if err = ha.Rand(); err != nil {
		return xerrors.Label(err, "rand_")
	}

	const (
		istap   = false
		persist = false
		owner   = -1
		group   = -1
	)

	tun, err := nettun.New(*tflag, istap, persist, owner, group, ha)
	if err != nil {
		return err
	}
	defer tun.Close()

	nif := netif.Named(tun.Name())
	if nif == nil {
		return xerrors.NotFound(tun.Name())
	}

	mtu := fmt.Sprintf("%d", box.ContentMTU)
	err = nif.Add(cctx, g.hostPrefix, dst, "up", "mtu", mtu)
	if err != nil {
		return xerrors.Label(err, "ifconfig", nif.Name,
			g.hostPrefix.String(), "mtu", mtu)
	}

	err = routeAdd(cctx, g.vpnPrefix, dst)
	if err != nil {
		return xerrors.Label(err, "route", "add", g.vpnPrefix.String())
	}
	defer routeDelete(cctx, g.vpnPrefix, dst)

	pktRxCh := make(chan *box.Box, 4)
	pktTxCh := make(chan *box.Box, 4)
	tunReadCh := make(chan *box.Box, 4)
	tunWriteCh := make(chan *box.Box, 4)

	maps.Copy(netpdu.TunPIprotos, TunPIprotos)

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

	goRoutineTrace.Printf("start (%s, %v)", nif.Name, g.hostPrefix)
	defer goRoutineTrace.Printf("stopped (%s, %v)", nif.Name, g.hostPrefix)

	var tunpi netph.TunPI
	var data []byte

guestLoop:
	for {
		select {
		case <-cctx.Done():
			verbose.Println("done")
			break guestLoop
		case t := <-kat.C:
			if err = g.hello(pktTxCh, g.via[iguest], t); err != nil {
				errata.Println(err)
			}
		case bx, ok := <-tunReadCh:
			if !ok {
				verbose.Println("tun read ch closed")
				break guestLoop
			}
			if ato := g.toWhom(bx.Contents); !ato.IsValid() {
				verbose.Println("dropped non-ip[6]")
				bx.Return()
			} else if ato.IsMulticast() {
				g.multicast(pktTxCh, bx)
			} else {
				g.unicast(pktTxCh, bx, ato)
			}
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
			afrom := bx.AddrPort
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
			tunpi, err = netpdu.TunPI(bx.Contents).Header()
			if err != nil {
				verbose.Println(err)
				bx.Return()
				continue guestLoop
			}
			data = netpdu.TunPI(bx.Contents).Data()
			switch tunpi.Proto {
			case VPN_P_HELLO:
				verbose.Printf("rx %d @ %v hello %v",
					ifrom, afrom, VpnHelloPDU(data))
				bx.Return()
				// FIXME re-checkin if registry era mismatch
				// else re-query exchange from registry
				// if that era is mismatched
			case VPN_P_PUBLIC_KEY:
				blk, _ := pem.Decode(data)
				if blk == nil {
					verbose.Println("encoding")
				} else if err = g.peer(blk); err != nil {
					verbose.Println(err)
				}
				bx.Return()
			case netpdu.ETH_P_IP, netpdu.ETH_P_IPV6:
				tunWriteCh <- bx
			default:
				verbose.Printf("proto[%#x]", tunpi.Proto)
				bx.Return()
			}
		}
	}
	return err
}

type guest struct {
	client
}

func (g *guest) multicast(ch chan<- *box.Box, bx *box.Box) {
	iguest := IdIndex(g.id)
	via := g.via[iguest]
	ivia := IdIndex(via)
	c := g.gcm[ivia]
	bx.From(g.id)
	bx.To(via)
	bx.Via(via)
	bx.CloseWith(c)
	bx.SealWith(c)
	bx.AddrPort = g.service[ivia]
	bx.NonBlockingPut(ch)
}

func (g *guest) unicast(ch chan<- *box.Box, bx *box.Box, addr netip.Addr) {
	bx.From(g.id)
	to, ok := g.addressed[addr]
	if !ok {
		g.whoisAddressed(ch, addr)
		bx.Return()
		return
	}
	bx.To(to)
	ito := IdIndex(to)
	cto, ok := g.gcm[ito]
	if !ok {
		verbose.Printf("%d@%v: %s", ito, addr, "no guest cipher")
		bx.Return()
		return
	}
	bx.CloseWith(cto)
	via, ok := g.via[ito]
	if !ok {
		errata.Println("%d@%v: %s", ito, addr, "no guest exchange")
		bx.Return()
		return
	}
	bx.Via(via)
	ivia := IdIndex(via)
	cvia, ok := g.gcm[ivia]
	if !ok {
		g.whoisId(ch, via)
		bx.Return()
		return
	}
	bx.AddrPort, ok = g.service[ivia]
	if !ok {
		errata.Println("%d: %s", ivia, "no exchange service")
		bx.Return()
		return
	}
	bx.SealWith(cvia)
	bx.NonBlockingPut(ch)
}

func (*guest) toWhom(pdu netpdu.TunPI) (addr netip.Addr) {
	pi, err := pdu.Header()
	d := pdu.Data()
	switch {
	case err != nil:
	case pi.Proto == netpdu.ETH_P_IP:
		if ip, err := netpdu.IP(d).Header(); err == nil {
			addr.UnmarshalBinary(ip.DA[:])
		}
	case pi.Proto == netpdu.ETH_P_IPV6:
		if ip6, err := netpdu.IP6(d).Header(); err == nil {
			addr.UnmarshalBinary(ip6.DA[:])
		}
	}
	return
}

func (g *guest) whoisAddressed(ch chan<- *box.Box, addr netip.Addr) {
	var err error
	a16 := addr.As16()
	bx := box.New()
	if bx.Contents, err = xnet.Add(bx.Contents, netph.TunPI{
		Proto: VPN_P_WHOIS_ADDRESSED,
	}); err != nil {
		errata.Print(err)
		bx.Return()
	} else if bx.Contents, err = xnet.
		Add(bx.Contents, a16[:]); err != nil {
		errata.Print(err)
		bx.Return()
	} else {
		verbose.Println("whois addressed:", addr)
		g.whois(ch, bx)
	}
}

func (g *guest) whoisId(ch chan<- *box.Box, id box.Id) {
	var err error
	bx := box.New()
	if bx.Contents, err = xnet.Add(bx.Contents, netph.TunPI{
		Proto: VPN_P_WHOIS_IDENTIFIED,
	}); err != nil {
		errata.Print(err)
		bx.Return()
	} else if bx.Contents, err = xnet.Add(bx.Contents, id); err != nil {
		errata.Print(err)
		bx.Return()
	} else {
		verbose.Println("whois id:", id)
		g.whois(ch, bx)
	}
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
	bx, err := box.NewReadContents(r)
	defer verbose.Println("stopped", name, "read routine:", err)
	for err == nil && ctx.Err() == nil {
		ch <- bx
		bx, err = box.NewReadContents(r)
	}
}

func tunWriteRoutine(
	ctx context.Context,
	wg *sync.WaitGroup,
	name string,
	w io.Writer,
	ch <-chan *box.Box,
) {
	var (
		tunpdu netpdu.TunPI
		err    error
	)
	defer wg.Done()
	goRoutineTrace.Println("start", name, "write routine")
	defer goRoutineTrace.Println("stopped", name, "write routine")
	for {
		select {
		case <-ctx.Done():
			goRoutineTrace.Println("done")
			return
		case bx, ok := <-ch:
			if !ok {
				goRoutineTrace.Println("tun write ch closed")
				return
			}
			tunpdu = netpdu.TunPI(bx.Contents)
			if _, err = bx.WriteTo(w); err != nil {
				verbose.Print(err)
			} else {
				verbose.Print(tunpdu)
			}
			bx.Return()
		}
	}
}
