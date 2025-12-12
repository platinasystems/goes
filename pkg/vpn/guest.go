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
	"strconv"
	"time"

	"github.com/platinasystems/goes/v2/pkg/cert"
	"github.com/platinasystems/goes/v2/pkg/netif"
	"github.com/platinasystems/goes/v2/pkg/nettun"
	"github.com/platinasystems/goes/v2/pkg/sig"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netpdu"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
	"github.com/platinasystems/goes/v2/pkg/xsignal"
)

const (
	GCMNonceSize = 12
	GCMTagSize   = 16
	GCMOverhead  = GCMNonceSize + GCMTagSize

	VPNMTU  = netph.ETHMTU - netph.TunPISize
	IP6MTU  = VPNMTU - netph.IP6Size
	UDP6MTU = IP6MTU - netph.UDPSize
	TunMTU  = UDP6MTU - GCMOverhead - SizeofLabel

	IsTap = false

	TunOwner = -1
	TunGroup = -1

	TunPersist = false

	noReplyLimit = 3

	minWhoisRetryTicks = 6
)

const GuestUsage = `
usage: {{.Name}} [flags]
Forward ciphered packets between exchange and tunnel interface.

{{flags .}}`

var (
	Guest_t = -1
	Guest_x string
)

var GuestFlags = xflag.Labels{
	xlog.TraceFlag,
	xlog.VerboseFlag,

	xmain.ConfigFlag,
	cert.ClientFlag,
	cert.ServerFlag,
	sig.Flag,

	RestCertFlag,
	RestServerFlag,
	RestPortFlag,

	{"t", `
Tunnel unit number, auto selected if negative.
(or $<main>_TUNNEL_UNIT)`[1:], func() *int {
		if s, ok := xmain.LookupEnv("TUNNEL_UNIT"); ok {
			if i, err := strconv.ParseInt(s, 10, 32); err == nil {
				Guest_t = int(i)
			}
		}
		return &Guest_t
	}},
	{"x", `
Comma separated list of preferred exchanges.
(or $<main>_GUEST_EXCHANGES)`[1:], func() *string {
		if s, ok := xmain.LookupEnv("GUEST_EXCHANGES"); ok {
			Guest_x = s
		}
		return &Guest_x
	}},
}

var guest struct {
	addressed map[netip.Addr]*Subscriber
	indexed   map[int]*Subscriber
	exchanges []*Subscriber

	//xnet.ByteOrder
	revbuf []byte

	tick uint64

	llu6 netip.Addr

	// msg pending whois response
	pending struct{ rx, tx *xnet.Msg }

	decapKey *mlkem.DecapsulationKey768

	receipt *GuestReceipt

	tun   *os.File
	tunpi []byte

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

	xflag.TemplateUsage(GuestUsage)
	err := GuestFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}

	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctx)

	if err = restInit(); err != nil {
		return err
	}
	defer close(rest.reqC)

	guest.addressed = make(map[netip.Addr]*Subscriber)
	guest.indexed = make(map[int]*Subscriber)

	if err = restVcsCheck(ctx); err != nil {
		return err
	}

	guest.decapKey, err = mlkem.GenerateKey768()
	if err != nil {
		return err
	}
	encapKey := guest.decapKey.EncapsulationKey()
	guest.receipt, err = RestCheckinGuestReq(ctx, encapKey.Bytes(), Guest_x)
	if err != nil {
		return err
	}

	xlog.Info.Printf("receipt.ExchangePrecedence: %q",
		guest.receipt.ExchangePrecedence)

	// the registry is always the exchange of last resort
	guest.exchanges = make([]*Subscriber, 0,
		1+len(guest.receipt.ExchangePrecedence))
	regname := rest.reg.Subject.CommonName
	x, err := RestWhoisReq(ctx, regname)
	if err != nil {
		return fmt.Errorf("%s: %w", regname, err)
	}
	// derrive registry's exchange service AddrPort
	regaddr, _ := netip.AddrFromSlice(rest.ips[0])
	if regaddr.Is4In6() {
		regaddr = regaddr.Unmap()
	}
	if x.Port == 0 {
		x.Port = DefaultExchangePort
	}
	x.ap = netip.AddrPortFrom(regaddr, x.Port)
	guest.exchanges = append(guest.exchanges, x)
	guest.addressed[x.Addr] = x
	guest.indexed[x.Id.Index()] = x

	for _, name := range guest.receipt.ExchangePrecedence {
		if len(name) == 0 {
			xlog.Errata.Println("skip missing exchange name")
			continue
		}
		x, err = RestWhoisReq(ctx, name)
		if err != nil {
			guest.exchanges = append(guest.exchanges, nil)
			xlog.Errata.Printf("%s: %w", name, err)
		} else {
			guest.exchanges = append(guest.exchanges, x)
			guest.addressed[x.Addr] = x
			guest.indexed[x.Id.Index()] = x
			if x.resolve(ctx); !x.ap.IsValid() {
				xlog.Errata.Println("unresolved", x)
			}
		}
	}

	prefix = guest.receipt.Prefix.Masked()

	ha := netif.NewHardwareAddr()
	if err = ha.Rand(); err != nil {
		return err
	}

	guest.fromVpnC, guest.toVpnC, err = startUDP(ctx, 0)
	if err != nil {
		return err
	}
	defer close(guest.toVpnC)

	guestHelloToAllExchanges(ctx, 0)

	guest.tun, err = nettun.New(Guest_t, IsTap, TunPersist, TunOwner,
		TunGroup, ha)
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

	if err = routeAdd(ctx, prefix, nif); err != nil {
		return err
	}
	xlog.Info.Println(nif.Name, prefix)
	defer routeDelete(ctx, prefix, nif)

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

	wg.Go(func() { restReqService(ctx) })

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
			xlog.Info.Toggle()
			xlog.Trace.Mute()
		case t := <-tkr.C:
			guest.tick += 1
			if (guest.tick & 1) == 1 {
				guestHelloToAllExchanges(ctx, t.UnixMicro())
			}
			switch guest.tick & 7 {
			case 4:
				restQueueVcsCheck(ctx)
			case 6:
				guestReqIdRev(ctx)
			}
		case v, ok := <-rest.rspC:
			if !ok {
				break selection
			} else if err, ok = v.(error); ok {
				xlog.Errata.Print(err)
				if !needsRestart(err) {
					err = nil
				}
			} else if b, ok := v.([]byte); ok {
				guestRevIds(b)
			} else if sub, ok := v.(*Subscriber); ok {
				guestFound(ctx, sub)
			} else {
				xlog.Errata.Printf("invalid response: %T", v)
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
			xlog.Trace.Println("peered w/", sub)
			if m := guest.pending.tx; m != nil {
				da, err := netpdu.TunPI(m.Data).ToWhom()
				if err == nil && da.Compare(sub.Addr) == 0 {
					guest.pending.tx = nil
					guestFromTun(ctx, m)
				}
			}
			if m := guest.pending.rx; m != nil {
				_, from := ScanLabel(m.Data)
				if sub.Id.Index() == from.Index() {
					guest.pending.rx = nil
					guestFromVpn(ctx, m)
				}
			}
		}
	}
	return xerrors.Suppress(err, context.Canceled, errEOC)
}

func guestDiscardPending(sub *Subscriber) {
	if m := guest.pending.rx; m != nil {
		_, from := ScanLabel(m.Data)
		if sub.Id.Index() == from.Index() {
			guest.pending.rx = nil
			mp.Put(m)
		}
	}
	if m := guest.pending.tx; m != nil {
		da, err := netpdu.TunPI(m.Data).ToWhom()
		if err == nil && da.Compare(sub.Addr) == 0 {
			guest.pending.tx = nil
			mp.Put(m)
		}
	}
}

func guestNewPendingTx(m *xnet.Msg) {
	if x := guest.pending.tx; x != nil {
		mp.Put(x)
	}
	guest.pending.tx = m
}

// returns non-zero indexed exchange or the registry if the indexed exchange
// hasn't replied w/in the noReplyLimit
func guestExchange(i int) *Subscriber {
	if i > 0 && i < len(guest.exchanges) {
		if x := guest.exchanges[i]; x != nil {
			if x.ticks(guest.tick) < noReplyLimit {
				return x
			}
		}
	}
	return guest.exchanges[0]
}

// use first matching exchange or, in last resort, the registry.
func guestExchangeMatch(sub *Subscriber) {
	sub.gxi = 0
	for _, name := range sub.ExchangePrecedence {
		for i, x := range guest.exchanges {
			if x != nil && x.name() == name {
				sub.gxi = i
				return
			}
		}
	}
}

func guestExchangeWith(ctx context.Context, x *Subscriber) {
	name := x.name()
	if name == rest.reg.Subject.CommonName {
		guest.exchanges[0] = x
	} else {
		for i, xname := range guest.receipt.ExchangePrecedence {
			if name == xname {
				guest.exchanges[1+i] = x
				break
			}
		}
	}
	guest.addressed[x.Addr] = x
	guest.indexed[x.Id.Index()] = x
	if x.resolve(ctx); !x.ap.IsValid() {
		xlog.Errata.Println(name, "unresolved")
		return
	}
	if hello := NewGreeting(0); hello != nil {
		xlog.Trace.Println("hello", x)
		hello.AddrPort = x.ap
		mp.Queue(ctx, guest.toVpnC, hello)
	}
}

func guestFound(ctx context.Context, sub *Subscriber) {
	var cipherText []byte

	if sub == nil {
		xlog.Errata.Print("nil subscriber")
		return
	}

	name := sub.name()

	if name == rest.reg.Subject.CommonName {
		guestExchangeWith(ctx, sub)
		return
	}
	for _, xname := range guest.receipt.ExchangePrecedence {
		if name == xname {
			guestExchangeWith(ctx, sub)
			return
		}
	}

	guestExchangeMatch(sub)

	sub.label.fromMe = MakeLabel(MyId, sub.Id)
	sub.label.toMe = MakeLabel(sub.Id, MyId)
	guest.indexed[sub.Id.Index()] = sub
	guest.addressed[sub.Addr] = sub

	xlog.Trace.Println(name, "via", guest.exchanges[sub.gxi])

	if len(sub.EncapKey) == 0 {
		xlog.Errata.Println(name, "unregistered")
		sub.lt = guest.tick
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

	wg.Go(func() {
		invite, err := RestInviteReq(ctx, name, cipherText)
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
	if len(guest.tunpi) == 0 {
		guest.tunpi = make([]byte, netph.TunPISize)
		copy(guest.tunpi, m.Data)
		xlog.Info.Printf("tunpi: %x", guest.tunpi)
	}
	pdu := netpdu.TunPI(m.Data)
	da, err := pdu.ToWhom()
	if err != nil {
		xlog.Errata.Println("dropped:", err)
		mp.Put(m)
	} else if !da.IsValid() {
		xlog.Trace.Println("dropped", pdu)
		mp.Put(m)
	} else if m.AddrPort = netip.AddrPortFrom(da, 0); da.IsMulticast() {
		xlog.Trace.Println("dropped", pdu)
		mp.Put(m)
	} else if da.Compare(guest.receipt.Prefix.Addr()) == 0 {
		xlog.Trace.Println("loopback", pdu)
		mp.Queue(ctx, guest.toTunC, m)
	} else if guest.llu6.IsValid() && da.Compare(guest.llu6) == 0 {
		xlog.Trace.Println("loopback", pdu)
		mp.Queue(ctx, guest.toTunC, m)
	} else if !prefix.Contains(da) {
		xlog.Trace.Println("dropped", pdu)
		mp.Put(m)
	} else if to, ok := guest.addressed[da]; !ok {
		guestNewPendingTx(m)
		restQueueWhois(ctx, da)
	} else if len(to.EncapKey) == 0 {
		if to.lt == 0 || to.ticks(guest.tick) > minWhoisRetryTicks {
			to.lt = guest.tick
			guestNewPendingTx(m)
			restQueueWhois(ctx, da)
		} else {
			mp.Put(m)
		}
	} else if to.gcm == nil {
		xlog.Trace.Println("pending invite", da)
		guestNewPendingTx(m)
	} else {
		m.AddrPort = guestExchange(to.gxi).ap
		xlog.Trace.Print("tx ", to.name(), " via ", m.AddrPort,
			netpdu.Mark, pdu)
		m.Data = to.gcm.Seal(m.Data[:0], nil, m.Data, to.label.fromMe)
		to.cb.Encrypt(m.Data[:aes.BlockSize], m.Data[:aes.BlockSize])
		m.Data = append(m.Data, to.label.fromMe...)
		mp.Queue(ctx, guest.toVpnC, m)
	}
}

func guestFromVpn(ctx context.Context, m *xnet.Msg) {
	tid, fid := ScanLabel(m.Data)
	fi := fid.Index()
	from, ok := guest.indexed[fi]
	if !ok || from.Id.Version() != fid.Version() {
		if cur := guest.pending.rx; cur != nil {
			mp.Put(cur)
		}
		guest.pending.rx = m
		xlog.Trace.Println("queue whois", fid)
		restQueueWhois(ctx, fid)
	} else if tid == fid {
		if err := from.helloCheck(m); err == nil {
			xlog.Trace.Println("hello from", from)
			from.lt = guest.tick
			from.ap = unmap4in6(m.AddrPort)
			if from.ap.Port() == 0 {
				from.ap = netip.AddrPortFrom(from.ap.Addr(),
					DefaultExchangePort)
			}
		} else {
			xlog.Errata.Println(err)
		}
		mp.Put(m)
	} else if tid != MyId {
		xlog.Trace.Println("dropped from", from.name(), "to", tid)
		mp.Put(m)
	} else {
		guestRx(ctx, from, m)
	}
}

func guestHelloToAllExchanges(ctx context.Context, now int64) {
	if now == 0 {
		now = time.Now().UnixMicro()
	}
	for i, x := range guest.exchanges {
		if x == nil {
			var name string
			names := guest.receipt.ExchangePrecedence
			if i == 0 {
				name = rest.reg.Subject.CommonName
			} else if i > len(names) {
				xlog.Errata.Println("index", i, "exceeds",
					names)
				continue
			} else {
				name = guest.receipt.ExchangePrecedence[i-1]
				if len(name) == 0 {
					xlog.Errata.Println("no name at", i,
						"of", names)
					continue
				}
			}
			xlog.Trace.Println("queue whois", name)
			restQueueWhois(ctx, name)
			continue
		}
		if !x.ap.IsValid() {
			if x.resolve(ctx); !x.ap.IsValid() {
				xlog.Errata.Println("unresolved", x)
				continue
			}
		}
		if x.ap.Port() == 0 {
			x.ap = netip.AddrPortFrom(x.ap.Addr(),
				DefaultExchangePort)
		}
		if hello := NewGreeting(now); hello != nil {
			xlog.Trace.Println("hello to", x)
			hello.AddrPort = x.ap
			if !mp.Queue(ctx, guest.toVpnC, hello) {
				break
			}
		}
	}
}

func guestReqIdRev(ctx context.Context) {
	var err error
	n := len(guest.indexed)
	if n == 0 {
		return
	}
	if len(guest.revbuf) == 0 {
		guest.revbuf = make([]byte, 0, n*SizeofId)
	} else {
		guest.revbuf = guest.revbuf[:0]
	}
	for _, sub := range guest.indexed {
		guest.revbuf, err = xnet.ByteOrderAppend(guest.revbuf, sub.Id)
		if err != nil {
			xlog.Errata.Print(err)
		}
	}
	xlog.Trace.Println("req revision of", n, "peer(s)")
	restQueueReviseIds(ctx, guest.revbuf)
}

func guestRevIds(b []byte) {
	n := len(b) / SizeofId
	if n == 0 {
		xlog.Trace.Print("no revised peers")
		return
	}
	for len(b) >= SizeofId {
		var id Id
		var err error
		b, err = xnet.ByteOrderRemove(b, &id)
		if err != nil {
			xlog.Errata.Print(err)
			break
		}
		if sub, ok := guest.indexed[id.Index()]; !ok {
			xlog.Errata.Println(id, xerrors.ErrNotFound)
		} else if sub.Id.Version() != id.Version() {
			xlog.Trace.Println("must update", sub.name())
			sub.EncapKey = sub.EncapKey[:0]
			sub.sharedKey = sub.sharedKey[:0]
			sub.cb = nil
			sub.gcm = nil
			sub.lt = guest.tick
		}
	}
	xlog.Trace.Println("revised", n, "peer(s)")
}

func guestRx(ctx context.Context, from *Subscriber, m *xnet.Msg) {
	var err error
	defer mp.Put(m)

	name := from.name()
	if from.cb == nil || from.gcm == nil {
		xlog.Trace.Println("dropped ", name, "w/o handshake")
		return
	}
	from.cb.Decrypt(m.Data[:aes.BlockSize], m.Data[:aes.BlockSize])
	i := len(m.Data) - SizeofLabel
	m.Data, err = from.gcm.Open(m.Data[:0], nil, m.Data[:i], m.Data[i:])
	if err != nil {
		xlog.Trace.Print("rx ", name, netpdu.Mark, err)
		return
	}
	if len(guest.tunpi) > 0 {
		copy(m.Data, guest.tunpi)
	}
	ap := unmap4in6(m.AddrPort)
	pdu := netpdu.TunPI(m.Data)
	if _, err = guest.tun.Write(m.Data); err != nil {
		xlog.Errata.Print("rx ", name, " via ", ap, netpdu.Mark, err)
	} else {
		xlog.Trace.Print("rx ", name, " via ", ap, netpdu.Mark, pdu)
	}
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
