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
	CipherOverhead = 28

	IP6MTU  = netph.ETHMTU - netph.IP6Size
	UDP6MTU = IP6MTU - netph.UDPSize
	VPNMTU  = UDP6MTU - VPNSize
	TunMTU  = VPNMTU - CipherOverhead

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
		addressed: make(map[netip.Addr]Id),
		cb:        make(map[int]cipher.Block),
		gcm:       make(map[int]cipher.AEAD),
		lbl:       make(map[int][]byte),
		tunmsgs:   makeGuestTunMsgs(),
		ver:       make(map[int]uint8),
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

	if err = udpListen(); err != nil {
		return err
	}
	wg.Go(func() {
		if t := udpStream(ctx); err == nil {
			err = t
		}
	})
	wg.Go(func() {
		defer xlog.Info.Println("closed", udp.LocalAddr())
		defer udp.Close()
		<-ctx.Done()
	})

	err = g.greetings(ctx, time.Now(), g.via.Id, g.via.Service)
	if err != nil {
		return err
	}

	svc = fmt.Sprintf("%s %v@%v via %s %v@%v",
		g.subscriber.rest.crt.Subject.CommonName,
		MyId, udp.LocalAddr(),
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

	wg.Go(func() {
		name := g.tun.Name()
		xlog.Info.Println("start read", name)
		err := mp.ReadMsgs(ctx, g.tun, g.tunmsgs)
		if err != nil {
			xlog.Errata.Println("egress read", name, err)
		} else {
			xlog.Info.Println("stopped read", name)
		}
	})

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
			err = g.greetings(ctx, t, g.via.Id, g.via.Service)
		case gx, ok := <-g.whoisRspCh:
			if !ok {
				break selection
			}
			g.peer(ctx, gx)
		case m, ok := <-g.tunmsgs:
			if !ok {
				break selection
			}
			err = g.tx(ctx, m)
		case m, ok := <-rmc:
			if !ok {
				break selection
			}
			err = g.rx(ctx, m)
		}
	}
	return xerrors.Suppress(err, context.Canceled, errEOC)
}

func makeGuestTunMsgs() chan *xnet.Msg {
	if runtime.NumCPU() > 1 {
		return make(chan *xnet.Msg, 4)
	}
	return make(chan *xnet.Msg)
}

type guest struct {
	subscriber
	tunmsgs   chan *xnet.Msg
	addressed map[netip.Addr]Id

	llu6 netip.Addr

	prefix,
	vpn netip.Prefix

	priv *ecdh.PrivateKey
	cb   map[int]cipher.Block
	gcm  map[int]cipher.AEAD
	lbl  map[int][]byte
	ver  map[int]uint8
	tun  *os.File

	// msgs pending whois response
	pending struct{ rx, tx []*xnet.Msg }
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
	xlog.Info.Println("registration:", MyId, "@", g.prefix)
	return nil
}

func (g *guest) isOK(id Id) bool {
	v, ok := g.ver[id.Index()]
	return ok && v == id.Version()
}

func (g *guest) pendingDiscard(reg *Registration) {
	g.pendingTxMatch(reg.Addr, mp.Put)
}

func (g *guest) pendingRx(ctx context.Context, reg *Registration) {
	g.pendingRxMatch(reg.Id, func(m *xnet.Msg) {
		g.rx(ctx, m)
	})
}

func (g *guest) pendingRxMatch(id Id, f func(*xnet.Msg)) {
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

func (g *guest) pendingTx(ctx context.Context, reg *Registration) {
	g.pendingTxMatch(reg.Addr, func(m *xnet.Msg) {
		g.tx(ctx, m)
	})
}

func (g *guest) pendingTxMatch(addr netip.Addr, f func(*xnet.Msg)) {
	var m *xnet.Msg
	for i := 0; i < len(g.pending.tx); {
		m = g.pending.tx[i]
		if da, err := netpdu.TunPI(m.Data).ToWhom(); err != nil {
			g.pending.tx = slices.Delete(g.pending.tx, i, i+1)
		} else if da.Compare(addr) == 0 {
			g.pending.tx = slices.Delete(g.pending.tx, i, i+1)
			f(m)
		} else {
			i += 1
		}
	}
}

func (g *guest) peer(ctx context.Context, reg *Registration) {
	var err error

	regi := reg.Id.Index()
	g.ver[regi] = reg.Id.Version()
	if g.verifiers[regi], err = reg.verifier(); err != nil {
		g.pendingDiscard(reg)
		xlog.Errata.Println(err)
		return
	}
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
	cb, err := aes.NewCipher(shared)
	if err != nil {
		g.pendingDiscard(reg)
		xlog.Errata.Println(err)
		return
	}
	gcm, err := cipher.NewGCMWithRandomNonce(cb)
	if err != nil {
		g.pendingDiscard(reg)
		xlog.Errata.Println(err)
		return
	}
	g.addressed[reg.Addr] = reg.Id
	g.lbl[regi] = MakeLabel(MyId, reg.Id)
	g.cb[regi] = cb
	g.gcm[regi] = gcm
	xlog.Info.Println("peered w/", reg.Name, reg.Id, reg.Addr)
	g.pendingTx(ctx, reg)
	g.pendingRx(ctx, reg)
}

func (g *guest) rx(ctx context.Context, m *xnet.Msg) error {
	var err error

	put := true
	defer func() {
		if put {
			mp.Put(m)
		}
	}()

	to, from := ScanLabel(m.Data)
	if to == from {
		g.verifyHello(m, from)
		return nil
	}
	if to != MyId {
		xlog.Trace.Print("dropped ", to, "<-", from)
		return nil
	}
	if !g.isOK(from) {
		put = false
		g.pending.rx = append(g.pending.rx, m)
		g.whoisQueue(ctx, from)
		return nil
	}
	ifrom := from.Index()
	cb, ok := g.cb[ifrom]
	if !ok {
		xlog.Trace.Println(from, "cipher-block not found")
		return nil
	}
	cb.Decrypt(m.Data[:aes.BlockSize], m.Data[:aes.BlockSize])
	gcm, ok := g.gcm[ifrom]
	if !ok {
		xlog.Trace.Println(from, "gcm not found")
		return nil
	}
	i := len(m.Data) - SizeofLabel
	m.Data, err = gcm.Open(m.Data[:0], nil, m.Data[:i], m.Data[i:])
	if err != nil {
		xlog.Trace.Println("open:", err)
		return nil
	}
	xlog.Trace.Print("rx ", to, "<-", from, netpdu.Mark, PDU(m.Data))
	_, err = g.tun.Write(m.Data)
	return err
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

func (g *guest) tx(ctx context.Context, m *xnet.Msg) error {
	put := true
	defer func() {
		if put {
			mp.Put(m)
		}
	}()
	da, err := netpdu.TunPI(m.Data).ToWhom()
	if err != nil {
		xlog.Trace.Println("dropped no tunPI")
		return nil
	}
	if !da.IsValid() {
		xlog.Trace.Println("dropped non-ip[6]")
		return nil
	}
	if da.IsMulticast() {
		xlog.Trace.Println("dropped multicast", da)
		return nil
	}
	if da.Compare(g.prefix.Addr()) == 0 {
		xlog.Trace.Println("loopback", da)
		g.tun.Write(m.Data)
		return nil
	}
	if g.llu6.IsValid() && da.Compare(g.llu6) == 0 {
		xlog.Trace.Println("link-local loopback", da)
		g.tun.Write(m.Data)
		return nil
	}
	if !g.vpn.Contains(da) {
		xlog.Trace.Println(da, "out of", g.vpn)
		return nil
	}
	to, ok := g.addressed[da]
	if !ok {
		put = false
		g.pending.tx = append(g.pending.tx, m)
		g.whoisQueue(ctx, da)
		return nil
	}

	pdu := netpdu.TunPI(m.Data)
	pdu.Proto(da.Is6())
	xlog.Trace.Print("tx ", to, "<-", MyId, netpdu.Mark, pdu)

	ito := to.Index()
	lbl, ok := g.lbl[ito]
	if !ok {
		xlog.Trace.Println(to, "label not found")
		return nil
	}
	gcm, ok := g.gcm[ito]
	if !ok {
		xlog.Trace.Println(to, "gcm not found")
		return nil
	}
	cb, ok := g.cb[ito]
	if !ok {
		xlog.Trace.Println(to, "cipher-block not found")
		return nil
	}
	m.Data = gcm.Seal(m.Data[:0], nil, m.Data, lbl)
	cb.Encrypt(m.Data[:aes.BlockSize], m.Data[:aes.BlockSize])
	m.Data = append(m.Data, lbl...)
	m.AddrPort = g.via.Service
	return udpSend(ctx, m)
}
