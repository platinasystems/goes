// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"encoding/pem"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/box/label"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/goes"
)

const (
	exHello          = "hello"
	exWhoisAddressed = "whois addressed"
	exWhoisLabelled  = "whois labelled"
)

type exchange struct {
	client
	pem struct {
		addressed map[netip.Addr]*pem.Block
		labelled  map[label.Index]*pem.Block
	}
	ch struct {
		pkt struct {
			rx chan *crate
			tx chan *crate
		}
		whois struct {
			response chan *pem.Block
		}
	}
	wg sync.WaitGroup
}

func (ex *exchange) daemon(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} [<options>] ` + vpnRegistrySyntax + `
Start VPN packet exchange.
{{flags .}}`

	if goes.ContextComplete(ctx) {
		return nil
	}
	svc, err := vpnDaemonFlags(ctx, usage, args)
	if err != nil {
		return err
	}

	args = goes.ContextFlags(ctx).Args()
	if len(args) < 1 {
		return ErrIncomplete
	}

	ex.pem.addressed = make(map[netip.Addr]*pem.Block)
	ex.pem.labelled = make(map[label.Index]*pem.Block)

	ex.ch.pkt.rx = make(chan *crate, 4)
	ex.ch.pkt.tx = make(chan *crate, 4)
	ex.ch.whois.response = make(chan *pem.Block)

	defer ex.wg.Wait()

	if err = ex.register(ctx, args[0], svc); err != nil {
		return err
	}

	udp, err := egress.MarkResult(net.ListenUDP("udp", &net.UDPAddr{
		IP:   svc.Addr().AsSlice(),
		Port: int(svc.Port()),
		// Zone: FIXME,
	}))
	if err != nil {
		return err
	}

	defer udp.Close()

	verbose.Printf("start (%d, %v)", ex.label, svc)
	defer verbose.Printf("stopped (%d, %v)", ex.label, svc)

	ex.wg.Add(1)
	go pktRxRoutine(ctx, &ex.wg, udp, ex.ch.pkt.rx)
	ex.wg.Add(1)
	go pktTxRoutine(ctx, &ex.wg, udp, ex.ch.pkt.tx)

pktRxLoop:
	for {
		select {
		case <-ctx.Done():
			return nil
		case c := <-ex.ch.pkt.rx:
			from := c.box.FromWhom()
			ifrom := from.Index()
			if from.Version() != ex.ver[ifrom] {
				ex.wg.Add(1)
				go ex.whoisLabelledRoutine(ctx, from)
				inventory.Put(c)
				continue pktRxLoop
			}
			ex.service[ifrom] = c.ap
			if gcm, ok := ex.gcm[ifrom]; !ok {
				ex.wg.Add(1)
				go ex.whoisLabelledRoutine(ctx, from)
				inventory.Put(c)
				continue pktRxLoop
			} else if err := c.box.UnsealWith(gcm); err != nil {
				verbose.Println(ifrom, c.ap, "unseal", err)
				inventory.Put(c)
				continue pktRxLoop
			} else if to := c.box.ToWhom(); to == ex.label {
				c.box, err = c.box.OpenWith(gcm)
				if err != nil {
					verbose.Println(ifrom, c.ap, "open", err)
					inventory.Put(c)
				} else {
					ex.rx(ctx, c)
				}
			} else if ito := to.Index(); ito == ex.label.Index() {
				verbose.Print("FIXME exchange version")
				inventory.Put(c)
			} else if vto, ok := ex.ver[ito]; !ok {
				verbose.Print("FIXME whois", ito)
				inventory.Put(c)
			} else if vto != to.Version() {
				verbose.Print("FIXME to version")
				inventory.Put(c)
			} else if c.ap, ok = ex.service[ito]; !ok {
				verbose.Println(ifrom, c.ap, "no service")
				inventory.Put(c)
			} else if gcm, ok = ex.gcm[ito]; !ok {
				verbose.Println(ifrom, c.ap, "unknown", ito)
				inventory.Put(c)
			} else {
				verbose.Println("reseal/send", to, c.ap)
				c.box.SealWith(gcm)
				c.Put(ex.ch.pkt.tx)
			}
		case blk := <-ex.ch.whois.response:
			if err = ex.whoisResponse(blk); err != nil {
				fmt.Fprintln(errata, err)
			}
		}
	}
}

func (ex *exchange) rx(ctx context.Context, c *crate) {
	var blk *pem.Block
	from := c.box.FromWhom()
	ifrom := from.Index()
	gcm := ex.gcm[ifrom]
	s := string(c.box.Contents())
	switch {
	case strings.HasPrefix(s, exHello):
		verbose.Println(ifrom, c.ap, strings.TrimSpace(s))
		inventory.Put(c)
		return
	case strings.HasPrefix(s, exWhoisAddressed):
		s = strings.TrimPrefix(s, exWhoisAddressed)
		s = strings.TrimSpace(s)
		addr, err := egress.MarkResult(netip.ParseAddr(s))
		if err != nil {
			fmt.Fprintln(verbose, err)
			inventory.Put(c)
			return
		} else if blk = ex.pem.addressed[addr]; blk == nil {
			ex.wg.Add(1)
			go ex.whoisAddressedRoutine(ctx, addr)
			inventory.Put(c)
			return
		}
	case strings.HasPrefix(s, exWhoisLabelled):
		var lbl label.Label
		s = strings.TrimPrefix(s, exWhoisLabelled)
		s = strings.TrimSpace(s)
		_, err := egress.MarkResult(fmt.Sscan(s, &lbl))
		if err != nil {
			fmt.Fprintln(verbose, err)
			inventory.Put(c)
			return
		} else if blk = ex.pem.labelled[lbl.Index()]; blk == nil {
			ex.wg.Add(1)
			go ex.whoisLabelledRoutine(ctx, lbl)
			inventory.Put(c)
			return
		}
	default:
		verbose.Println(ifrom, c.ap, "unknown", s)
		inventory.Put(c)
		return
	}
	c.box = c.box.From(ex.label)
	c.box = c.box.To(from)
	c.box = c.box.Empty()
	if err := egress.Mark(pem.Encode(c, blk)); err != nil {
		fmt.Fprint(errata, err)
		inventory.Put(c)
	} else {
		c.box = c.box.CloseWith(gcm)
		c.box.SealWith(gcm)
		c.Put(ex.ch.pkt.tx)
	}
}

func (ex *exchange) whoisAddressedRoutine(
	ctx context.Context, addr netip.Addr,
) {
	defer ex.wg.Done()
	verbose.Println("whois addressed", addr)
	blk, err := httpWhoIsAddressed(ctx, ex.registry, addr)
	if err != nil {
		verbose.Println(err)
	} else {
		ex.ch.whois.response <- blk
	}
}

func (ex *exchange) whoisLabelledRoutine(
	ctx context.Context, lbl label.Label,
) {
	defer ex.wg.Done()
	verbose.Println("whois labelled", lbl.Index())
	blk, err := httpWhoIsLabelled(ctx, ex.registry, lbl)
	if err != nil {
		verbose.Println(err)
	} else {
		ex.ch.whois.response <- blk
	}
}

func (ex *exchange) whoisResponse(blk *pem.Block) error {
	addr, err := egress.MarkResult(addressHeader(blk))
	if err != nil {
		return err
	}
	lbl, err := egress.MarkResult(labelHeader(blk))
	if err != nil {
		return err
	}
	ilbl := lbl.Index()
	if err = ex.peer(blk); err != nil {
		return err
	}
	ex.pem.addressed[addr] = blk
	ex.pem.labelled[ilbl] = blk
	verbose.Printf("peer[%d] ok", ilbl)
	return nil
}
