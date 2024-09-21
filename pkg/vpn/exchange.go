// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
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
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/box"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netpdu"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

// Exchange is a UDP server that forwards ciphered packets between guest's.
func Exchange(ctx context.Context, args []string) error {
	const defport = 8003
	var wg sync.WaitGroup
	var ex exchange

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] [vpn]
Exchange ciphered packets between guests.

{{flags .}}`)

	nflag := NatFlag()

	ex.xFlag = ExchangeFlag()
	err := ex.flags(ctx, defport, args)
	if err != nil {
		return err
	}

	ex.pem.addressed = make(map[netip.Addr]*pem.Block)
	ex.pem.identified = make(map[int]*pem.Block)

	ex.pktRxCh = make(chan *box.Box, 4)
	ex.pktTxCh = make(chan *box.Box, 4)

	ex.whoisResponseCh = make(chan *pem.Block)

	cctx, cancel := context.WithCancel(ctx)

	udp, err := net.ListenUDP(ex.udpv, &net.UDPAddr{
		IP:   ex.sap.Addr().AsSlice(),
		Port: int(ex.sap.Port()),
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
	if !(*nflag).Addr().IsUnspecified() {
		sap = *nflag
	}
	if err = ex.register(ctx, sap); err != nil {
		return err
	}

	iex, vex := IdIndex(ex.id), IdVersion(ex.id)

	svc := fmt.Sprintf("%d@%v", iex, sap)
	goRoutineTrace.Println("start", svc)
	defer goRoutineTrace.Println("stopped", svc)
	defer wg.Wait()
	defer cancel()
	defer goRoutineTrace.Println("stopping", svc, "...")

	wg.Add(1)
	go pktRxRoutine(cctx, &wg, udp, ex.pktRxCh)
	wg.Add(1)
	go pktTxRoutine(cctx, &wg, udp, ex.pktTxCh)
	defer close(ex.pktTxCh)

pktRxLoop:
	for {
		select {
		case <-cctx.Done():
			goRoutineTrace.Println("done", cctx.Err())
			break pktRxLoop
		case bx, ok := <-ex.pktRxCh:
			if !ok {
				goRoutineTrace.Println("pkt rx ch closed")
				break pktRxLoop
			}
			afrom := bx.AddrPort
			from := bx.FromWhom()
			ifrom, vfrom := IdIndex(from), IdVersion(from)
			via := bx.ViaWhom()
			ivia, vvia := IdIndex(via), IdVersion(via)
			if ivia != iex {
				verbose.Printf("via %d != exchange %d",
					ivia, iex)
				bx.Return()
				continue pktRxLoop
			}
			if vvia != vex {
				verbose.Printf("FIXME version via "+
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
				verbose.Printf("whois %d@%v", ifrom, afrom)
				wg.Add(1)
				go ex.whoisIdRoutine(cctx, &wg, from)
				bx.Return()
				continue pktRxLoop
			}
			if err := bx.UnsealWith(cfrom); err != nil {
				verbose.Printf("%d@%v unseal %v",
					ifrom, afrom, err)
				bx.Return()
				continue pktRxLoop
			}
			to := bx.ToWhom()
			ito, vto := IdIndex(to), IdVersion(to)
			if ito == iex {
				if vto != vex {
					verbose.Printf("%d@%v "+
						"FIXME prompt exchange update",
						ifrom, afrom)
					bx.Return()
				} else {
					err = bx.OpenWith(cfrom)
					if err != nil {
						verbose.Printf("%d@%v "+
							"open %v",
							ifrom, afrom, err)
						bx.Return()
					} else {
						ex.rx(cctx, &wg, bx)
					}
				}
			} else if vto != ex.ver[ito] {
				verbose.Printf("%d@%v "+
					"FIXME prompt host %d update",
					ifrom, afrom, ito)
				bx.Return()
			} else if ato, ok := ex.service[ito]; !ok {
				verbose.Printf("%d@%v no service to %d",
					ifrom, afrom, ito)
				bx.Return()
			} else if cto, ok := ex.gcm[ito]; !ok {
				verbose.Printf("%d@%v not peered with %d",
					ifrom, afrom, ito)
				bx.Return()
			} else {
				verbose.Printf("%d@%v reseal and send to "+
					"%d @ %v", ifrom, afrom, ito, ato)
				bx.AddrPort = ato
				bx.SealWith(cto)
				bx.NonBlockingPut(ex.pktTxCh)
			}
		case blk, ok := <-ex.whoisResponseCh:
			if !ok {
				verbose.Println("closed whois response ch")
				break pktRxLoop
			}
			if err = ex.whoisResponse(blk); err != nil {
				verbose.Println(err)
			}
		}
	}
	return nil
}

type exchange struct {
	client
	pktRxCh,
	pktTxCh chan *box.Box
	pem struct {
		addressed  map[netip.Addr]*pem.Block
		identified map[int]*pem.Block
	}
	whoisResponseCh chan *pem.Block
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
	pi, d, err := netpdu.TunPI(bx.Contents).Parse()
	if err != nil {
		verbose.Println("rx %d@%v pi %v", ifrom, afrom, err)
		return
	}
	switch pi.Proto {
	case VPN_P_HELLO:
		verbose.Printf("rx %d@%v hello %v", ifrom, afrom,
			VpnHelloPDU(d))
		ex.txHelloAck(ex.pktTxCh, from)
	case VPN_P_WHOIS_ADDRESSED:
		if addr := VpnWhoisAddress(d); !addr.IsValid() {
			verbose.Printf("rx %d@%v whois underrun", ifrom, afrom)
		} else if blk := ex.pem.addressed[addr]; blk == nil {
			verbose.Printf("rx %d@%v whois %v", ifrom, afrom,
				addr)
			wg.Add(1)
			go ex.whoisAddressedRoutine(ctx, wg, addr)
		} else {
			ex.txPubKey(from, blk)
		}
	case VPN_P_WHOIS_IDENTIFIED:
		if id, err := VpnWhoisId(d); err != nil {
			verbose.Printf("rx %d@%v whois %v", ifrom, afrom, err)
		} else if blk := ex.pem.identified[IdIndex(id)]; blk == nil {
			verbose.Printf("rx %d@%v whois %d", ifrom, afrom, id)
			wg.Add(1)
			go ex.whoisIdRoutine(ctx, wg, id)
		} else {
			ex.txPubKey(from, blk)
		}
	case VPN_P_WHOIS_SERVICE:
		verbose.Printf("rx %d@%v whois %v", ifrom, afrom,
			VpnWhoisService(d))
	case netph.ETH_P_IP:
		ip := netpdu.IP(d)
		verbose.Printf("rx %d@%v %v", ifrom, afrom, ip)
		ex.rxIP(from, ip)
	case netph.ETH_P_IPV6:
		ip6 := netpdu.IP6(d)
		verbose.Printf("rx %d@%v %v", ifrom, afrom, ip6)
		ex.rxIP6(from, ip6)
	default:
		verbose.Printf("rx %d@%v unknown %#x", ifrom, afrom, pi.Proto)
	}
}

func (ex *exchange) rxIP(from box.Id, pdu netpdu.IP) {
	h, _, err := pdu.Parse()
	if err != nil {
		return
	}
	da := netip.AddrFrom4(h.DA)
	if h.Protocol == netph.IPPROTO_ICMP {
		if da.IsMulticast() || da == ex.addr {
			// FIXME respond
		} else {
			// FIXME route
		}
	} else if da.IsMulticast() {
		// FIXME multicast
	} else if da != ex.addr {
		// FIXME route
	} else {
		// // ignore
	}
}

func (ex *exchange) rxIP6(from box.Id, pdu netpdu.IP6) {
	h, d, err := pdu.Parse()
	if err != nil {
		return
	}
	da := netip.AddrFrom16(h.DA)
	if h.NextHeader == netph.IPPROTO_ICMPV6 {
		if da.IsMulticast() || da == ex.addr {
			ex.rxICMP6(from, h.SA, pdu, netpdu.ICMP6(d))
		} else {
			// FIXME route
		}
	} else if da.IsMulticast() {
		// FIXME multicast
	} else if da != ex.addr {
		// FIXME route
	} else {
		// ignore
	}
}

func (ex *exchange) rxICMP6(
	from box.Id,
	sa [netph.IPv6len]byte,
	ip6 netpdu.IP6,
	icmp6 netpdu.ICMP6,
) {
	sum := ip6.Checksum(netph.IPPROTO_ICMPV6, icmp6)
	if sum != 0 {
		verbose.Print("bad sum")
		return
	}
	switch icmp6.Type() {
	case netph.ICMP6TypeEchoRequest:
		ex.txICMP6EchoReply(from, sa, netpdu.ICMP6EchoRequest(icmp6))
	case netph.ICMP6TypeRouterSolicitation:
		// FIXME router-advertisement
	}
}

func (ex *exchange) txHelloAck(ch chan<- *box.Box, to box.Id) {
	var err error
	bx := box.New()
	i := IdIndex(to)
	c := ex.gcm[i]
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
	ap := ex.service[i]
	verbose.Printf("tx %d@%v hello ack", i, ap)
	bx.AddrPort = ap
	bx.From(ex.id)
	bx.To(to)
	bx.CloseWith(c)
	bx.SealWith(c)
	bx.NonBlockingPut(ch)
}

func (ex *exchange) txPubKey(to box.Id, blk *pem.Block) {
	var err error
	i := IdIndex(to)
	c := ex.gcm[i]
	bx := box.New()
	bx.Contents, err = xnet.Add(bx.Contents, netph.TunPI{
		Proto: VPN_P_PUBLIC_KEY,
	})
	if err != nil {
		bx.Return()
		return
	}
	pem.Encode(bx, blk)
	ap := ex.service[i]
	verbose.Printf("tx %d@%v pub-key", i, ap)
	bx.AddrPort = ap
	bx.From(ex.id)
	bx.To(to)
	bx.CloseWith(c)
	bx.SealWith(c)
	bx.NonBlockingPut(ex.pktTxCh)
}

func (ex *exchange) txICMP6EchoReply(
	to box.Id,
	da [netph.IPv6len]byte,
	req netpdu.ICMP6EchoRequest,
) {
	i := IdIndex(to)
	ap := ex.service[i]
	c := ex.gcm[i]
	reqh, reqd, err := req.Parse()
	if err != nil {
		return
	}
	bx := box.New()
	bx.Contents, err = xnet.Add(bx.Contents, netph.TunPI{
		Proto: netph.ETH_P_IPV6,
	})
	if err != nil {
		bx.Return()
		return
	}
	class := uint8(0)
	flow := uint32(0)
	ip6i := len(bx.Contents)
	bx.Contents, err = xnet.Add(bx.Contents, netph.IP6{
		VCF:        netph.ConstructVCF(class, flow),
		LEN:        0, // updated after appending icmp6
		NextHeader: netph.IPPROTO_ICMPV6,
		HopLimit:   255,
		SA:         ex.addr.As16(),
		DA:         da,
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
			Sum:  0, // updated after appending reply
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
	verbose.Printf("tx %d@%v %v", i, ap, ip6)
	bx.AddrPort = ap
	bx.From(ex.id)
	bx.To(to)
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
		verbose.Printf("whois %v: %v", addr, err)
	} else {
		ex.whoisResponseCh <- blk
	}
}

func (ex *exchange) whoisIdRoutine(
	ctx context.Context, wg *sync.WaitGroup, id box.Id,
) {
	defer wg.Done()
	blk, err := ex.whoisIdentified(ctx, id)
	if err != nil {
		verbose.Printf("whois %v: %v", IdIndex(id), err)
	} else {
		ex.whoisResponseCh <- blk
	}
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
