// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/netip"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet"
)

// Exchange is a UDP server that forwards ciphered packets between guest's.
func Exchange(ctx context.Context, args []string) error {
	var svc, via string
	defer xlog.Info.Println("stopped", svc)

	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [via[:port]]
Exchange ciphered packets between guests.

{{flags .}}`)

	defineListen(defaultServicePort)
	definePublic()
	defineRestFlags()
	enableTrace()
	enableQuiet()
	enableVerbose()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if flag.CommandLine.NArg() > 0 {
		via = flag.CommandLine.Arg(0)
	}
	ex := exchange{
		rap: make(map[int]netip.AddrPort),
		ver: make(map[int]uint8),
	}
	if err = ex.subscriber.config(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer wg.Wait()
	defer cancel()

	wg.Go(func() { xlog.AlarmHandler(ctx) })

	if err = ex.checkin(ctx, via); err != nil {
		return err
	}
	if err = udpListen(); err != nil {
		return err
	}
	wg.Go(func() {
		defer xlog.Info.Println("closed", udp.LocalAddr())
		defer udp.Close()
		<-ctx.Done()
	})
	wg.Go(func() {
		if t := udpStream(ctx); err == nil {
			err = t
		}
	})
	sap, err := udpLocalAddrPort(udp)
	if err != nil {
		return err
	} else if !vpnPublic.Addr().IsUnspecified() {
		sap = vpnPublic
	}
	if sap.Addr().IsLoopback() {
		return xerrors.Invalid("address", sap)
	}

	svc = fmt.Sprintf("%v @ %v", MyId, sap)
	xlog.Info.Println("start", svc)
	defer xlog.Info.Println("stopping", svc, "...")

	wg.Go(func() { ex.whoisService(ctx) })
	defer close(ex.whoisReqCh)

	var tc <-chan time.Time
	if ex.via != nil {
		tkr := time.NewTicker(10 * time.Second)
		defer tkr.Stop()
	} else {
		tc = make(chan time.Time)
	}

selection:
	for err == nil {
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case err = <-ex.fault:
		case t := <-tc:
			err = ex.greetings(ctx, t, ex.via.Id, ex.via.Service)
		case gx, ok := <-ex.whoisRspCh:
			if !ok {
				break selection
			}
			iid := gx.Id.Index()
			ex.ver[iid] = gx.Id.Version()
			ex.verifiers[iid], err = gx.verifier()
		case m, ok := <-rmc:
			if !ok {
				break selection
			}
			err = ex.rx(ctx, m)
			mp.Put(m)
		}
	}
	return xerrors.Suppress(err, context.Canceled, errEOC)
}

type exchange struct {
	subscriber
	rap map[int]netip.AddrPort
	ver map[int]uint8
}

func (ex *exchange) checkin(ctx context.Context, via string) error {
	var id uint

	buf := ex.subscriber.rest.alloc()
	defer ex.subscriber.rest.free(buf)

	rsp, err := ex.subscriber.rest.request(ctx, buf, http.MethodPut,
		"", nil,
		RestOp, RestOpCheckin,
		RestOpCheckin, RestOpCheckinExchange,
		RestOpCheckinExchange, RestOpCheckinExchangeVia,
		RestOpCheckinExchangeVia, via)
	if err != nil {
		return err
	}
	if err = ex.validateCheckinResponse(rsp); err != nil {
		return err
	}
	if _, err = fmt.Fscan(buf, &id); err != nil {
		return err
	}
	MyId = Id(id)
	MyLabel = MakeLabel(MyId, MyId)
	xlog.Info.Println("registration:", MyId)
	if len(via) > 0 {
		err = ex.waitForExchange(ctx, via)
	}
	return err
}

func (ex *exchange) isOK(id Id) bool {
	v, ok := ex.ver[id.Index()]
	return ok && v == id.Version()
}

func (ex *exchange) rx(ctx context.Context, m *xnet.Msg) error {
	if to, from := ScanLabel(m.Data); !ex.isOK(from) {
		xlog.Trace.Println("whois from", from)
		ex.whoisQueue(ctx, from.Index())
	} else if to == from {
		if ex.verifyHello(m, from) {
			ex.rap[from.Index()] = m.AddrPort
			return ex.greetings(ctx, time.Now(), from, m.AddrPort)
		}
	} else if !ex.isOK(to) {
		xlog.Trace.Println("whois from", to)
		ex.whoisQueue(ctx, to.Index())
	} else if rap, ok := ex.rap[to.Index()]; ok {
		xlog.Trace.Print("forward ", to, "@", rap,
			"<-", from, "@", m.AddrPort)
		m.AddrPort = rap
		return udpSend(ctx, m)
	} else if ex.via.Service.IsValid() {
		xlog.Trace.Print("forward ", to, "<-", from, "@", m.AddrPort,
			" via ", ex.via.Id, "@", ex.via.Service)
		m.AddrPort = ex.via.Service
		return udpSend(ctx, m)
	} else {
		// FIXME forward to closest via
		xlog.Trace.Printf("dropped %v -> %v", from, to)
	}
	return nil
}
