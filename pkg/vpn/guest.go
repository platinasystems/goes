// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"os"
	"runtime"
	"slices"
	"time"

	"github.com/platinasystems/goes/v2/pkg/netif"
	"github.com/platinasystems/goes/v2/pkg/nettun"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netpdu"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
	"github.com/platinasystems/goes/v2/pkg/xos"
)

const (
	GCMNonceSize = 12
	GCMTagSize   = 16
	GCMOverhead  = GCMNonceSize + GCMTagSize

	VPNMTU  = netph.ETHMTU - VPNSize
	IP6MTU  = VPNMTU - netph.IP6Size
	UDP6MTU = IP6MTU - netph.UDPSize
	TunMTU  = UDP6MTU - GCMOverhead - SizeofLabel

	IsTap = false

	TunOwner = -1
	TunGroup = -1

	TunPersist = false
)

type GuestReceipt struct {
	Id     Id
	Prefix netip.Prefix
}

// Guest is a UDP server that forwards ciphered packets between an exchange
// and a network tunnel interface.
func Guest(ctx context.Context, args []string) error {
	var svc string
	defer xlog.Info.Println("stopped", svc)

	xflag.TemplateUsage(`
usage: {{.Name}} [flags] <via[:port]>
Forward ciphered packets between exchange and tunnel interface.

{{flags .}}`)

	defineListen(0)
	defineTunnel()
	defineRestFlags()
	enableQuiet()
	enableTrace()
	enableVerbose()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if flag.CommandLine.NArg() == 0 {
		return xerrors.Incomplete("via[:port]")
	}
	via := flag.CommandLine.Arg(0)
	g := guest{
		addressed: make(map[netip.Addr]*contact),
		indexed:   make(map[int]*contact),
		netC:      make(chan *xnet.Msg),
		tunC:      make(chan *xnet.Msg),
	}
	if err = g.subscriber.config(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer wg.Wait()
	defer cancel()

	wg.Go(func() { xlog.AlarmHandler(ctx) })

	if err = g.checkin(ctx, via); err != nil {
		return err
	}
	if err = g.waitForExchange(ctx, via); err != nil {
		return err
	}

	if g.udp, err = udpListen(vpnListen); err != nil {
		return err
	}
	wg.Go(func() { g.netRx(ctx) })
	wg.Go(func() {
		defer xlog.Info.Println("closed", g.udp.LocalAddr())
		defer g.udp.Close()
		<-ctx.Done()
	})

	now := time.Now().UnixMicro()
	err = greetings(ctx, g.udp, g.start, now, g.via.Id, g.via.Service)
	if err != nil {
		return err
	}

	svc = fmt.Sprintf("%s %v@%v via %s %v@%v",
		g.subscriber.rest.crt.Subject.CommonName,
		MyId, g.udp.LocalAddr(),
		g.via.Name, g.via.Id, g.via.Service)

	xlog.Info.Println("start", svc)
	defer xlog.Info.Println("stopping", svc, "...")

	ha := netif.NewHardwareAddr()
	if err = ha.Rand(); err != nil {
		return err
	}

	g.tun, err = nettun.
		New(vpnTunnel, IsTap, TunPersist, TunOwner, TunGroup, ha)
	if err != nil {
		return err
	}
	wg.Go(func() {
		defer g.tun.Close()
		<-ctx.Done()
	})

	if runtime.NumCPU() == 1 {
		if err = xos.SetNonblockFile(g.tun, true); err != nil {
			return err
		}
	}

	nif, err := netif.Named(ctx, g.tun.Name())
	if err != nil {
		return err
	}
	if nif == nil {
		return xerrors.NotFound(g.tun.Name())
	}

	if err = nif.Config(ctx, "up", "mtu", fmt.Sprint(TunMTU)); err != nil {
		return err
	}
	xlog.Info.Println(nif.Name, "up", "mtu", TunMTU)

	addr := g.prefix.Addr()
	if addr.Is6() {
		addr = addr.WithZone(nif.Name)
	}
	bits := 32
	if addr.Is6() {
		bits = 128
	}
	// FIXME probably still need dst w/ ipv4
	if err = nif.Add(ctx, addr, netip.Addr{}, bits); err != nil {
		return err
	}
	xlog.Info.Println(nif.Name, addr)

	g.vpn = g.prefix.Masked()
	if err = routeAdd(ctx, g.vpn, nif); err != nil {
		return err
	}
	xlog.Info.Println(nif.Name, g.vpn)
	defer routeDelete(ctx, g.vpn, nif)

	addrs, err := nif.Addrs()
	if err != nil {
		return err
	} else if len(addrs) > 0 {
		xlog.Info.Println(nif.Name, "addrs", addrs)
	} else if itf, err := net.InterfaceByName(nif.Name); err != nil {
		return err
	} else if addrs, err = itf.Addrs(); err != nil {
		return err
	}
	for _, addr := range addrs {
		p, err := netip.ParsePrefix(addr.String())
		if err == nil {
			pa := p.Addr()
			if pa.IsLinkLocalUnicast() {
				g.llu6 = pa
				break
			}
		}
	}

	wg.Go(func() { g.tunRead(ctx) })
	wg.Go(func() { g.whoisService(ctx) })
	defer close(g.whoisReqCh)

	tkr := time.NewTicker(10 * time.Second)
	defer tkr.Stop()

selection:
	for err == nil {
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case err = <-g.fault:
		case t := <-tkr.C:
			now = t.UnixMicro()
			err = greetings(ctx, g.udp, g.start, now, g.via.Id,
				g.via.Service)
		case reg, ok := <-g.whoisRspCh:
			if !ok {
				break selection
			}
			g.peer(ctx, reg)
		case m, ok := <-g.tunC:
			if !ok {
				break selection
			}
			da := m.AddrPort.Addr()
			if to, ok := g.addressed[da]; ok {
				g.tx(ctx, to, m)
				g.last.send = to
				mp.Put(m)
			} else {
				g.pending.tx = append(g.pending.tx, m)
				g.whoisQueue(ctx, da)
			}
		case m, ok := <-g.netC:
			if !ok {
				break selection
			}
			tid, fid := ScanLabel(m.Data)
			if tid == fid {
				g.verifyHello(m, fid)
				continue selection
			}
			if tid != MyId {
				xlog.Trace.Print("dropped ", tid, "<-", fid)
				continue selection
			}
			from, ok := g.indexed[fid.Index()]
			if ok && from.id.Version() == fid.Version() {
				g.rx(ctx, from, m)
				g.last.recv = from
				mp.Put(m)
			} else {
				g.pending.rx = append(g.pending.rx, m)
				g.whoisQueue(ctx, fid)
			}
		}
	}
	return xerrors.Suppress(err, context.Canceled, errEOC)
}

type guest struct {
	subscriber

	addressed map[netip.Addr]*contact
	indexed   map[int]*contact

	netC, tunC chan *xnet.Msg

	last struct{ recv, send *contact }

	llu6 netip.Addr

	prefix,
	vpn netip.Prefix

	priv *ecdh.PrivateKey
	tun  *os.File

	// msgs pending whois response
	pending struct{ rx, tx []*xnet.Msg }

	udp *net.UDPConn
}

type contact struct {
	id   Id
	addr netip.Addr
	cb   cipher.Block
	gcm  cipher.AEAD
	fromMe,
	toMe []byte
}

func (g *guest) checkin(ctx context.Context, via string) error {
	var err error
	var receipt GuestReceipt

	buf := g.subscriber.rest.alloc()
	defer g.subscriber.rest.free(buf)

	g.priv, err = ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	pubder, err := x509.MarshalPKIXPublicKey(g.priv.PublicKey())
	if err != nil {
		return err
	}
	rsp, err := g.subscriber.rest.request(ctx, buf, http.MethodPut,
		"application/pkixcmp", bytes.NewReader(pubder),
		RestOp, RestOpCheckin,
		RestOpCheckin, RestOpCheckinGuest,
		RestOpCheckinGuestVia, via)
	if err != nil {
		return err
	}
	if err = g.validateCheckinResponse(rsp); err != nil {
		return err
	}
	if err = json.Unmarshal(buf.Bytes(), &receipt); err != nil {
		return err
	}
	MyId = receipt.Id
	MyLabel = MakeLabel(MyId, MyId)
	g.prefix = receipt.Prefix
	xlog.Info.Print("registered: ", MyId, "@", g.prefix)
	return nil
}

func (g guest) netRx(ctx context.Context) {
	var (
		n   int
		err error
	)
	la := g.udp.LocalAddr()
	xlog.Info.Println("start stream from", la)
	defer xlog.Info.Println("stopped stream from", la)
	defer close(g.netC)
	m := mp.Get()
	defer mp.Put(m)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		m.Data = m.Data[:cap(m.Data)]
		n, m.AddrPort, err = g.udp.ReadFromUDPAddrPort(m.Data)
		if err != nil {
			err = xerrors.Suppress(err, net.ErrClosed)
			if err != nil {
				xlog.Errata.Print(err)
			}
			return
		}
		m.Data = m.Data[:n]
		i := len(m.Data) - SizeofLabel
		if i < 0 {
			xlog.Errata.Print("underrun")
			continue
		}
		last := g.last.recv
		if last == nil ||
			bytes.Compare(m.Data[i:], g.last.recv.toMe) != 0 {
			select {
			case <-ctx.Done():
				mp.Put(m)
				return
			case g.netC <- m:
				m = mp.Get()
			}
			continue
		}
		to, from := ScanLabel(m.Data)
		last.cb.Decrypt(m.Data[:aes.BlockSize], m.Data[:aes.BlockSize])
		m.Data, err = last.gcm.
			Open(m.Data[:0], nil, m.Data[:i], m.Data[i:])
		if err != nil {
			xlog.Trace.Println("open:", err)
			continue
		}
		xlog.Trace.Print("rx ", to, "<-", from, netpdu.Mark, PDU(m.Data))
		_, err = g.tun.Write(m.Data)
		if err != nil {
			xlog.Trace.Print(g.tun.Name, ":write: ", err)
			return
		}
	}
}

func (g *guest) pendingDiscard(reg *Registration) {
	g.pendingMatchRx(reg.Id, func(m *xnet.Msg) {
		mp.Put(m)
	})
	g.pendingMatchTx(reg.Addr, func(m *xnet.Msg) {
		mp.Put(m)
	})
}

func (g *guest) pendingMatchRx(id Id, f func(*xnet.Msg)) {
	var m *xnet.Msg
	for i := 0; i < len(g.pending.rx); {
		m = g.pending.rx[i]
		_, from := ScanLabel(m.Data)
		if id.Index() == from.Index() {
			g.pending.rx = slices.Delete(g.pending.rx, i, i+1)
			f(m)
		} else {
			i += 1
		}
	}
}

func (g *guest) pendingMatchTx(addr netip.Addr, f func(*xnet.Msg)) {
	for i := 0; i < len(g.pending.tx); {
		m := g.pending.tx[i]
		if addr.Compare(m.AddrPort.Addr()) == 0 {
			g.pending.tx = slices.Delete(g.pending.tx, i, i+1)
			f(m)
		} else {
			i += 1
		}
	}
}

func (g *guest) peer(ctx context.Context, reg *Registration) {
	c := new(contact)
	c.id = reg.Id
	c.addr = reg.Addr
	c.fromMe = MakeLabel(MyId, reg.Id)
	c.toMe = MakeLabel(reg.Id, MyId)
	regi := reg.Id.Index()
	verifier, err := reg.verifier()
	if err != nil {
		g.pendingDiscard(reg)
		xlog.Errata.Println(err)
		return
	}
	g.verifiers[regi] = verifier
	anyk, err := x509.ParsePKIXPublicKey(reg.PubKey)
	if err != nil {
		g.pendingDiscard(reg)
		xlog.Errata.Println(err)
		return
	}
	ecdhpub, ok := anyk.(*ecdh.PublicKey)
	if !ok {
		g.pendingDiscard(reg)
		xlog.Errata.Printf("%T: unsupported", anyk)
		return
	}
	shared, err := g.priv.ECDH(ecdhpub)
	if err != nil {
		g.pendingDiscard(reg)
		xlog.Errata.Println(err)
		return
	}
	if c.cb, err = aes.NewCipher(shared); err != nil {
		g.pendingDiscard(reg)
		xlog.Errata.Println(err)
		return
	}
	if c.gcm, err = cipher.NewGCMWithRandomNonce(c.cb); err != nil {
		g.pendingDiscard(reg)
		xlog.Errata.Println(err)
		return
	}
	g.addressed[reg.Addr] = c
	g.indexed[regi] = c
	g.pendingMatchRx(c.id, func(m *xnet.Msg) {
		g.rx(ctx, c, m)
		mp.Put(m)
	})
	g.last.recv = c
	g.pendingMatchTx(c.addr, func(m *xnet.Msg) {
		g.tx(ctx, c, m)
		mp.Put(m)
	})
	g.last.send = c
}

func (g *guest) rx(ctx context.Context, from *contact, m *xnet.Msg) {
	var err error
	from.cb.Decrypt(m.Data[:aes.BlockSize], m.Data[:aes.BlockSize])
	i := len(m.Data) - SizeofLabel
	m.Data, err = from.gcm.Open(m.Data[:0], nil, m.Data[:i], m.Data[i:])
	if err != nil {
		xlog.Trace.Println("open:", err)
		return
	}
	if _, err = g.tun.Write(m.Data); err != nil {
		xlog.Trace.Print(g.tun.Name(), ":write: ", err)
		return
	}
	xlog.Trace.Print("rx ", MyId, "<-", from.id, netpdu.Mark, PDU(m.Data))
}

func (*guest) setVpnProto(m *xnet.Msg, is6 bool) {
	ethp := uint16(VPN_P_IP)
	if is6 {
		ethp = VPN_P_IP6
	}
	xnet.Encode(m.Data, netph.TunPI{
		Proto: ethp,
	})
}

func (g *guest) tunRead(ctx context.Context) {
	name := g.tun.Name()
	xlog.Info.Println("start", name, "read")
	xlog.Info.Println("stopped", name, "read")
	defer close(g.tunC)
	m := mp.Get()
	defer mp.Put(m)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		m.Data = m.Data[:cap(m.Data)]
		n, err := g.tun.Read(m.Data)
		if err != nil {
			if xos.IsBlocked(err) {
				runtime.Gosched()
				continue
			}
			xlog.Errata.Print(err)
			return
		}
		m.Data = m.Data[:n]
		da, err := netpdu.TunPI(m.Data).ToWhom()
		if err != nil {
			xlog.Trace.Println("dropped no tunPI")
			continue
		}
		if !da.IsValid() {
			xlog.Trace.Println("dropped non-ip[6]")
			continue
		}
		m.AddrPort = netip.AddrPortFrom(da, 0)
		if g.last.send != nil && g.last.send.addr.Compare(da) == 0 {
			g.tx(ctx, g.last.send, m)
		}
		if da.IsMulticast() {
			xlog.Trace.Println("dropped multicast", da)
		} else if da.Compare(g.prefix.Addr()) == 0 {
			xlog.Trace.Println("loopback", da)
			g.tun.Write(m.Data)
		} else if g.llu6.IsValid() && da.Compare(g.llu6) == 0 {
			xlog.Trace.Println("link-local loopback", da)
			g.tun.Write(m.Data)
		} else if !g.vpn.Contains(da) {
			xlog.Trace.Println(da, "out of", g.vpn)
			continue
		} else {
			select {
			case <-ctx.Done():
				mp.Put(m)
				return
			case g.tunC <- m:
				m = mp.Get()
			}
		}
	}
}

func (g *guest) tx(ctx context.Context, to *contact, m *xnet.Msg) {
	pdu := netpdu.TunPI(m.Data)
	pdu.Proto(m.AddrPort.Addr().Is6())
	xlog.Trace.Print("tx ", to.id, "<-", MyId, netpdu.Mark, pdu)
	m.Data = to.gcm.Seal(m.Data[:0], nil, m.Data, to.fromMe)
	to.cb.Encrypt(m.Data[:aes.BlockSize], m.Data[:aes.BlockSize])
	m.Data = append(m.Data, to.fromMe...)
	g.udp.WriteToUDPAddrPort(m.Data, g.via.Service)
}
