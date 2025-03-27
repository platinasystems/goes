// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"crypto/rand"
	"encoding/pem"
	"flag"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/box"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netpdu"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

type exchangeWhoisResponse struct {
	blk *pem.Block
	err error
}

// Exchange is a UDP server that forwards ciphered packets between guest's.
func Exchange(ctx context.Context, args []string) error {
	const defport = 8003
	var wg sync.WaitGroup
	var ex exchange
	var pub netip.AddrPort

	xlog.SetPrefixes("exchange/")

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] [vpn]
Exchange ciphered packets between guests.

{{flags .}}`)

	flag.TextVar(&pub, NamePublicFlag,
		netip.AddrPortFrom(netip.IPv4Unspecified(), 0),
		`NAT'd listen {addr}:{port}. (0.0.0.0:0 ignored)`)
	trace := flag.Bool(NameTraceFlag, false, "Log packet forwarding.")
	err := ex.defineAndParseFlags(ctx, defport, args)
	if err != nil {
		return err
	}
	if *trace {
		xlog.UnmuteTrace()
	}

	ex.pem.addressed = make(map[netip.Addr]*pem.Block)
	ex.pem.identified = make(map[int]*pem.Block)

	ex.pktRxCh = make(chan *box.Box, 4)
	ex.pktTxCh = make(chan *box.Box, 4)

	ex.whoisResponseCh = make(chan exchangeWhoisResponse)

	if ex.lladdr, err = randLinkLocalAddr(); err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(ctx)

	udp, err := net.ListenUDP(ex.udpv, &net.UDPAddr{
		IP:   ex.lap.Addr().AsSlice(),
		Port: int(ex.lap.Port()),
	})
	if err != nil {
		return xerrors.Label(err, "ListenUDP")
	}

	defer udp.Close()

	lap, err := netip.ParseAddrPort(udp.LocalAddr().String())
	if err != nil {
		return xerrors.Label(err, "LocalAddr")
	}
	sap := lap
	if !pub.Addr().IsUnspecified() {
		sap = pub
	}
	if err = ex.register(ctx, sap); err != nil {
		return err
	}

	iex, vex := IdIndex(ex.id), IdVersion(ex.id)

	svc := fmt.Sprintf("%d@%v", iex, sap)
	xlog.Info.Println("start", svc)
	defer xlog.Info.Println("stopped", svc)
	defer wg.Wait()
	defer cancel()
	defer xlog.Info.Println("stopping", svc, "...")

	wg.Add(1)
	go xlog.AlarmHandler(cctx, &wg)
	wg.Add(1)
	go pktRxRoutine(cctx, &wg, udp, ex.pktRxCh)
	wg.Add(1)
	go pktTxRoutine(cctx, &wg, udp, ex.pktTxCh)
	defer close(ex.pktTxCh)

pktRxLoop:
	for {
		select {
		case <-cctx.Done():
			xlog.Info.Println("done", cctx.Err())
			break pktRxLoop
		case bx, ok := <-ex.pktRxCh:
			if !ok {
				xlog.Info.Println("pkt rx ch closed")
				break pktRxLoop
			}
			afrom := bx.AddrPort
			from := bx.FromWhom()
			ifrom, vfrom := IdIndex(from), IdVersion(from)
			via := bx.ViaWhom()
			ivia, vvia := IdIndex(via), IdVersion(via)
			if ivia != iex {
				xlog.Info.Printf("via %d != exchange %d",
					ivia, iex)
				bx.Return()
				continue pktRxLoop
			}
			if vvia != vex {
				xlog.Info.Printf("FIXME version via "+
					"%d != exchange %d", vvia, vex)
				bx.Return()
				continue pktRxLoop
			}
			if vfrom != ex.ver[ifrom] {
				wg.Add(1)
				go ex.whoisIdRoutine(cctx, &wg, from)
				bx.Return()
				continue pktRxLoop
			}
			ex.service[ifrom] = afrom
			cfrom, ok := ex.gcm[ifrom]
			if !ok {
				xlog.Info.Printf("whois %d@%v", ifrom, afrom)
				wg.Add(1)
				go ex.whoisIdRoutine(cctx, &wg, from)
				bx.Return()
				continue pktRxLoop
			}
			if err := bx.UnsealWith(cfrom); err != nil {
				xlog.Info.Printf("%d@%v unseal %v",
					ifrom, afrom, err)
				bx.Return()
				continue pktRxLoop
			}
			to := bx.ToWhom()
			ito, vto := IdIndex(to), IdVersion(to)
			if ito == iex {
				if vto != vex {
					xlog.Info.Printf("%d@%v "+
						"FIXME prompt exchange update",
						ifrom, afrom)
					bx.Return()
				} else {
					err = bx.OpenWith(cfrom)
					if err != nil {
						xlog.Info.Printf("%d@%v "+
							"open %v",
							ifrom, afrom, err)
						bx.Return()
					} else {
						ex.rx(cctx, &wg, bx)
					}
				}
			} else if vto != ex.ver[ito] {
				xlog.Info.Printf("%d@%v "+
					"FIXME prompt host %d update",
					ifrom, afrom, ito)
				bx.Return()
			} else if ato, ok := ex.service[ito]; !ok {
				xlog.Info.Printf("%d@%v no service to %d",
					ifrom, afrom, ito)
				bx.Return()
			} else if cto, ok := ex.gcm[ito]; !ok {
				xlog.Info.Printf("%d@%v not peered with %d",
					ifrom, afrom, ito)
				bx.Return()
			} else {
				xlog.Info.Printf("%d@%v reseal and send to "+
					"%d@%v", ifrom, afrom, ito, ato)
				bx.AddrPort = ato
				bx.SealWith(cto)
				bx.NonBlockingPut(ex.pktTxCh)
			}
		case rsp, ok := <-ex.whoisResponseCh:
			if !ok {
				xlog.Info.Println("closed whois response ch")
				break pktRxLoop
			}
			if rsp.err != nil {
				xlog.Errata.Print(rsp.err)
			} else if err = ex.whoisResponse(rsp.blk); err != nil {
				xlog.Info.Println(err)
			}
		}
	}
	return nil
}

type exchange struct {
	client
	lladdr netip.Addr
	pktRxCh,
	pktTxCh chan *box.Box
	pem struct {
		addressed  map[netip.Addr]*pem.Block
		identified map[int]*pem.Block
	}
	whoisResponseCh chan exchangeWhoisResponse
	all6nodes,
	all6routers netip.Addr
}

func (ex *exchange) rx(
	ctx context.Context,
	wg *sync.WaitGroup,
	bx *box.Box,
) {
	defer bx.Return()
	ex.all6nodes = netip.IPv6LinkLocalAllNodes()
	ex.all6routers = netip.IPv6LinkLocalAllRouters()
	afrom := bx.AddrPort
	from := bx.FromWhom()
	ifrom := IdIndex(from)
	pdu := VpnPDU(bx.Contents)
	h, d, err := pdu.Parse()
	if err != nil {
		xlog.Errata.Print(err)
		return
	}
	xlog.Info.Printf("rx %d@%v %v", ifrom, afrom, pdu)
	switch h.Proto {
	case VPN_P_HELLO:
		ex.txHelloAck(from)
	case VPN_P_WHOIS_ADDRESSED:
		if addr := VpnWhoisAddress(d); !addr.IsValid() {
			xlog.Errata.Printf("rx %d@%v whois underrun; %v",
				ifrom, afrom, pdu)
		} else if blk := ex.pem.addressed[addr]; blk == nil {
			xlog.Info.Printf("rx %d@%v whois %v",
				ifrom, afrom, addr)
			wg.Add(1)
			go ex.whoisAddressedRoutine(ctx, wg, addr)
		} else {
			ex.txPubKey(from, blk)
		}
	case VPN_P_WHOIS_IDENTIFIED:
		if id, err := VpnWhoisId(d); err != nil {
			xlog.Errata.Printf("rx %d@%v whois %v",
				ifrom, afrom, err)
		} else if blk := ex.pem.identified[IdIndex(id)]; blk == nil {
			xlog.Info.Printf("rx %d@%v whois %d",
				ifrom, afrom, id)
			wg.Add(1)
			go ex.whoisIdRoutine(ctx, wg, id)
		} else {
			ex.txPubKey(from, blk)
		}
	case VPN_P_WHOIS_SERVICE:
		xlog.Info.Printf("rx %d@%v whois %v", ifrom, afrom,
			VpnWhoisService(d))
	case VPN_P_IP:
		xlog.Info.Printf("dropped %d@%v %v", ifrom, afrom, netpdu.IP(d))
	case VPN_P_IP6:
		ex.rxIP6(from, netpdu.IP6(d))
	default:
		xlog.Info.Printf("rx %d@%v unknown %#x", ifrom, afrom, h.Proto)
	}
}

func (ex *exchange) rxIP6(from box.Id, pdu netpdu.IP6) {
	h, d, err := pdu.Parse()
	if err != nil {
		xlog.Info.Print(err)
		return
	}
	sum := pdu.Checksum(h.NextHeader, d)
	if sum != 0 {
		xlog.Info.Print("bad sum")
		return
	}
	switch h.NextHeader {
	case netph.IPPROTO_ICMPV6:
		ex.rxICMP6(from, &h, netpdu.ICMP6(d))
	case netph.IPPROTO_UDP:
		if da := netip.AddrFrom16(h.DA); ex.tome(da) {
			ex.rxUDP6(from, &h, netpdu.UDP(d))
		} else {
			xlog.Info.Println("dropped udp to", da)
		}
	default:
		xlog.Info.Println("dropped", pdu)
	}
}

func (ex *exchange) rxICMP6(
	from box.Id,
	req6 *netph.IP6,
	pdu netpdu.ICMP6,
) {
	switch pdu.Type() {
	case netph.ICMP6TypeEchoRequest:
		if da := netip.AddrFrom16(req6.DA); ex.tome(da) {
			req := netpdu.ICMP6EchoRequest(pdu)
			ex.txICMP6EchoReply(from, req6, req)
		} else {
			xlog.Info.Println("dropped icmp6 echo req to", da)
		}
	case netph.ICMP6TypeRouterSolicitation:
		ex.txICMP6RouterAdvertisement(from, req6)
	}
}

func (ex *exchange) rxUDP6(
	from box.Id,
	req6 *netph.IP6,
	pdu netpdu.UDP,
) {
	h, d, err := pdu.Parse()
	if err != nil {
		xlog.Info.Print(err)
		return
	}
	switch h.DP {
	case netph.UDPEcho:
		ex.txUDP6EchoReply(from, req6, &h, d)
	case netph.UDPDomain:
		// FIXME
	default:
		xlog.Info.Println("dropped", pdu)
	}
}

func (ex *exchange) tome(a netip.Addr) bool {
	return a.Compare(ex.lladdr) == 0 || a.Compare(ex.addr) == 0
}

func (ex *exchange) txHelloAck(to box.Id) {
	var err error
	bx := box.New()
	toi := IdIndex(to)
	bx.AddrPort = ex.service[toi]
	bx.From(ex.id)
	bx.To(to)
	bx.Contents, err = xnet.Add(bx.Contents, netph.TunPI{
		Proto: VPN_P_HELLO,
	})
	if err != nil {
		bx.Return()
		return
	}
	bx.Contents, err = xnet.Add(bx.Contents, time.Now().UnixMicro())
	if err != nil {
		bx.Return()
		return
	}
	xlog.Info.Printf("tx %d@%v hello ack", toi, bx.AddrPort)
	c := ex.gcm[toi]
	bx.CloseWith(c)
	bx.SealWith(c)
	bx.NonBlockingPut(ex.pktTxCh)
}

func (ex *exchange) txICMP6EchoReply(
	to box.Id,
	req6 *netph.IP6,
	req netpdu.ICMP6EchoRequest,
) {
	toi := IdIndex(to)
	reqh, reqd, err := req.Parse()
	if err != nil {
		return
	}
	bx := box.New()
	bx.AddrPort = ex.service[toi]
	bx.From(ex.id)
	bx.To(to)
	bx.Contents, err = xnet.Add(bx.Contents, netph.TunPI{
		Proto: netph.ETH_P_IPV6,
	})
	if err != nil {
		bx.Return()
		return
	}
	ip6i := len(bx.Contents)
	bx.Contents, err = xnet.Add(bx.Contents, netph.IP6{
		VCF:        req6.VCF,
		NextHeader: netph.IPPROTO_ICMPV6,
		HopLimit:   255,
		SA:         req6.DA,
		DA:         req6.SA,
	})
	if err != nil {
		bx.Return()
		return
	}
	icmp6i := len(bx.Contents)
	bx.Contents, err = xnet.Add(bx.Contents, netph.ICMP6EchoReply{
		ICMP6: netph.ICMP6{
			Type: netph.ICMP6TypeEchoReply,
			Code: 0,
		},
		Identifier: reqh.Identifier,
		Sequence:   reqh.Sequence,
	})
	if err != nil {
		bx.Return()
		return
	}
	bx.Contents, err = xnet.Add(bx.Contents, reqd)
	if err != nil {
		bx.Return()
		return
	}
	ip6 := netpdu.IP6(bx.Contents[ip6i:])
	ip6.SetLen()
	icmp6 := netpdu.ICMP6(bx.Contents[icmp6i:])
	sum := ip6.Checksum(netph.IPPROTO_ICMPV6, icmp6)
	icmp6.SetSum(sum)
	xlog.Info.Printf("tx %d@%v %v", toi, bx.AddrPort, ip6)
	c := ex.gcm[toi]
	bx.CloseWith(c)
	bx.SealWith(c)
	bx.NonBlockingPut(ex.pktTxCh)
}

func (ex *exchange) txICMP6RouterAdvertisement(to box.Id, req6 *netph.IP6) {
	const infiniteLifetime = 0xffffffff
	var err error
	toi := IdIndex(to)
	bx := box.New()
	bx.AddrPort = ex.service[toi]
	bx.From(ex.id)
	bx.To(to)
	bx.Contents, err = xnet.Add(bx.Contents, netph.TunPI{
		Proto: netph.ETH_P_IPV6,
	})
	if err != nil {
		bx.Return()
		return
	}
	ip6i := len(bx.Contents)
	bx.Contents, err = xnet.Add(bx.Contents, netph.IP6{
		VCF:        req6.VCF,
		NextHeader: netph.IPPROTO_ICMPV6,
		HopLimit:   255, // 1 ir 2 ?
		SA:         ex.lladdr.As16(),
		DA:         req6.SA,
	})
	if err != nil {
		bx.Return()
		return
	}
	icmp6i := len(bx.Contents)
	bx.Contents, err = xnet.Add(bx.Contents, netph.ICMP6RouterAdvertisement{
		ICMP6: netph.ICMP6{
			Type: netph.ICMP6TypeRouterAdvertisement,
			Code: 0,
		},
		CurHopLimit: 2, // ?
		// netph.ICMP6RouterAdvertisementOtherConfiguration |
		// netph.ICMP6RouterAdvertisementManagedAddress
		Flags: 0,
		// not default
		RouterLifetime: 0,
		// unspecified
		ReachableTime: 0,
		// unspecified
		RetransTimer: 0,
	})
	if err != nil {
		bx.Return()
		return
	}
	prefixopti := len(bx.Contents)
	bx.Contents, err = xnet.Add(bx.Contents, netph.ICMP6PrefixInformation{
		ICMP6Option: netph.ICMP6Option{
			Type:   netph.ICMP6OptionTypePrefixInformation,
			Length: 0, // updated later
		},
		// FIXME 96 or 98 for ID derrived address?
		PrefixLength: 64,
		Flags: netph.ICMP6PrefixAutonomousAddressConfiguration |
			netph.ICMP6PrefixOnLink,
		ValidLifetime:     infiniteLifetime,
		PreferredLifetime: infiniteLifetime,
		Prefix:            netip.MustParseAddr("fc00:5678::").As16(),
	})
	bx.Contents[prefixopti+netph.ICMP6OptionLengthIndex] =
		uint8(len(bx.Contents) - prefixopti)
	// FIXME add MTU, RDNSS and DNSSL options
	ip6 := netpdu.IP6(bx.Contents[ip6i:])
	ip6.SetLen()
	icmp6 := netpdu.ICMP6(bx.Contents[icmp6i:])
	icmp6.SetSum(ip6.Checksum(netph.IPPROTO_ICMPV6, icmp6))
	xlog.Info.Printf("tx %d@%v %v", toi, bx.AddrPort, ip6)
	c := ex.gcm[toi]
	bx.CloseWith(c)
	bx.SealWith(c)
	bx.NonBlockingPut(ex.pktTxCh)
}

func (ex *exchange) txPubKey(to box.Id, blk *pem.Block) {
	var err error
	toi := IdIndex(to)
	bx := box.New()
	bx.AddrPort = ex.service[toi]
	bx.From(ex.id)
	bx.To(to)
	bx.Contents, err = xnet.Add(bx.Contents, netph.TunPI{
		Proto: VPN_P_PUBLIC_KEY,
	})
	if err != nil {
		bx.Return()
		return
	}
	pem.Encode(bx, blk)
	xlog.Info.Printf("tx %d@%v pub-key", toi, bx.AddrPort)
	c := ex.gcm[toi]
	bx.CloseWith(c)
	bx.SealWith(c)
	bx.NonBlockingPut(ex.pktTxCh)
}

func (ex *exchange) txUDP6EchoReply(
	to box.Id,
	req6 *netph.IP6,
	requdp *netph.UDP,
	data []byte,
) {
	var err error
	toi := IdIndex(to)
	bx := box.New()
	bx.From(ex.id)
	bx.To(to)
	bx.AddrPort = ex.service[toi]
	bx.Contents, err = xnet.Add(bx.Contents, netph.TunPI{
		Proto: netph.ETH_P_IPV6,
	})
	if err != nil {
		bx.Return()
		return
	}
	ip6i := len(bx.Contents)
	bx.Contents, err = xnet.Add(bx.Contents, netph.IP6{
		VCF:        req6.VCF,
		NextHeader: netph.IPPROTO_UDP,
		HopLimit:   255,
		SA:         req6.DA,
		DA:         req6.SA,
	})
	if err != nil {
		bx.Return()
		return
	}
	udpi := len(bx.Contents)
	bx.Contents, err = xnet.Add(bx.Contents, netph.UDP{
		SP: requdp.DP,
		DP: requdp.SP,
	})
	if err != nil {
		bx.Return()
		return
	}
	bx.Contents, err = xnet.Add(bx.Contents, data)
	if err != nil {
		bx.Return()
		return
	}
	ip6 := netpdu.IP6(bx.Contents[ip6i:])
	ip6.SetLen()
	udp := netpdu.UDP(bx.Contents[udpi:])
	udp.SetLen()
	udp.SetSum(ip6.Checksum(netph.IPPROTO_UDP, udp))
	xlog.Info.Printf("tx %d@%v %v", toi, bx.AddrPort, ip6)
	c := ex.gcm[toi]
	bx.CloseWith(c)
	bx.SealWith(c)
	bx.NonBlockingPut(ex.pktTxCh)
}

func (ex *exchange) whoisAddressedRoutine(
	ctx context.Context, wg *sync.WaitGroup, addr netip.Addr,
) {
	defer wg.Done()
	blk, err := ex.whoisAddressed(ctx, addr)
	if err != nil {
		err = fmt.Errorf("%w (whois %v)", err, addr)
	}
	ex.whoisResponseCh <- exchangeWhoisResponse{blk, err}
}

func (ex *exchange) whoisIdRoutine(
	ctx context.Context, wg *sync.WaitGroup, id box.Id,
) {
	defer wg.Done()
	blk, err := ex.whoisIdentified(ctx, id)
	if err != nil {
		err = fmt.Errorf("%w (whois %d)", err, IdIndex(id))
	}
	ex.whoisResponseCh <- exchangeWhoisResponse{blk, err}
}

func (ex *exchange) whoisResponse(blk *pem.Block) error {
	addr, err := addressHeader(blk)
	if err != nil {
		return xerrors.Label(err, "HeaderAddress")
	}
	id, err := idHeader(blk)
	if err != nil {
		return xerrors.Label(err, "HeaderId")
	}
	idi := IdIndex(id)
	if err = ex.peer(blk); err != nil {
		return err
	}
	ex.pem.addressed[addr] = blk
	ex.pem.identified[idi] = blk
	return nil
}

func randLinkLocalAddr() (lladdr netip.Addr, err error) {
	var a [netph.IPv6len]byte
	a[0] = 0xfe
	a[1] = 0x80
	n, err := rand.Read(a[8:])
	if err != nil {
	} else if n != len(a[8:]) {
		err = ErrLLAddrUnderrun
	} else {
		lladdr = netip.AddrFrom16(a)
		xlog.Info.Println("link-local address:", lladdr)
	}
	return
}
