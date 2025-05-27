// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"encoding/pem"
	"flag"
	"fmt"
	"net"
	"net/netip"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/platinasystems/goes/v2/pkg/box"
	"github.com/platinasystems/goes/v2/pkg/netif"
	"github.com/platinasystems/goes/v2/pkg/nettun"
	"github.com/platinasystems/goes/v2/pkg/xcontext"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netpdu"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
	"github.com/platinasystems/goes/v2/pkg/xos"
)

const MTU = box.ContentMTU - netph.TunPISize

// Guest is a UDP server that forwards ciphered packets between an exchange
// and a network tunnel interface.
func Guest(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [vpn]
Forward ciphered packets between exchange and tunnel interface.

{{flags .}}`)

	var g guest

	defineListen(0)
	defineTunnel()
	enableQuiet()
	enableTrace()
	enableVerbose()
	g.rest.defineFlags()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if err = g.config(); err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(ctx)

	conn, err := net.ListenUDP(g.udpv, &net.UDPAddr{
		IP:   vpnListen.Addr().AsSlice(),
		Port: int(vpnListen.Port()),
	})
	if err != nil {
		return xerrors.Label(err, "ListenUDP")
	}
	wg.Go(func() {
		defer conn.Close()
		<-ctx.Done()
	})

	lap, err := netip.ParseAddrPort(conn.LocalAddr().String())
	if err != nil {
		return xerrors.Label(err, "LocalAddr")
	}
	if err = g.register(ctx, netip.AddrPort{}); err != nil {
		return err
	}

	iguest := g.id.Index()
	vguest := g.id.Version()
	via := g.via[iguest]

	svc := fmt.Sprintf("%v via %v@%v", iguest, via, lap)
	xlog.Info.Println("start", svc)

	defer xlog.Info.Println("stopped", svc)
	defer wg.Wait()
	defer cancel()
	defer xlog.Info.Println("stopping", svc, "...")

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

	g.tun, err = nettun.New(vpnTunnel, istap, persist, owner, group, ha)
	if err != nil {
		return err
	}
	wg.Go(func() {
		defer g.tun.Close()
		<-ctx.Done()
	})

	if err := xos.SetNonblockFile(g.tun, true); err != nil {
		return err
	}

	nif, err := netif.Named(ctx, g.tun.Name())
	if err != nil {
		return xerrors.Mark(err)
	} else if nif == nil {
		return xerrors.NotFound(g.tun.Name())
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

	wg.Go(func() { xlog.AlarmHandler(cctx) })
	wg.Go(func() { pktRx(cctx, conn, g.pkt.rx) })
	wg.Go(func() { pktTx(cctx, conn, g.pkt.tx) })
	wg.Go(func() { g.tunRead(cctx) })
	wg.Go(func() { g.tunWrite(cctx) })
	wg.Go(func() { g.whoisRoutine(cctx) })

	xlog.Info.Printf("start (%s, %v)", nif.Name, g.hostPrefix)
	defer xlog.Info.Printf("stopped (%s, %v)", nif.Name, g.hostPrefix)
	defer func() {
		r := recover()
		if r != nil {
			xlog.Errata.Print(r)
			debug.PrintStack()
		}
	}()

	if err = g.hello(ctx, via, time.Now()); err != nil {
		return err
	}

	kat := time.NewTicker(10 * time.Second)
	defer kat.Stop()

guestLoop:
	for {
		select {
		case <-cctx.Done():
			xlog.Info.Println("done")
			break guestLoop
		case err = <-g.fault:
			break guestLoop
		case t := <-kat.C:
			if err = g.hello(ctx, via, t); err != nil {
				xlog.Errata.Println(err)
			}
		case buf, ok := <-g.whois.response:
			if !ok {
				xlog.Info.Println("closed whois response")
				break guestLoop
			}
			blk, _ := pem.Decode(buf.Bytes())
			if blk == nil {
				err = xerrors.Invalid("encoding")
				break guestLoop
			}
			if err = g.peer(blk); err != nil {
				break guestLoop
			}
		case bx, ok := <-g.q.read:
			if !ok {
				xlog.Info.Println("closed tun read")
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
				bx.Queue(ctx, g.q.write)
			} else if llu6.IsValid() && ato.Compare(llu6) == 0 {
				xlog.Info.Println("link-local loopback", ato)
				bx.Queue(ctx, g.q.write)
			} else if ato.IsMulticast() {
				g.multicast(cctx, bx, ato)
			} else if ato.IsLinkLocalUnicast() ||
				g.vpnPrefix.Contains(ato) {
				g.unicast(cctx, bx, ato)
			} else {
				xlog.Info.Println(ato, "out of", g.vpnPrefix)
				bx.Return()
			}
		case bx, ok := <-g.pkt.rx:
			if !ok {
				xlog.Info.Println("closed pkt rx")
				break guestLoop
			}
			from := bx.FromWhom()
			ifrom, vfrom := from.Index(), from.Version()
			if v, ok := g.ver[ifrom]; !ok || v != vfrom {
				g.queueWhoisRequest(ctx, ifrom)
				bx.Return()
				continue guestLoop
			}
			via := bx.ViaWhom()
			ivia, vvia := via.Index(), via.Version()
			if v, ok := g.ver[ivia]; !ok || v != vvia {
				xlog.Info.Println("update", ivia)
				g.queueWhoisRequest(ctx, ivia)
				bx.Return()
				continue guestLoop
			}
			svc, ok := g.service[ivia]
			if !ok || svc.Compare(bx.AddrPort) != 0 {
				g.service[ivia] = bx.AddrPort
			}
			opener, ok := g.gcm[ivia]
			if !ok {
				xlog.Errata.Print("no cipher for exchange %d",
					ivia)
				bx.Return()
				continue guestLoop
			}
			if err = bx.UnsealWith(opener); err != nil {
				xlog.Errata.Println("unseal:", err)
				bx.Return()
				continue guestLoop
			}
			to := bx.ToWhom()
			ito, vto := to.Index(), to.Version()
			if ito == iguest {
				if vto != vguest {
					xlog.Errata.Print("wrong version")
					bx.Return()
					continue guestLoop
				}
				if opener, ok = g.gcm[ifrom]; !ok {
					xlog.Errata.Printf("no cipher for "+
						"guest %d", ifrom)
					bx.Return()
					continue guestLoop
				}
			}
			if err = bx.OpenWith(opener); err != nil {
				xlog.Info.Print(err)
				bx.Return()
				continue guestLoop
			}
			xlog.Trace.Println("rx", Box{bx})
			hvpn, dvpn, err := Box{bx}.PDU().Parse()
			if err != nil {
				xlog.Errata.Println(err)
				bx.Return()
				continue guestLoop
			}
			switch hvpn.Proto {
			case VPN_P_HELLO:
				// FIXME verify exchange version.
				xlog.Trace.Println(Box{bx})
				bx.Return()
			case VPN_P_PUBLIC_KEY:
				blk, _ := pem.Decode(dvpn)
				if blk == nil {
					xlog.Errata.Println("encoding?")
				} else if err = g.peer(blk); err != nil {
					xlog.Errata.Println(err)
				}
				bx.Return()
			case VPN_P_IP:
				xnet.Encode(bx.Contents, netph.TunPI{
					Proto: netph.TUN_P_IP,
				})
				if ip, _, err := netpdu.
					IP(dvpn).Parse(); err != nil {
					xlog.Errata.Print(err)
					bx.Return()
				} else {
					sa := netip.AddrFrom4(ip.SA)
					g.addressed[sa] = from
					bx.Queue(ctx, g.q.write)
				}
			case VPN_P_IP6:
				xnet.Encode(bx.Contents, netph.TunPI{
					Proto: netph.TUN_P_IP6,
				})
				if ip, _, err := netpdu.
					IP6(dvpn).Parse(); err != nil {
					xlog.Errata.Print(err)
					bx.Return()
				} else {
					sa := netip.AddrFrom16(ip.SA)
					g.addressed[sa] = from
					bx.Queue(ctx, g.q.write)
				}
			default:
				xlog.Errata.Printf("invalid proto %#x",
					hvpn.Proto)
				bx.Return()
			}
		}
	}
	return err
}

type guest struct {
	client
	q struct {
		read, write chan *box.Box
	}
	tun *os.File
}

func (g *guest) config() error {
	g.q.read = make(chan *box.Box, 16)
	g.q.write = make(chan *box.Box, 16)
	return g.client.config()
}

func (g *guest) hello(ctx context.Context, to box.Id, now time.Time) error {
	var err error
	ito := to.Index()
	cto, ok := g.gcm[ito]
	if !ok {
		xlog.Info.Println("hello, whois", to)
		g.queueWhoisRequest(ctx, to)
		return nil
	}
	via, ok := g.via[ito]
	if !ok {
		via = to
	}
	ivia := via.Index()
	cvia, ok := g.gcm[ivia]
	if !ok {
		return fmt.Errorf("%d: no exchange cipher", ivia)
	}
	avia, ok := g.service[ivia]
	if !ok {
		return fmt.Errorf("%d: no exchange service", ivia)
	}
	bx := box.New()
	bx.Contents, err = xnet.Attach(bx.Contents, netph.TunPI{
		Proto: VPN_P_HELLO,
	})
	if err != nil {
		bx.Return()
		return err
	}
	bx.Contents, err = xnet.Attach(bx.Contents, now.UnixMicro())
	if err != nil {
		bx.Return()
		return err
	}
	bx.AddrPort = avia
	bx.From(g.id)
	bx.To(to)
	bx.Via(via)
	xlog.Trace.Println(Box{bx})
	bx.CloseWith(cto)
	bx.SealWith(cvia)
	if !xcontext.Queue(ctx, g.pkt.tx, bx) {
		bx.Return()
	}
	return nil
}

func (g *guest) multicast(ctx context.Context, bx *box.Box, addr netip.Addr) {
	g.setVpnProto(bx, addr.Is6())
	iguest := g.id.Index()
	via := g.via[iguest]
	ivia := via.Index()
	c, ok := g.gcm[ivia]
	if !ok || c == nil {
		xlog.Info.Println("whois", ivia)
		g.queueWhoisRequest(ctx, ivia)
		bx.Return()
		return
	}
	bx.From(g.id)
	bx.To(via)
	bx.Via(via)
	xlog.Trace.Println("multicast", Box{bx})
	bx.CloseWith(c)
	bx.SealWith(c)
	bx.AddrPort = g.service[ivia]
	bx.Queue(ctx, g.pkt.tx)
}

func (g *guest) unicast(ctx context.Context, bx *box.Box, addr netip.Addr) {
	g.setVpnProto(bx, addr.Is6())
	bx.From(g.id)
	to, ok := g.addressed[addr]
	if !ok {
		xlog.Info.Println("whois", addr)
		g.queueWhoisRequest(ctx, addr)
		bx.Return()
		return
	}
	bx.To(to)
	ito := to.Index()
	cto, ok := g.gcm[ito]
	if !ok || cto == nil {
		xlog.Info.Print("whois", ito)
		bx.Return()
		return
	}
	via, ok := g.via[ito]
	if !ok {
		if _, ok := g.service[ito]; ok {
			via = to
		} else {
			xlog.Errata.Printf("%v@%v: %s", to, addr, "no via")
			bx.Return()
			return
		}
	}
	bx.Via(via)
	ivia := via.Index()
	cvia, ok := g.gcm[ivia]
	if !ok {
		g.queueWhoisRequest(ctx, ivia)
		bx.Return()
		return
	}
	bx.AddrPort, ok = g.service[ivia]
	if !ok {
		xlog.Errata.Printf("%d: %s", ivia, "no exchange service")
		bx.Return()
		return
	}
	xlog.Trace.Println("unicast", Box{bx})
	bx.CloseWith(cto)
	bx.SealWith(cvia)
	bx.Queue(ctx, g.pkt.tx)
}

func (*guest) setVpnProto(bx *box.Box, is6 bool) {
	ethp := uint16(VPN_P_IP)
	if is6 {
		ethp = VPN_P_IP6
	}
	xnet.Encode(bx.Contents, netph.TunPI{
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

func (g *guest) tunRead(ctx context.Context) {
	name := g.tun.Name()
	xlog.Info.Println("start", name, "read routine")
	defer xlog.Info.Println("stopped", name, "read routine")

	bx := box.New()
	for !xcontext.IsDone(ctx) {
		if err := bx.ReadContents(g.tun); err == nil {
			pi := netpdu.TunPI(bx.Contents)
			xlog.Trace.Println("read", name, pi)
			if xcontext.Queue(ctx, g.q.read, bx) {
				bx = box.New()
			} else {
				break
			}
		} else if !xos.IsBlocked(err) {
			xlog.Errata.Print(name, err)
			break
		} else {
			runtime.Gosched()
		}
	}
	bx.Return()
}

func (g *guest) tunWrite(ctx context.Context) {
	name := g.tun.Name()
	xlog.Info.Println("start", name, "write routine")
	defer xlog.Info.Println("stopped", name, "write routine")

	xcontext.Range(ctx, g.q.write, func(bx *box.Box) (ok bool) {
		for {
			if _, err := g.tun.Write(bx.Contents); err == nil {
				xlog.Trace.Println("wrote", name,
					netpdu.TunPI(bx.Contents))
				ok = true
				break
			} else if !xos.IsBlocked(err) {
				xlog.Errata.Print(name, err)
				break
			}
			runtime.Gosched()
			if xcontext.IsDone(ctx) {
				break
			}
		}
		bx.Return()
		return
	})
}

func (g *guest) queueWhoisRequest(ctx context.Context, req any) bool {
	select {
	case <-ctx.Done():
		return false
	case g.whois.request <- req:
		return true
	}
}

func (g *guest) whoisRoutine(ctx context.Context) {
	xlog.Info.Println("start guest whois routine")

	defer xlog.Info.Println("stopped guest whois routine")
	defer close(g.whois.response)

	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-g.whois.request:
			if !ok {
				return
			}
			err := g.client.rest.whois(ctx, g.whois.response, v)
			if err != nil {
				xlog.Errata.Printf("whois %v: %v", v, err)
			}
		}
	}
}
