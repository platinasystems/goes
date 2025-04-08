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

type exchange struct {
	client
	lladdr netip.Addr
	pktRxCh,
	pktTxCh chan *box.Box
	pem struct {
		addressed  map[netip.Addr]*pem.Block
		identified map[box.Id]*pem.Block
	}
	whoisResponseCh chan exchangeWhoisResponse
}

type exchangeWhoisResponse struct {
	blk *pem.Block
	err error
}

var exAll6nodes, exAll6routers netip.Addr

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

	exAll6nodes = netip.IPv6LinkLocalAllNodes()
	exAll6routers = netip.IPv6LinkLocalAllRouters()

	ex.pem.addressed = make(map[netip.Addr]*pem.Block)
	ex.pem.identified = make(map[box.Id]*pem.Block)

	ex.pktRxCh = make(chan *box.Box, 4)
	ex.pktTxCh = make(chan *box.Box, 4)

	ex.whoisResponseCh = make(chan exchangeWhoisResponse)

	if ex.lladdr, err = randLinkLocalAddr(); err != nil {
		return err
	} else {
		xlog.Info.Println("link-local address:", ex.lladdr)
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

	iex, vex := ex.id.Index(), ex.id.Version()

	svc := fmt.Sprintf("%v@%v", ex.id, sap)
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
			ap := bx.AddrPort
			from := bx.FromWhom()
			ifrom, vfrom := from.Index(), from.Version()
			via := bx.ViaWhom()
			ivia, vvia := via.Index(), via.Version()
			if ivia != iex {
				xlog.Errata.Printf("via %d != %d", ivia, iex)
				bx.Return()
				continue pktRxLoop
			}
			if vvia != vex {
				xlog.Errata.Printf("FIXME version via "+
					"%d != exchange %d", vvia, vex)
				bx.Return()
				continue pktRxLoop
			}
			if vfrom != ex.ver[ifrom] {
				xlog.Info.Println("update", ifrom)
				wg.Add(1)
				go ex.whoisRoutine(cctx, &wg, RestKeyId, from)
				bx.Return()
				continue pktRxLoop
			}
			ex.service[ifrom] = ap
			cfrom, ok := ex.gcm[ifrom]
			if !ok {
				xlog.Info.Println("whois", ifrom)
				wg.Add(1)
				go ex.whoisRoutine(cctx, &wg, RestKeyId, from)
				bx.Return()
				continue pktRxLoop
			}
			if err := bx.UnsealWith(cfrom); err != nil {
				xlog.Errata.Println("unseal", bx, err)
				bx.Return()
				continue pktRxLoop
			}
			to := bx.ToWhom()
			ito, vto := to.Index(), to.Version()
			if ito == iex {
				if vto != vex {
					// FIXME prompt exchange update
					xlog.Errata.Print("mismatch", bx)
				} else {
					err = bx.OpenWith(cfrom)
					if err != nil {
						xlog.Errata.Print("open",
							bx, err)
					} else {
						ex.rx(cctx, &wg, Box{bx})
					}
				}
				bx.Return()
			} else if vto != ex.ver[ito] {
				// FIXME prompt host %d update
				xlog.Errata.Print("mismatch", bx)
				bx.Return()
			} else if ato, ok := ex.service[ito]; !ok {
				xlog.Errata.Print("no service", bx)
				bx.Return()
			} else if cto, ok := ex.gcm[ito]; !ok {
				xlog.Errata.Print("not peered", bx)
				bx.Return()
			} else {
				bx.AddrPort = ato
				xlog.Info.Println("fwd", bx)
				bx.SealWith(cto)
				bx.NonBlockingPut(ex.pktTxCh)
			}
		case rsp, ok := <-ex.whoisResponseCh:
			if !ok {
				xlog.Errata.Println("closed whois response ch")
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

func (ex *exchange) rx(
	ctx context.Context,
	wg *sync.WaitGroup,
	bx Box,
) {
	xlog.Info.Println("rx", bx)
	from := bx.FromWhom()
	h, d, err := bx.PDU().Parse()
	if err != nil {
		return
	}
	switch h.Proto {
	case VPN_P_HELLO:
		ex.txHelloAck(from)
		ex.replicate("relay", bx.Box, from)
	case VPN_P_WHOIS_ADDRESSED:
		if addr := WhoisAddress(d); !addr.IsValid() {
			xlog.Errata.Println("rx whois underrun")
		} else if blk := ex.pem.addressed[addr]; blk == nil {
			xlog.Info.Println("rx whois:", addr)
			wg.Add(1)
			go ex.whoisRoutine(ctx, wg, RestKeyAddress, addr)
		} else {
			ex.txPubKey("whois addressed rsponse", from, blk)
		}
	case VPN_P_WHOIS_IDENTIFIED:
		if id, err := WhoisId(d); err != nil {
			xlog.Errata.Print("rx whois:", err)
		} else if blk := ex.pem.identified[id]; blk == nil {
			xlog.Info.Println("rx whois:", id)
			wg.Add(1)
			go ex.whoisRoutine(ctx, wg, RestKeyId, id)
		} else {
			ex.txPubKey("whois identified response", from, blk)
		}
	case VPN_P_WHOIS_SERVICE:
		xlog.Info.Println("FIXME rx whois:", WhoisService(d))
	case VPN_P_IP:
		xlog.Info.Println("rx dropped", bx)
	case VPN_P_IP6:
		ex.rxIP6(bx, from, netpdu.IP6(d))
	default:
		xlog.Info.Println("rx unknown proto:", h.Proto)
	}
}

func (ex *exchange) rxIP6(bx Box, from box.Id, ip6pdu netpdu.IP6) {
	ip6h, ip6d, err := ip6pdu.Parse()
	if err != nil {
		xlog.Info.Print(err)
		return
	}
	sum := ip6pdu.Checksum(ip6h.NextHeader, ip6d)
	if sum != 0 {
		xlog.Info.Print("bad sum")
		return
	}
	da := netip.AddrFrom16(ip6h.DA)
	if da.IsMulticast() {
		ex.replicate("multicast", bx.Box, from)
		return
	}
	if !ex.tome(da) {
		xlog.Info.Println("!me", bx)
		return
	}
	switch ip6h.NextHeader {
	case netph.IPPROTO_ICMPV6:
		icmp6pdu := netpdu.ICMP6(ip6d)
		if icmp6pdu.Type() == netph.ICMP6TypeEchoRequest {
			echo6req := netpdu.ICMP6EchoRequest(icmp6pdu)
			ex.txICMP6EchoReply(from, &ip6h, echo6req)
		} else {
			xlog.Info.Println("!echo", icmp6pdu)
		}
	case netph.IPPROTO_UDP:
		ex.rxUDP6(from, &ip6h, netpdu.UDP(ip6d))
	default:
		xlog.Info.Println("!(icmp|udp)", ip6pdu)
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
		xlog.Info.Println("FIXME reply", pdu)
	default:
		xlog.Info.Println("dropped", pdu)
	}
}

func (ex *exchange) tome(a netip.Addr) bool {
	return a.Compare(ex.lladdr) == 0 || a.Compare(ex.addr) == 0
}

func (ex *exchange) replicate(lbl string, bx *box.Box, from box.Id) {
	iex, ifrom := ex.id.Index(), from.Index()
	for id := range ex.pem.identified {
		if i := id.Index(); i != ifrom && i != iex {
			gcm := ex.gcm[i]
			clone := bx.Clone()
			clone.To(ex.id)
			clone.Via(ex.id)
			clone.From(from)
			clone.AddrPort = ex.service[i]
			xlog.Info.Println(lbl, Box{clone})
			clone.CloseWith(gcm)
			clone.SealWith(gcm)
			clone.NonBlockingPut(ex.pktTxCh)
		}
	}
}

func (ex *exchange) txHelloAck(to box.Id) {
	var err error
	ito := to.Index()
	c, ok := ex.gcm[ito]
	if !ok {
		xlog.Errata.Println("%d: no GCM", ito)
		return
	}
	bx := box.New()
	bx.AddrPort = ex.service[ito]
	bx.From(ex.id)
	bx.To(to)
	bx.Contents, err = xnet.Attach(bx.Contents, netph.TunPI{
		Proto: VPN_P_HELLO,
	})
	if err != nil {
		bx.Return()
		return
	}
	bx.Contents, err = xnet.Attach(bx.Contents, time.Now().UnixMicro())
	if err != nil {
		bx.Return()
		return
	}
	xlog.Info.Println("hello ack", Box{bx})
	bx.CloseWith(c)
	bx.SealWith(c)
	bx.NonBlockingPut(ex.pktTxCh)
}

func (ex *exchange) txICMP6EchoReply(
	to box.Id,
	req6 *netph.IP6,
	req netpdu.ICMP6EchoRequest,
) {
	ito := to.Index()
	c, ok := ex.gcm[ito]
	if !ok {
		xlog.Errata.Println("%d: no GCM", ito)
		return
	}
	reqh, reqd, err := req.Parse()
	if err != nil {
		return
	}
	bx := box.New()
	bx.AddrPort = ex.service[ito]
	bx.From(ex.id)
	bx.To(to)
	bx.Contents, err = xnet.Attach(bx.Contents, netph.TunPI{
		Proto: netph.ETH_P_IPV6,
	})
	if err != nil {
		bx.Return()
		return
	}
	ip6i := len(bx.Contents)
	bx.Contents, err = xnet.Attach(bx.Contents, netph.IP6{
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
	bx.Contents, err = xnet.Attach(bx.Contents, netph.ICMP6EchoReply{
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
	bx.Contents, err = xnet.Attach(bx.Contents, reqd)
	if err != nil {
		bx.Return()
		return
	}
	ip6 := netpdu.IP6(bx.Contents[ip6i:])
	ip6.SetLen()
	icmp6 := netpdu.ICMP6(bx.Contents[icmp6i:])
	sum := ip6.Checksum(netph.IPPROTO_ICMPV6, icmp6)
	icmp6.SetSum(sum)
	xlog.Info.Println("reply", Box{bx})
	bx.CloseWith(c)
	bx.SealWith(c)
	bx.NonBlockingPut(ex.pktTxCh)
}

func (ex *exchange) txPubKey(lbl string, to box.Id, blk *pem.Block) {
	var err error
	ito := to.Index()
	c, ok := ex.gcm[ito]
	if !ok {
		xlog.Errata.Println("%d: no GCM", ito)
		return
	}
	bx := box.New()
	bx.AddrPort = ex.service[ito]
	bx.From(ex.id)
	bx.To(to)
	bx.Contents, err = xnet.Attach(bx.Contents, netph.TunPI{
		Proto: VPN_P_PUBLIC_KEY,
	})
	if err != nil {
		bx.Return()
		return
	}
	pem.Encode(bx, blk)
	xlog.Info.Println(lbl, Box{bx})
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
	ito := to.Index()
	c, ok := ex.gcm[ito]
	if !ok {
		xlog.Errata.Println("%d: no GCM", ito)
		return
	}
	bx := box.New()
	bx.From(ex.id)
	bx.To(to)
	bx.AddrPort = ex.service[ito]
	bx.Contents, err = xnet.Attach(bx.Contents, netph.TunPI{
		Proto: netph.ETH_P_IPV6,
	})
	if err != nil {
		bx.Return()
		return
	}
	ip6i := len(bx.Contents)
	bx.Contents, err = xnet.Attach(bx.Contents, netph.IP6{
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
	bx.Contents, err = xnet.Attach(bx.Contents, netph.UDP{
		SP: requdp.DP,
		DP: requdp.SP,
	})
	if err != nil {
		bx.Return()
		return
	}
	bx.Contents, err = xnet.Attach(bx.Contents, data)
	if err != nil {
		bx.Return()
		return
	}
	ip6 := netpdu.IP6(bx.Contents[ip6i:])
	ip6.SetLen()
	udp := netpdu.UDP(bx.Contents[udpi:])
	udp.SetLen()
	udp.SetSum(ip6.Checksum(netph.IPPROTO_UDP, udp))
	xlog.Info.Println("tx", Box{bx})
	bx.CloseWith(c)
	bx.SealWith(c)
	bx.NonBlockingPut(ex.pktTxCh)
}

func (ex *exchange) whoisRoutine(
	ctx context.Context, wg *sync.WaitGroup, key string, value any,
) {
	defer wg.Done()
	blk, err := ex.whois(ctx, key, value)
	if err != nil {
		err = fmt.Errorf("%w (whois %s %v)", err, key, value)
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
	if err = ex.peer(blk); err != nil {
		return err
	}
	ex.pem.addressed[addr] = blk
	ex.pem.identified[id] = blk
	return nil
}

func randLinkLocalAddr() (lladdr netip.Addr, err error) {
	var a [netph.IPv6len]byte
	a[0] = 0xfe
	a[1] = 0x80
	n, err := rand.Read(a[8:])
	if err != nil {
	} else if n != len(a[8:]) {
		err = xerrors.Underrun("rand-link-local")
	} else {
		lladdr = netip.AddrFrom16(a)
	}
	return
}
