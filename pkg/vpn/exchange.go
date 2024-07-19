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

	"github.com/platinasystems/goes/v2/pkg/box"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

func exchangeDaemon(ctx context.Context, args []string) error {
	var wg sync.WaitGroup
	var ex exchange

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] `+RegistryURL+`
Exchange ciphered packets between guests.

{{flags .}}`)

	opts.svc = DefaultService()

	err := parseOpts(ctx, args)
	if err != nil {
		return err
	} else if args = flag.Args(); len(args) == 0 {
		return xerrors.Incomplete("registry")
	}
	reg := args[0]

	ex.pem.addressed = make(map[netip.Addr]*pem.Block)
	ex.pem.identified = make(map[int]*pem.Block)

	pktRxCh := make(chan *box.Box, 4)
	pktTxCh := make(chan *box.Box, 4)

	ex.whoisResponseCh = make(chan *pem.Block)

	if err = ex.register(ctx, reg, opts.svc); err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(ctx)

	iex, vex := IdIndex(ex.id), IdVersion(ex.id)

	id := fmt.Sprintf("(%d, %v)", iex, opts.svc)
	verbose.Println("start", id)
	defer verbose.Println("stopped", id, err)
	defer wg.Wait()
	defer cancel()
	defer verbose.Println("stopping", id, "...")

	udp, err := xerrors.MarkResult(net.ListenUDP("udp", &net.UDPAddr{
		IP:   opts.svc.Addr().AsSlice(),
		Port: int(opts.svc.Port()),
		// Zone: FIXME,
	}))
	if err != nil {
		return err
	}

	defer udp.Close()

	wg.Add(1)
	go pktRxRoutine(cctx, &wg, udp, pktRxCh)
	wg.Add(1)
	go pktTxRoutine(cctx, &wg, udp, pktTxCh)
	defer close(pktTxCh)

pktRxLoop:
	for {
		select {
		case <-cctx.Done():
			verbose.Println("done", cctx.Err())
			break pktRxLoop
		case bx, ok := <-pktRxCh:
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
				verbose.Printf("(%d, %v) who?", ifrom, afrom)
				wg.Add(1)
				go ex.whoisIdRoutine(cctx, &wg, from)
				bx.Return()
				continue pktRxLoop
			}
			if err := bx.UnsealWith(cfrom); err != nil {
				verbose.Printf("(%d, %v) unseal %v",
					ifrom, afrom, err)
				bx.Return()
				continue pktRxLoop
			}
			to := bx.ToWhom()
			ito, vto := IdIndex(to), IdVersion(to)
			if ito == iex {
				if vto != vex {
					verbose.Printf("(%d, %v) "+
						"FIXME prompt exchange update",
						ifrom, afrom)
					bx.Return()
				} else {
					err = bx.OpenWith(cfrom)
					if err != nil {
						verbose.Printf("(%d, %v) "+
							"open %v",
							ifrom, afrom, err)
						bx.Return()
					} else {
						ex.rx(cctx, &wg, bx, pktTxCh)
					}
				}
			} else if vto != ex.ver[ito] {
				verbose.Printf("(%d, %v) "+
					"FIXME prompt host %d update",
					ifrom, afrom, ito)
				bx.Return()
			} else if ato, ok := ex.service[ito]; !ok {
				verbose.Printf("(%d, %v) no service to %d",
					ifrom, afrom, ito)
				bx.Return()
			} else if cto, ok := ex.gcm[ito]; !ok {
				verbose.Printf("(%d, %v) not peered with %d",
					ifrom, afrom, ito)
				bx.Return()
			} else {
				verbose.Printf("(%d, %v) reseal and send to "+
					"(%d, %v)", ifrom, afrom, ito, ato)
				bx.AddrPort = ato
				bx.SealWith(cto)
				bx.NonBlockingPut(pktTxCh)
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
	pem struct {
		addressed  map[netip.Addr]*pem.Block
		identified map[int]*pem.Block
	}
	whoisResponseCh chan *pem.Block
}

func (ex *exchange) rx(
	ctx context.Context,
	wg *sync.WaitGroup,
	bx *box.Box,
	pktTxCh chan<- *box.Box,
) {
	var pi netph.TunPI
	afrom := bx.AddrPort
	from := bx.FromWhom()
	ifrom := IdIndex(from)
	cfrom := ex.gcm[ifrom]
	_, err := pi.ReadFrom(bx)
	if err != nil {
		verbose.Println("rx (%d, %v) pi %v", ifrom, afrom, err)
		bx.Return()
		return
	}
	switch pi.Proto {
	case VPN_P_HELLO:
		verbose.Printf("rx (%d, %v) hello %v",
			ifrom, afrom, VpnHelloTimeSpan(bx))
		bx.Return()
	case VPN_P_WHOIS_ADDRESSED:
		if addr := VpnWhoisAddress(bx); !addr.IsValid() {
			verbose.Printf("rx (%d, %v) whois %v",
				ifrom, afrom, "underrun")
			bx.Return()
		} else if blk := ex.pem.addressed[addr]; blk == nil {
			verbose.Printf("rc (%d, %v) whois %v",
				ifrom, afrom, addr)
			bx.Return()
			wg.Add(1)
			go ex.whoisAddressedRoutine(ctx, wg, addr)
		} else {
			netph.TunPI{
				Proto: VPN_P_PUBLIC_KEY,
			}.WriteTo(bx)
			pem.Encode(bx, blk)
			bx.From(ex.id)
			bx.To(from)
			bx.CloseWith(cfrom)
			bx.SealWith(cfrom)
			bx.NonBlockingPut(pktTxCh)
		}
	case VPN_P_WHOIS_IDENTIFIED:
		if id, ok := VpnWhoisId(bx); !ok {
			verbose.Printf("rx (%d, %v) whois %v",
				ifrom, afrom, "underrun")
			bx.Return()
		} else if blk := ex.pem.identified[IdIndex(id)]; blk == nil {
			verbose.Printf("rx (%d, %v) whois %d",
				ifrom, afrom, id)
			bx.Return()
			wg.Add(1)
			go ex.whoisIdRoutine(ctx, wg, id)
		} else {
			netph.TunPI{
				Proto: VPN_P_PUBLIC_KEY,
			}.WriteTo(bx)
			pem.Encode(bx, blk)
			bx.From(ex.id)
			bx.To(from)
			bx.CloseWith(cfrom)
			bx.SealWith(cfrom)
			bx.NonBlockingPut(pktTxCh)
		}
	case VPN_P_WHOIS_SERVICE:
		verbose.Printf("rx (%d, %v) whois-service %v",
			ifrom, afrom, VpnWhoisService(bx))
		bx.Return()
	default:
		verbose.Printf("rx (%d, %v) unknown proto %#x",
			ifrom, afrom, pi.Proto)
		bx.Return()
	}
}

func (ex *exchange) whoisAddressedRoutine(
	ctx context.Context, wg *sync.WaitGroup, addr netip.Addr,
) {
	defer wg.Done()
	verbose.Println("whois", addr)
	blk, err := httpWhoIsAddressed(ctx, ex.registry, addr)
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
	blk, err := httpWhoIsIdentified(ctx, ex.registry, id)
	if err != nil {
		verbose.Println(err)
	} else {
		ex.whoisResponseCh <- blk
	}
}

func (ex *exchange) whoisResponse(blk *pem.Block) error {
	addr, err := xerrors.MarkResult(addressHeader(blk))
	if err != nil {
		return err
	}
	id, err := xerrors.MarkResult(idHeader(blk))
	if err != nil {
		return err
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
