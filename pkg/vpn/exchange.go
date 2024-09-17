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

	svc := fmt.Sprintf("%d @ %v", iex, sap)
	verbose.Println("start", svc)
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
			verbose.Println("done", cctx.Err())
			break pktRxLoop
		case bx, ok := <-ex.pktRxCh:
			if !ok {
				verbose.Println("pkt rx ch closed")
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
				verbose.Printf("%d @ %v who?", ifrom, afrom)
				wg.Add(1)
				go ex.whoisIdRoutine(cctx, &wg, from)
				bx.Return()
				continue pktRxLoop
			}
			if err := bx.UnsealWith(cfrom); err != nil {
				verbose.Printf("%d @ %v unseal %v",
					ifrom, afrom, err)
				bx.Return()
				continue pktRxLoop
			}
			to := bx.ToWhom()
			ito, vto := IdIndex(to), IdVersion(to)
			if ito == iex {
				if vto != vex {
					verbose.Printf("%d @ %v "+
						"FIXME prompt exchange update",
						ifrom, afrom)
					bx.Return()
				} else {
					err = bx.OpenWith(cfrom)
					if err != nil {
						verbose.Printf("%d @ %v "+
							"open %v",
							ifrom, afrom, err)
						bx.Return()
					} else {
						ex.rx(cctx, &wg, bx)
					}
				}
			} else if vto != ex.ver[ito] {
				verbose.Printf("%d @ %v "+
					"FIXME prompt host %d update",
					ifrom, afrom, ito)
				bx.Return()
			} else if ato, ok := ex.service[ito]; !ok {
				verbose.Printf("%d @ %v no service to %d",
					ifrom, afrom, ito)
				bx.Return()
			} else if cto, ok := ex.gcm[ito]; !ok {
				verbose.Printf("%d @ %v not peered with %d",
					ifrom, afrom, ito)
				bx.Return()
			} else {
				verbose.Printf("%d @ %v reseal and send to "+
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
				errata.Println(err)
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
	pi, err := netpdu.TunPI(bx.Contents).Header()
	if err != nil {
		verbose.Println("rx %d @ %v pi %v", ifrom, afrom, err)
		return
	}
	d := netpdu.TunPI(bx.Contents).Data()
	switch pi.Proto {
	case VPN_P_HELLO:
		verbose.Printf("rx %d @ %v hello %v",
			ifrom, afrom, VpnHelloPDU(d))
		ex.ack(ex.pktTxCh, from, afrom)
	case VPN_P_WHOIS_ADDRESSED:
		if addr := VpnWhoisAddress(d); !addr.IsValid() {
			verbose.Printf("rx %d @ %v whois %v",
				ifrom, afrom, "underrun")
		} else if blk := ex.pem.addressed[addr]; blk == nil {
			verbose.Printf("rc %d @ %v whois %v",
				ifrom, afrom, addr)
			wg.Add(1)
			go ex.whoisAddressedRoutine(ctx, wg, addr)
		} else {
			ex.txpubkey(from, blk)
		}
	case VPN_P_WHOIS_IDENTIFIED:
		if id, err := VpnWhoisId(d); err != nil {
			verbose.Printf("rx %d @ %v whois %v",
				ifrom, afrom, err)
		} else if blk := ex.pem.identified[IdIndex(id)]; blk == nil {
			verbose.Printf("rx %d @ %v whois %d",
				ifrom, afrom, id)
			wg.Add(1)
			go ex.whoisIdRoutine(ctx, wg, id)
		} else {
			ex.txpubkey(from, blk)
		}
	case VPN_P_WHOIS_SERVICE:
		verbose.Printf("rx %d @ %v whois-service %v",
			ifrom, afrom, VpnWhoisService(d))
	case netph.ETH_P_IP:
		ex.rxIP(from, netpdu.IP(d))
	case netph.ETH_P_IPV6:
		ex.rxIP6(from, netpdu.IP6(d))
	default:
		verbose.Printf("rx %d @ %v unknown proto %#x",
			ifrom, afrom, pi.Proto)
	}
}

func (ex *exchange) ack(ch chan<- *box.Box, to box.Id, ap netip.AddrPort) {
	var err error
	bx := box.New()
	defer bx.Return()
	i := IdIndex(to)
	c := ex.gcm[i]
	bx.Contents, err = xnet.Add(bx.Contents, netph.TunPI{
		Proto: VPN_P_HELLO,
	})
	if err != nil {
		errata.Print(err)
		return
	}
	bx.Contents, err = xnet.Add(bx.Contents, time.Now().UnixMicro())
	if err != nil {
		errata.Print(err)
		return
	}
	bx.AddrPort = ap
	bx.From(ex.id)
	bx.To(to)
	bx.CloseWith(c)
	bx.SealWith(c)
	bx.NonBlockingPut(ch)
	verbose.Printf("ack %d @ %v", i, ap)
}

func (ex *exchange) txpubkey(to box.Id, blk *pem.Block) {
	var err error
	i := IdIndex(to)
	c := ex.gcm[i]
	bx := box.New()
	defer bx.Return()
	bx.Contents, err = xnet.Add(bx.Contents, netph.TunPI{
		Proto: VPN_P_PUBLIC_KEY,
	})
	if err != nil {
		errata.Print(err)
		return
	}
	pem.Encode(bx, blk)
	bx.From(ex.id)
	bx.To(to)
	bx.CloseWith(c)
	bx.SealWith(c)
	bx.NonBlockingPut(ex.pktTxCh)
}

func (ex *exchange) rxIP(from box.Id, pdu netpdu.IP) {
	_, err := pdu.Header()
	if err != nil {
		verbose.Println("ip:", err)
		return
	}
	verbose.Println("FIXME mcast", pdu)
}

func (ex *exchange) rxIP6(from box.Id, pdu netpdu.IP6) {
	h, err := pdu.Header()
	if err != nil {
		verbose.Println("ip6:", err)
		return
	}
	d := pdu.Data()
	da := netip.AddrFrom16(h.DA)
	if h.NextHeader == netph.IPPROTO_ICMPV6 {
		if da.IsMulticast() || da == ex.addr {
			ex.rxICMP6(from, h.SA[:], pdu, netpdu.ICMP6(d))
		} else {
			verbose.Println("FIXME route:", pdu)
		}
	} else if da.IsMulticast() {
		verbose.Println("FIXME mcast:", pdu)
	} else if da != ex.addr {
		verbose.Println("FIXME route:", pdu)
	} else if true {
		verbose.Println("dropped:", pdu)
	}
}

func (ex *exchange) rxICMP6(
	from box.Id,
	addr []byte,
	ip6 netpdu.IP6,
	icmp6 netpdu.ICMP6,
) {
	sum := ip6.Checksum(netph.IPPROTO_ICMPV6, uint(len(icmp6)), icmp6)
	if sum != 0 && ^sum != 0 {
		verbose.Printf("bad sum: %#04x", sum)
		return
	}
	h, err := icmp6.Header()
	if err != nil {
		verbose.Print(err)
		return
	}
	d := icmp6.Data()
	switch h.Type {
	case netph.ICMP6TypeEchoRequest:
		ex.txICMP6EchoReply(from, addr, netpdu.ICMP6EchoRequest(d))
	case netph.ICMP6TypeRouterSolicitation:
		verbose.Println("FIXME reply:", ip6)
	default:
		verbose.Println("dropped:", "type", h.Type, ip6)
	}
}

func (ex *exchange) txICMP6EchoReply(
	id box.Id,
	addr []byte,
	req netpdu.ICMP6EchoRequest,
) {
	verbose.Println("FIXME", ex.addr, "reply to", net.IP(addr))
}

func (ex *exchange) whoisAddressedRoutine(
	ctx context.Context, wg *sync.WaitGroup, addr netip.Addr,
) {
	defer wg.Done()
	verbose.Println("whois", addr)
	blk, err := ex.whoisAddressed(ctx, addr)
	if err != nil {
		verbose.Println(err)
	} else {
		ex.whoisResponseCh <- blk
	}
}

func (ex *exchange) whoisIdRoutine(
	ctx context.Context, wg *sync.WaitGroup, id box.Id,
) {
	defer wg.Done()
	verbose.Println("whois", IdIndex(id))
	blk, err := ex.whoisIdentified(ctx, id)
	if err != nil {
		verbose.Println(err)
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
	verbose.Printf("peer[%d] ok", idi)
	return nil
}
