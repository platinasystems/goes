// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/mlkem"
	"flag"
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/signal"
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

	decapKey *mlkem.DecapsulationKey768

	receipt *GuestReceipt

	tun *os.File

	newPeerC chan *Subscriber

	fromTunC chan *xnet.Msg
	toTunC   chan *xnet.Msg

	fromVpnC <-chan *xnet.Msg
	toVpnC   chan<- *xnet.Msg
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

	guest.decapKey, err = mlkem.GenerateKey768()
	if err != nil {
		return err
	}
	encapKey := guest.decapKey.EncapsulationKey()
	guest.receipt, err = restGuestCheckin(ctx, encapKey.Bytes())
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

	guest.fromVpnC, guest.toVpnC, err = startUDP(ctx, 0)
	if err != nil {
		return err
	}
	defer close(guest.toVpnC)

	guestHelloOrWhoisAllVias(ctx, time.Now().UnixMicro())

	guest.tun, err = nettun.
		New(vpnTunnel, IsTap, TunPersist, TunOwner, TunGroup, ha)
	if err != nil {
		return err
	}
	defer guest.tun.Close()

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

	guestStartTunneling(ctx)
	defer close(guest.toTunC)

	xlog.Info.Println("start guest", MyId)
	defer cancel()
	defer xlog.Trace.Println("stopping guest", MyId, "...")

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
			if x, ok := guest.via[sub.name()]; !ok {
				guestFound(ctx, sub)
			} else {
				if x != nil {
					xlog.Trace.Println("update",
						x.name(), sub.via)
				}
				guestExchangeWith(ctx, sub)
			}
		case m, ok := <-guest.fromTunC:
			if !ok {
				break selection
			}
			guestFromTun(ctx, m)
		case m, ok := <-guest.fromVpnC:
			if !ok {
				break selection
			}
			guestFromVpn(ctx, m)
		case sub, ok := <-guest.newPeerC:
			if !ok {
				break selection
			}
			for i := 0; i < len(guest.pending.tx); {
				m := guest.pending.tx[i]
				da, err := netpdu.TunPI(m.Data).ToWhom()
				if err == nil && da.Compare(sub.Addr) == 0 {
					guest.pending.tx = slices.
						Delete(guest.pending.tx, i, i+1)
					guestFromTun(ctx, m)
				} else {
					i += 1
				}
			}
			for i := 0; i < len(guest.pending.rx); {
				m := guest.pending.rx[i]
				_, from := ScanLabel(m.Data)
				if sub.Id.Index() == from.Index() {
					guest.pending.rx = slices.
						Delete(guest.pending.rx, i, i+1)
					guestFromVpn(ctx, m)
				} else {
					i += 1
				}
			}
		}
	}
	return xerrors.Suppress(err, context.Canceled, errEOC)
}

func guestDiscardPending(sub *Subscriber) {
	for i := 0; i < len(guest.pending.rx); {
		m := guest.pending.rx[i]
		_, from := ScanLabel(m.Data)
		if sub.Id.Index() == from.Index() {
			guest.pending.rx = slices.
				Delete(guest.pending.rx, i, i+1)
			mp.Put(m)
		} else {
			i += 1
		}
	}
	for i := 0; i < len(guest.pending.tx); {
		m := guest.pending.tx[i]
		da, err := netpdu.TunPI(m.Data).ToWhom()
		if err == nil && da.Compare(sub.Addr) == 0 {
			guest.pending.tx = slices.
				Delete(guest.pending.tx, i, i+1)
			mp.Put(m)
		} else {
			i += 1
		}
	}
}

func guestExchangeWith(ctx context.Context, sub *Subscriber) {
	guest.via[sub.name()] = sub
	guest.addressed[sub.Addr] = sub
	guest.indexed[sub.Id.Index()] = sub
	sub.resolve(ctx)
	if sub.via.IsValid() {
		if hello := newGreeting(0); hello != nil {
			hello.AddrPort = sub.via
			mp.Queue(ctx, guest.toVpnC, hello)
		}
	}
}

func guestFound(ctx context.Context, sub *Subscriber) {
	var cipherText []byte

	name := sub.name()

	if len(sub.EncapKey) == 0 {
		xlog.Errata.Println(name, "missing cipher key")
		guestDiscardPending(sub)
		return
	}
	encap, err := mlkem.NewEncapsulationKey768(sub.EncapKey)
	if err != nil {
		xlog.Errata.Println(name, "bad encap key:", err)
		guestDiscardPending(sub)
		return
	}
	sub.sharedKey, cipherText = encap.Encapsulate()

	// use best matching exchange or, in last resort, the registry.
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

	sub.label.fromMe = MakeLabel(MyId, sub.Id)
	sub.label.toMe = MakeLabel(sub.Id, MyId)
	guest.indexed[sub.Id.Index()] = sub
	guest.addressed[sub.Addr] = sub

	xlog.Trace.Println(name, "via", sub.via)

	wg.Go(func() {
		invite, err := restInvite(ctx, name, cipherText)
		if err != nil {
			xlog.Errata.Println(name, "invite:", err)
			guestDiscardPending(sub)
			return
		}
		if len(invite) > 0 {
			sub.sharedKey, err = guest.decapKey.Decapsulate(invite)
			if err != nil {
				xlog.Errata.Println(name, "decap:", err)
				guestDiscardPending(sub)
				return
			}
		}
		if sub.cb, err = aes.NewCipher(sub.sharedKey); err != nil {
			xlog.Errata.Println(name, "shared cipher:", err)
			guestDiscardPending(sub)
			return
		}
		sub.gcm, err = cipher.NewGCMWithRandomNonce(sub.cb)
		if err != nil {
			xlog.Errata.Println(name, "gcm:", err)
			guestDiscardPending(sub)
			return
		}
		guest.newPeerC <- sub
	})
}

func guestFromTun(ctx context.Context, m *xnet.Msg) {
	tunpi := netpdu.TunPI(m.Data)
	da, err := tunpi.ToWhom()
	if err != nil {
		xlog.Errata.Println("dropped:", err)
		mp.Put(m)
	} else if !da.IsValid() {
		xlog.Trace.Println("dropped", tunpi)
		mp.Put(m)
	} else if m.AddrPort = netip.AddrPortFrom(da, 0); da.IsMulticast() {
		xlog.Trace.Println("dropped", tunpi)
		mp.Put(m)
	} else if da.Compare(guest.receipt.Prefix.Addr()) == 0 {
		xlog.Trace.Println("loopback", tunpi)
		mp.Queue(ctx, guest.toTunC, m)
	} else if guest.llu6.IsValid() && da.Compare(guest.llu6) == 0 {
		xlog.Trace.Println("loopback", tunpi)
		mp.Queue(ctx, guest.toTunC, m)
	} else if !vpnPrefix.Contains(da) {
		xlog.Trace.Println("dropped", tunpi)
		mp.Put(m)
	} else if to, ok := guest.addressed[da]; !ok {
		xlog.Trace.Println("queue whois", da)
		restQueueWhois(ctx, da)
		guest.pending.tx = append(guest.pending.tx, m)
	} else if to.gcm == nil {
		xlog.Trace.Println("pending invite", da)
		guest.pending.tx = append(guest.pending.tx, m)
	} else if !to.via.IsValid() {
		xlog.Trace.Println("dropped", to.name(), netpdu.Mark, tunpi)
		mp.Put(m)
	} else {
		vpn := PDU(m.Data)
		// FIXME vpn.Proto(m.AddrPort.Addr().Is6())
		xlog.Trace.Print("tx ", to.name(), netpdu.Mark, vpn)
		m.Data = to.gcm.Seal(m.Data[:0], nil, m.Data, to.label.fromMe)
		to.cb.Encrypt(m.Data[:aes.BlockSize], m.Data[:aes.BlockSize])
		m.Data = append(m.Data, to.label.fromMe...)
		m.AddrPort = to.via
		mp.Queue(ctx, guest.toVpnC, m)
	}
}

func guestFromVpn(ctx context.Context, m *xnet.Msg) {
	tid, fid := ScanLabel(m.Data)
	fi := fid.Index()
	from, ok := guest.indexed[fi]
	if !ok || from.Id.Version() != fid.Version() {
		guest.pending.rx = append(guest.pending.rx, m)
		xlog.Trace.Println("queue whois", fid)
		restQueueWhois(ctx, fid)
	} else if tid == fid {
		if from.helloIsOK(m) {
			from.via = m.AddrPort
		}
		mp.Put(m)
	} else if tid != MyId {
		xlog.Trace.Println("dropped from", from.name(), "to", tid)
		mp.Put(m)
	} else {
		guestRx(ctx, from, m)
	}
}

func guestHelloOrWhoisAllVias(ctx context.Context, now int64) {
	if now == 0 {
		now = time.Now().UnixMicro()
	}
	for name, x := range guest.via {
		if x == nil {
			xlog.Trace.Println("queue whois", name)
			restQueueWhois(ctx, name)
		} else if !x.via.IsValid() {
			xlog.Trace.Println("re-ask whois", name, x)
			restQueueWhois(ctx, name)
		} else if hello := newGreeting(0); hello != nil {
			hello.AddrPort = x.via
			if !mp.Queue(ctx, guest.toVpnC, hello) {
				break
			}
		}
	}
}

func guestRx(ctx context.Context, from *Subscriber, m *xnet.Msg) {
	var err error

	name := from.name()
	if from.cb == nil || from.gcm == nil {
		xlog.Trace.Println("dropped ", name, "w/o handshake")
		mp.Put(m)
		return
	}
	from.cb.Decrypt(m.Data[:aes.BlockSize], m.Data[:aes.BlockSize])
	i := len(m.Data) - SizeofLabel
	m.Data, err = from.gcm.Open(m.Data[:0], nil, m.Data[:i], m.Data[i:])
	if err != nil {
		xlog.Trace.Print("rx ", name, netpdu.Mark, err)
	} else if _, err = guest.tun.Write(m.Data); err != nil {
		xlog.Trace.Print("rx ", name, netpdu.Mark, err)
	} else {
		xlog.Trace.Print("rx ", name, netpdu.Mark, PDU(m.Data))
		from.via = m.AddrPort
	}
	mp.Put(m)
}

func guestStartTunneling(ctx context.Context) {
	guest.newPeerC = make(chan *Subscriber, 1)
	guest.fromTunC = make(chan *xnet.Msg, FromTunCap)
	guest.toTunC = make(chan *xnet.Msg, ToTunCap)
	name := guest.tun.Name() + " read service"
	wg.Go(func() {
		const kind = "read service"
		xlog.Trace.Println("start", name, kind)
		err := mp.ReadService(guest.fromTunC, guest.tun)
		if err != nil {
			xlog.Errata.Println("quit", name, kind, err)
		} else {
			xlog.Trace.Println("stopped", name, kind)
		}
	})
	wg.Go(func() {
		const kind = "write service"
		xlog.Trace.Println("start", name, kind)
		err := mp.WriteService(guest.tun, guest.toTunC)
		if err != nil {
			xlog.Errata.Println("quit", name, kind, err)
		} else {
			xlog.Trace.Println("stopped", name, kind)
		}
	})
}
