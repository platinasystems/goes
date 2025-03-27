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

type guest struct {
	client
}

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
		return xerrors.Label(err, "rand")
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

	err = nif.Config(cctx, "up", "mtu", fmt.Sprint(MTU))
	if err != nil {
		return xerrors.Label(err, nif.Name, "mtu", MTU)
	}
	xlog.Info.Println("OK", nif.Name, "up", "mtu", MTU)

	addr := g.hostPrefix.Addr()
	if addr.Is6() {
		addr = addr.WithZone(nif.Name)
	}
	// FIXME probably still need dst w/ ipv4
	_ = dst
	err = nif.Add(cctx, addr, netip.Addr{}, g.hostPrefix.Bits())
	if err != nil {
		return err
	}

	err = routeAdd(cctx, g.vpnPrefix, nif)
	if err != nil {
		return xerrors.Note(err,
			"route", "add", g.vpnPrefix, "via", nif)
	}
	defer routeDelete(cctx, g.vpnPrefix, nif)

	var llu6 netip.Addr
	var addrs []net.Addr

	// FIXME nif.Addrs returns empty
	if addrs, err = nif.Addrs(); err != nil {
		xlog.Errata.Println(err)
	} else if len(addrs) > 0 {
		xlog.Info.Println(nif.Name, "addrs", addrs)
	} else if itf, err := net.InterfaceByName(nif.Name); err != nil {
		xlog.Errata.Println(err)
	} else if addrs, err = itf.Addrs(); err != nil {
		xlog.Errata.Println(err)
	}
	for _, addr := range addrs {
		p, err := netip.ParsePrefix(addr.String())
		if err == nil {
			pa := p.Addr()
			if pa.IsLinkLocalUnicast() {
				llu6 = pa
				break
			}
		}
	}

	pktRxCh := make(chan *box.Box, 4)
	pktTxCh := make(chan *box.Box, 4)
	tunReadCh := make(chan *box.Box, 4)
	tunWriteCh := make(chan *box.Box, 4)

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
			if ato, err := g.toWhom(bx.Contents); err != nil {
				xlog.Errata.Println(err)
				bx.Return()
			} else if !ato.IsValid() {
				xlog.Info.Println("dropped non-ip[6]")
				bx.Return()
			} else if ato.Compare(g.addr) == 0 {
				xlog.Info.Println("loopback", ato)
				tunWriteCh <- bx
			} else if llu6.IsValid() && ato.Compare(llu6) == 0 {
				xlog.Info.Println("link-local loopback", ato)
				tunWriteCh <- bx
			} else if ato.IsMulticast() {
				g.multicast(pktTxCh, bx, ato)
			} else if g.vpnPrefix.Contains(ato) {
				g.unicast(pktTxCh, bx, ato)
			} else {
				xlog.Info.Println(ato, "out of", g.vpnPrefix)
				bx.Return()
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
			// afrom := bx.AddrPort
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
			pdu := VpnPDU(bx.Contents)
			xlog.Info.Print(pdu)
			h, d, err := pdu.Parse()
			if err != nil {
				xlog.Errata.Println(err)
				bx.Return()
				continue guestLoop
			}
			switch h.Proto {
			case VPN_P_HELLO:
				// FIXME re-checkin if registry era mismatch
				// else re-query exchange from registry
				// if that era is mismatched
				bx.Return()
			case VPN_P_PUBLIC_KEY:
				blk, _ := pem.Decode(d)
				if blk == nil {
					xlog.Info.Println("encoding")
				} else if err = g.peer(blk); err != nil {
					xlog.Info.Println(err)
				}
				bx.Return()
			case VPN_P_IP:
				xnet.Overwrite(bx.Contents, netph.TunPI{
					Proto: netph.TUN_P_IP,
				})
				tunWriteCh <- bx
			case VPN_P_IP6:
				xnet.Overwrite(bx.Contents, netph.TunPI{
					Proto: netph.TUN_P_IP6,
				})
				tunWriteCh <- bx
			default:
				bx.Return()
			}
		}
	}
	return err
}

func (g *guest) multicast(ch chan<- *box.Box, bx *box.Box, addr netip.Addr) {
	setVpnProto(bx, addr.Is6())
	xlog.Info.Println("multicast", VpnPDU(bx.Contents))
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
	setVpnProto(bx, addr.Is6())
	xlog.Info.Println("unicast", VpnPDU(bx.Contents))
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

func setVpnProto(bx *box.Box, is6 bool) {
	ethp := uint16(VPN_P_IP)
	if is6 {
		ethp = VPN_P_IP6
	}
	xnet.Overwrite(bx.Contents, netph.TunPI{
		Proto: ethp,
	})
}

func (*guest) toWhom(pdu netpdu.TunPI) (addr netip.Addr, err error) {
	h, d, err := pdu.Parse()
	if err != nil {
		return
	}
	switch h.Proto {
	case netph.TUN_P_IP:
		var ip netph.IP
		if ip, _, err = netpdu.IP(d).Parse(); err == nil {
			addr.UnmarshalBinary(ip.DA[:])
		}
	case netph.TUN_P_IP6:
		var ip6 netph.IP6
		if ip6, _, err = netpdu.IP6(d).Parse(); err == nil {
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
	defer close(ch)
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
