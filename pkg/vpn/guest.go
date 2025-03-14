// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
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
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netpdu"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

const MTU = box.ContentMTU - netph.TunPISize

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

	trace := flag.Bool(NameTraceFlag, false, "Log packet forwarding.")
	tflag := flag.Uint(NameTunnelFlag, 0, "Tunnel unit number.")
	err := g.defineAndParseFlags(ctx, defport, args)
	if err != nil {
		return err
	}
	if *trace {
		xlog.UnmuteTrace()
	}

	cctx, cancel := context.WithCancel(ctx)

	udp, err := net.ListenUDP(g.udpv, &net.UDPAddr{
		IP:   g.lap.Addr().AsSlice(),
		Port: int(g.lap.Port()),
	})
	if err != nil {
		return xerrors.Label(err, "ListenUDP")
	}

	defer udp.Close()

	lap, err := netip.ParseAddrPort(udp.LocalAddr().String())
	if err != nil {
		return xerrors.Label(err, "LocalAddr")
	}
	if err = g.register(ctx, netip.AddrPort{}); err != nil {
		return err
	}

	iguest := IdIndex(g.id)
	via := g.via[iguest]

	svc := fmt.Sprintf("(%d via %d@%v)", iguest, via, lap)
	xlog.Info.Println("start", svc)
	defer xlog.Info.Println("stopped", svc)
	defer wg.Wait()
	defer cancel()
	defer xlog.Info.Println("stopping", svc, "...")

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

	mtu := fmt.Sprint(MTU)
	err = nif.Add(cctx, g.hostPrefix, dst, "up", "mtu", mtu)
	if err != nil {
		return xerrors.Label(err, "ifconfig", nif.Name,
			g.hostPrefix.String(), "mtu", mtu)
	}

	err = routeAdd(cctx, g.vpnPrefix, nif)
	if err != nil {
		return xerrors.Note(err,
			"route", "add", g.vpnPrefix, "via", nif)
	}
	defer routeDelete(cctx, g.vpnPrefix, nif)

	pktRxCh := make(chan *box.Box, 4)
	pktTxCh := make(chan *box.Box, 4)
	tunReadCh := make(chan *box.Box, 4)
	tunWriteCh := make(chan *box.Box, 4)

	maps.Copy(netpdu.TunPIprotos, TunPIprotos)

	wg.Add(1)
	go xlog.AlarmHandler(cctx, &wg)
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

	xlog.Info.Printf("start (%s, %v)", nif.Name, g.hostPrefix)
	defer xlog.Info.Printf("stopped (%s, %v)", nif.Name, g.hostPrefix)

	var tunpi netph.TunPI
	var tund []byte

guestLoop:
	for {
		select {
		case <-cctx.Done():
			xlog.Info.Println("done")
			break guestLoop
		case t := <-kat.C:
			err = g.hello(pktTxCh, g.via[iguest], t)
			if err != nil {
				xlog.Errata.Println(err)
			}
		case bx, ok := <-tunReadCh:
			if !ok {
				xlog.Info.Println("tun read ch closed")
				break guestLoop
			}
			if ato := g.toWhom(bx.Contents); !ato.IsValid() {
				xlog.Info.Println("dropped non-ip[6]")
				bx.Return()
			} else if ato.IsMulticast() {
				g.multicast(pktTxCh, bx)
			} else {
				g.unicast(pktTxCh, bx, ato)
			}
		case bx, ok := <-pktRxCh:
			if !ok {
				xlog.Info.Println("pkt tx ch closed")
				break guestLoop
			}
			avia := bx.AddrPort
			ex, ok := g.via[iguest]
			ivia := IdIndex(ex)
			if !ok {
				xlog.Errata.Printf("no cipher for exchange %d",
					ivia)
				bx.Return()
				continue guestLoop
			}
			cex, ok := g.gcm[ivia]
			if !ok {
				xlog.Errata.Printf("no cipher for exchange %d",
					ivia)
				bx.Return()
				continue guestLoop
			}
			if svc := g.service[ivia]; avia != svc {
				xlog.Info.Println(avia, "!=", svc)
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
				xlog.Errata.Printf("no cipher for guest %d",
					ifrom)
				bx.Return()
				continue guestLoop
			}
			if err = bx.UnsealWith(cex); err != nil {
				xlog.Info.Println("unseal:", err)
				bx.Return()
				continue guestLoop
			}
			to := bx.ToWhom()
			if to != g.id {
				err = g.hello(pktTxCh, from, time.Now())
				if err != nil {
					xlog.Errata.Println(err)
				}
				bx.Return()
				continue guestLoop
			}
			if err = bx.OpenWith(cfrom); err != nil {
				xlog.Info.Print(err)
				bx.Return()
				continue guestLoop
			}
			tunpi, tund, err = netpdu.TunPI(bx.Contents).Parse()
			if err != nil {
				xlog.Info.Println(err)
				bx.Return()
				continue guestLoop
			}
			switch tunpi.Proto {
			case VPN_P_HELLO:
				xlog.Info.Printf("rx %d@%v hello %v",
					ifrom, afrom, VpnHelloPDU(tund))
				bx.Return()
				// FIXME re-checkin if registry era mismatch
				// else re-query exchange from registry
				// if that era is mismatched
			case VPN_P_PUBLIC_KEY:
				blk, _ := pem.Decode(tund)
				if blk == nil {
					xlog.Info.Println("encoding")
				} else if err = g.peer(blk); err != nil {
					xlog.Info.Println(err)
				}
				bx.Return()
			case netph.ETH_P_IP, netph.ETH_P_IPV6:
				tunWriteCh <- bx
			default:
				xlog.Info.Printf("proto[%#x]", tunpi.Proto)
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
		xlog.Info.Printf("%d@%v: %s", ito, addr, "no guest cipher")
		bx.Return()
		return
	}
	bx.CloseWith(cto)
	via, ok := g.via[ito]
	if !ok {
		if _, ok := g.service[ito]; ok {
			via = to
		} else {
			xlog.Errata.Printf("%d@%v: %s",
				ito, addr, "no guest exchange")
			bx.Return()
			return
		}
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
		xlog.Errata.Printf("%d: %s", ivia, "no exchange service")
		bx.Return()
		return
	}
	bx.SealWith(cvia)
	bx.NonBlockingPut(ch)
}

func (*guest) toWhom(pdu netpdu.TunPI) (addr netip.Addr) {
	pi, d, err := pdu.Parse()
	switch {
	case err != nil:
	case pi.Proto == netph.ETH_P_IP:
		if ip, _, err := netpdu.IP(d).Parse(); err == nil {
			addr.UnmarshalBinary(ip.DA[:])
		}
	case pi.Proto == netph.ETH_P_IPV6:
		if ip6, _, err := netpdu.IP6(d).Parse(); err == nil {
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
		xlog.Errata.Print(err)
		bx.Return()
	} else if bx.Contents, err = xnet.
		Add(bx.Contents, a16[:]); err != nil {
		xlog.Errata.Print(err)
		bx.Return()
	} else {
		xlog.Info.Println("whois", addr)
		g.whois(ch, bx)
	}
}

func (g *guest) whoisId(ch chan<- *box.Box, id box.Id) {
	var err error
	bx := box.New()
	if bx.Contents, err = xnet.Add(bx.Contents, netph.TunPI{
		Proto: VPN_P_WHOIS_IDENTIFIED,
	}); err != nil {
		xlog.Errata.Print(err)
		bx.Return()
	} else if bx.Contents, err = xnet.Add(bx.Contents, id); err != nil {
		xlog.Errata.Print(err)
		bx.Return()
	} else {
		xlog.Info.Println("whois", id)
		g.whois(ch, bx)
	}
}

func (g *guest) whois(ch chan<- *box.Box, bx *box.Box) {
	via, ok := g.via[IdIndex(g.id)]
	if !ok {
		xlog.Errata.Print("no assigned exchange")
		bx.Return()
		return
	}
	ivia := IdIndex(via)
	cvia, ok := g.gcm[ivia]
	if !ok {
		xlog.Errata.Print("no shared cipher")
		bx.Return()
		return
	}
	svc, ok := g.service[ivia]
	if !ok {
		xlog.Errata.Print("no assigned service")
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
	defer xlog.Info.Println("stopped", name, "read routine")
	xlog.Info.Println("start", name, "read routine")
	for ctx.Err() == nil {
		bx, err := box.NewReadContents(r)
		if err != nil {
			xlog.Errata.Print(name, ": ", err)
			break
		}
		xlog.Info.Println(name, "read", netpdu.TunPI(bx.Contents))
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
	defer xlog.Info.Println("stopped", name, "write routine")
	xlog.Info.Println("start", name, "write routine")
	for {
		select {
		case <-ctx.Done():
			xlog.Info.Print("done")
			return
		case bx, ok := <-ch:
			if !ok {
				xlog.Info.Println(name, "write ch closed")
				return
			}
			xlog.Info.Println(name, "write",
				netpdu.TunPI(bx.Contents))
			_, err := bx.WriteTo(w)
			bx.Return()
			if err != nil {
				xlog.Errata.Print(name, ": ", err)
				return
			}
		}
	}
}
