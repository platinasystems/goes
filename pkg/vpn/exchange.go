// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
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
	if ex.udp, err = udpListen(vpnListen); err != nil {
		return err
	}
	wg.Go(func() {
		defer xlog.Info.Println("closed", ex.udp.LocalAddr())
		defer ex.udp.Close()
		<-ctx.Done()
	})
	sap, err := udpLocalAddrPort(ex.udp)
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

	m := mp.Get()
	defer mp.Put(m)
selection:
	for err == nil {
		var n int
		select {
		case <-ctx.Done():
			break selection
		case err = <-ex.fault:
		case t := <-tc:
			err = greetings(ctx, ex.udp, ex.start, t.UnixMicro(),
				ex.via.Id, ex.via.Service)
			if err != nil {
				return err
			}
		case gx, ok := <-ex.whoisRspCh:
			if !ok {
				break selection
			}
			iid := gx.Id.Index()
			ex.ver[iid] = gx.Id.Version()
			ex.verifiers[iid], err = gx.verifier()
		default:
		}
		m.Data = m.Data[:cap(m.Data)]
		n, m.AddrPort, err = ex.udp.ReadFromUDPAddrPort(m.Data)
		if err != nil {
			return xerrors.Suppress(err, net.ErrClosed)
		}
		m.Data = m.Data[:n]
		i := n - SizeofLabel
		if i < 0 {
			xlog.Errata.Print("underrun")
		} else if tid, fid := ScanLabel(m.Data); !ex.isOK(fid) {
			xlog.Trace.Println("whois from", fid)
			ex.whoisQueue(ctx, fid.Index())
		} else if tid == fid {
			if ex.verifyHello(m, fid) {
				ex.rap[fid.Index()] = m.AddrPort
				now := time.Now().UnixMicro()
				err = greetings(ctx, ex.udp, ex.start, now,
					fid, m.AddrPort)
				if err != nil {
					return err
				}
			}
		} else if !ex.isOK(tid) {
			xlog.Trace.Println("whois from", tid)
			ex.whoisQueue(ctx, tid.Index())
		} else if rap, ok := ex.rap[tid.Index()]; ok {
			xlog.Trace.Print("forward ", tid, "@", rap,
				"<-", fid, "@", m.AddrPort)
			_, err = ex.udp.WriteToUDPAddrPort(m.Data, rap)
			if err != nil {
				return err
			}
		} else {
			xlog.Trace.Print("dropped ", tid, "<-", fid, "@",
				m.AddrPort)
		}
	}
	return nil
}

type exchange struct {
	subscriber
	rap map[int]netip.AddrPort
	ver map[int]uint8
	udp *net.UDPConn
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
