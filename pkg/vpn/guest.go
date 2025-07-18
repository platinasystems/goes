// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"net/netip"
	"os"
	"os/signal"
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
	"github.com/platinasystems/goes/v2/pkg/xsignal"
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

var guest struct {
	addressed map[netip.Addr]*Subscriber
	indexed   map[int]*Subscriber
	via       map[string]*Subscriber

	llu6 netip.Addr

	// msgs pending whois response
	pending struct{ rx, tx []*xnet.Msg }

	// FIXME replace w/ MLKEM
	priv *ecdh.PrivateKey

	receipt *GuestReceipt

	tun  *os.File
	tunC chan *xnet.Msg
}

// Guest is a UDP server that forwards ciphered packets between an exchange
// and a network tunnel interface.
func Guest(ctx context.Context, args []string) error {
	defer xlog.Info.Println("stopped")

	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
Forward ciphered packets between exchange and tunnel interface.

{{flags .}}`)

	defineRestFlags()
	defineTunnel()
	enableQuiet()
	enableTrace()
	enableVerbose()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctx)

	if err = restInit(); err != nil {
		return err
	}
	defer close(rest.whoisReqC)

	guest.addressed = make(map[netip.Addr]*Subscriber)
	guest.indexed = make(map[int]*Subscriber)
	guest.via = make(map[string]*Subscriber)

	guest.priv, err = ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	pubder, err := x509.MarshalPKIXPublicKey(guest.priv.PublicKey())
	if err != nil {
		return err
	}
	pubpem := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubder,
	})
	guest.receipt, err = restGuestCheckin(ctx, pubpem)
	if err != nil {
		return err
	}

	if len(guest.receipt.ExchangePrecedence) > RestWhoisDepth {
		guest.receipt.ExchangePrecedence =
			guest.receipt.ExchangePrecedence[:RestWhoisDepth-2]
	}
	for _, name := range append(guest.receipt.ExchangePrecedence,
		rest.reg.Subject.CommonName) {
		guest.via[name] = nil
	}

	vpnPrefix = guest.receipt.Prefix.Masked()

	ha := netif.NewHardwareAddr()
	if err = ha.Rand(); err != nil {
		return err
	}

	if err = udpInit(0); err != nil {
		return err
	}
	defer udp.Close()

	guestHelloOrWhoisAllVias(ctx, time.Now().UnixMicro())

	guest.tunC = make(chan *xnet.Msg, 1)
	guest.tun, err = nettun.
		New(vpnTunnel, IsTap, TunPersist, TunOwner, TunGroup, ha)
	if err != nil {
		return err
	}
	defer guest.tun.Close()

	// FIXME do we ever want non-blocking reads or rely on stdlib?
	if false && runtime.NumCPU() == 1 {
		if err = xos.SetNonblockFile(guest.tun, true); err != nil {
			return err
		}
	}

	nif, err := netif.Named(ctx, guest.tun.Name())
	if err != nil {
		return err
	}
	if nif == nil {
		return xerrors.NotFound(guest.tun.Name())
	}

	if err = nif.Config(ctx, "up", "mtu", fmt.Sprint(TunMTU)); err != nil {
		return err
	}
	xlog.Info.Println(nif.Name, "up", "mtu", TunMTU)

	addr := guest.receipt.Prefix.Addr()
	if addr.Is6() {
		addr = addr.WithZone(nif.Name)
	}
	bits := 32
	if addr.Is6() {
		bits = 128
	}

	// FIXME may need dst w/ ipv4 VPN
	if err = nif.Add(ctx, addr, netip.Addr{}, bits); err != nil {
		return err
	}
	xlog.Info.Println(nif.Name, addr)

	if err = routeAdd(ctx, vpnPrefix, nif); err != nil {
		return err
	}
	xlog.Info.Println(nif.Name, vpnPrefix)
	defer routeDelete(ctx, vpnPrefix, nif)

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
				guest.llu6 = pa
				break
			}
		}
	}

	tkr := time.NewTicker(10 * time.Second)
	defer tkr.Stop()

	alarm := make(chan os.Signal, 2)
	signal.Notify(alarm, xsignal.Alarm)
	defer signal.Stop(alarm)

	wg.Go(func() { restWhoisService(ctx) })
	wg.Go(guestTunStream)
	wg.Go(udpStream)

	svc := fmt.Sprintf("%s %v@%v via %v",
		rest.crt.Subject.CommonName,
		MyId, udp.LocalAddr(),
		guest.receipt.ExchangePrecedence)
	xlog.Info.Println("start", svc)
	defer cancel()
	defer xlog.Trace.Println("stopping", svc, "...")

selection:
	for err == nil {
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case <-alarm:
			xlog.Info = xlog.ToggleMute(xlog.Info)
			xlog.Trace = xlog.Mute(xlog.Trace)
		case err = <-rest.fault:
		case t := <-tkr.C:
			guestHelloOrWhoisAllVias(ctx, t.UnixMicro())
		case sub, ok := <-rest.whoisRspC:
			if !ok {
				break selection
			}
			if _, ok = guest.via[sub.name()]; ok {
				guestExchangeWith(ctx, sub)
			} else {
				guestPeerWith(ctx, sub)
			}
		case m, ok := <-guest.tunC:
			if !ok {
				break selection
			}
			guestFromTun(ctx, m)
		case m, ok := <-udpC:
			if !ok {
				break selection
			}
			guestFromUDP(ctx, m)
		}
	}
	return xerrors.Suppress(err, context.Canceled, errEOC)
}

func guestExchangeWith(ctx context.Context, sub *Subscriber) {
	guest.via[sub.name()] = sub
	guest.addressed[sub.Addr] = sub
	guest.indexed[sub.Id.Index()] = sub
	if err := sub.resolve(ctx); err != nil {
		xlog.Errata.Print(err)
	}
	sub.hello(ctx, 0)
}

func guestFromTun(ctx context.Context, m *xnet.Msg) {
	da, err := netpdu.TunPI(m.Data).ToWhom()
	if err != nil {
		xlog.Errata.Println("dropped:", err)
	} else if !da.IsValid() {
		xlog.Trace.Println("dropped non-ip[6]")
	} else if da.IsMulticast() {
		xlog.Trace.Println("dropped multicast", da)
	} else if da.Compare(guest.receipt.Prefix.Addr()) == 0 {
		xlog.Trace.Println("loopback", da)
		guest.tun.Write(m.Data)
	} else if guest.llu6.IsValid() && da.Compare(guest.llu6) == 0 {
		xlog.Trace.Println("link-local loopback", da)
		guest.tun.Write(m.Data)
	} else if !vpnPrefix.Contains(da) {
		xlog.Trace.Println("dropped ", da, "out of", vpnPrefix)
	} else if to, ok := guest.addressed[da]; ok {
		guestTx(ctx, to, m)
	} else {
		restQueueWhois(ctx, da)
		xlog.Trace.Println("queue to", da, "pending whois")
		guest.pending.tx = append(guest.pending.tx, m)
		return
	}
	mp.Put(m)
}

func guestFromUDP(ctx context.Context, m *xnet.Msg) {
	tid, fid := ScanLabel(m.Data)
	fi := fid.Index()
	from, ok := guest.indexed[fi]
	if !ok || from.Id.Version() != fid.Version() {
		guest.pending.rx = append(guest.pending.rx, m)
		restQueueWhois(ctx, fid)
		xlog.Trace.Println("queue from", fid, "pending whois")
		return
	}
	if tid == fid {
		if from.helloIsOK(ctx, m) {
			from.via = m.AddrPort
		}
	} else if tid != MyId {
		xlog.Trace.Println("dropped", tid, "<-", fid)
	} else {
		guestRx(ctx, from, m)
	}
	mp.Put(m)
}

func guestHelloOrWhoisAllVias(ctx context.Context, now int64) {
	if now == 0 {
		now = time.Now().UnixMicro()
	}
	for name, sub := range guest.via {
		if sub == nil || !sub.via.IsValid() {
			restQueueWhois(ctx, name)
		} else {
			sub.hello(ctx, now)
		}
	}
}

func guestDiscardPending(sub *Subscriber) {
	guestMatchPendingRx(sub.Id, func(m *xnet.Msg) {
		mp.Put(m)
	})
	guestMatchPendingTx(sub.Addr, func(m *xnet.Msg) {
		mp.Put(m)
	})
}

func guestMatchPendingRx(id Id, f func(*xnet.Msg)) {
	var m *xnet.Msg
	for i := 0; i < len(guest.pending.rx); {
		m = guest.pending.rx[i]
		_, from := ScanLabel(m.Data)
		if id.Index() == from.Index() {
			guest.pending.rx = slices.
				Delete(guest.pending.rx, i, i+1)
			f(m)
		} else {
			i += 1
		}
	}
}

func guestMatchPendingTx(addr netip.Addr, f func(*xnet.Msg)) {
	for i := 0; i < len(guest.pending.tx); {
		m := guest.pending.tx[i]
		da, err := netpdu.TunPI(m.Data).ToWhom()
		if err == nil && da.Compare(addr) == 0 {
			guest.pending.tx = slices.
				Delete(guest.pending.tx, i, i+1)
			f(m)
		} else {
			i += 1
		}
	}
}

func guestPeerWith(ctx context.Context, sub *Subscriber) {
	sub.label.fromMe = MakeLabel(MyId, sub.Id)
	sub.label.toMe = MakeLabel(sub.Id, MyId)
	i := sub.Id.Index()

	// FIXME replace w/ MLKEM
	if len(sub.CipherKeyDER) == 0 {
		guestDiscardPending(sub)
		xlog.Errata.Print("cipher key")
		return
	}
	k, err := x509.ParsePKIXPublicKey(sub.CipherKeyDER)
	if err != nil {
		guestDiscardPending(sub)
		xlog.Errata.Print(err)
		return
	}
	ecdhpub, ok := k.(*ecdh.PublicKey)
	if !ok {
		guestDiscardPending(sub)
		xlog.Errata.Printf("unsupported cipher key: %T", k)
		return
	}
	shared, err := guest.priv.ECDH(ecdhpub)
	if err != nil {
		guestDiscardPending(sub)
		xlog.Errata.Print(err)
		return
	}

	if sub.cb, err = aes.NewCipher(shared); err != nil {
		guestDiscardPending(sub)
		xlog.Errata.Print(err)
		return
	}
	if sub.gcm, err = cipher.NewGCMWithRandomNonce(sub.cb); err != nil {
		guestDiscardPending(sub)
		xlog.Errata.Print(err)
		return
	}
	guest.addressed[sub.Addr] = sub
	guest.indexed[i] = sub
	// First see if there are msgs pending this whois response.
	// If there are, then the received remote AddrPort is the
	// preferred subscriber exchange.
	if len(guest.pending.rx) > 0 {
		guestMatchPendingRx(sub.Id, func(m *xnet.Msg) {
			sub.via = m.AddrPort
			guestRx(ctx, sub, m)
			mp.Put(m)
		})
	}
	// If there weren't any pending rx msgs, use the best matching
	// exchange or, in last resort, the registry.
	if !sub.via.IsValid() {
		for _, name := range sub.ExchangePrecedence {
			if x, ok := guest.via[name]; ok {
				if x.via.IsValid() {
					sub.via = x.via
					break
				} else {
					xlog.Trace.Println("unaddressed", name)
				}
			}
		}
		if !sub.via.IsValid() {
			sub.via = guest.via[rest.reg.Subject.CommonName].via
		}
	}
	if len(guest.pending.tx) > 0 {
		guestMatchPendingTx(sub.Addr, func(m *xnet.Msg) {
			guestTx(ctx, sub, m)
			mp.Put(m)
		})
	}
}

func guestRx(ctx context.Context, from *Subscriber, m *xnet.Msg) {
	var err error

	name := from.name()
	from.cb.Decrypt(m.Data[:aes.BlockSize], m.Data[:aes.BlockSize])
	i := len(m.Data) - SizeofLabel
	m.Data, err = from.gcm.Open(m.Data[:0], nil, m.Data[:i], m.Data[i:])
	if err != nil {
		xlog.Trace.Print("from ", name, netpdu.Mark, err)
	} else if _, err = guest.tun.Write(m.Data); err != nil {
		xlog.Trace.Print("from ", name, netpdu.Mark, err)
	} else {
		xlog.Trace.Print("from ", name, netpdu.Mark, PDU(m.Data))
	}
}

func guestTunRead(ctx context.Context) {
	name := guest.tun.Name()
	xlog.Trace.Println("start", name, "read")
	defer xlog.Trace.Println("stopped", name, "read")
	defer close(guest.tunC)
	m := mp.Get()
	defer mp.Put(m)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		m.Data = m.Data[:cap(m.Data)]
		n, err := guest.tun.Read(m.Data)
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
		if da.IsMulticast() {
			xlog.Trace.Println("dropped multicast", da)
		} else if da.Compare(guest.receipt.Prefix.Addr()) == 0 {
			xlog.Trace.Println("loopback", da)
			guest.tun.Write(m.Data)
		} else if guest.llu6.IsValid() && da.Compare(guest.llu6) == 0 {
			xlog.Trace.Println("link-local loopback", da)
			guest.tun.Write(m.Data)
		} else if !vpnPrefix.Contains(da) {
			xlog.Trace.Println(da, "out of", vpnPrefix)
			continue
		} else {
			select {
			case <-ctx.Done():
				mp.Put(m)
				return
			case guest.tunC <- m:
				m = mp.Get()
			}
		}
	}
}

func guestTunStream() {
	xlog.Trace.Println("start", guest.tun.Name(), "stream")
	err := mp.StreamReader(guest.tunC, guest.tun)
	err = xerrors.Suppress(err, fs.ErrClosed)
	if err != nil {
		xlog.Errata.Println("quit", guest.tun.Name(), "stream", err)
	} else {
		xlog.Trace.Println("stopped", guest.tun.Name(), "stream")
	}

}

func guestTx(ctx context.Context, to *Subscriber, m *xnet.Msg) {
	name := to.name()
	pdu := PDU(m.Data)
	// FIXME pdu.Proto(m.AddrPort.Addr().Is6())
	if to.via.IsValid() {
		xlog.Trace.Print("to ", name, netpdu.Mark, pdu)
		m.Data = to.gcm.Seal(m.Data[:0], nil, m.Data, to.label.fromMe)
		to.cb.Encrypt(m.Data[:aes.BlockSize], m.Data[:aes.BlockSize])
		m.Data = append(m.Data, to.label.fromMe...)
		udp.WriteToUDPAddrPort(m.Data, to.via)
	} else {
		xlog.Trace.Print("dropped w/o path to ", name, netpdu.Mark, pdu)
	}
}
