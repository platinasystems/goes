// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xcontext"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

// common to guest and exchange
type subscriber struct {
	rest
	via *Registration
	// Registry Unix Micro Start at checkin
	start      int64
	whoisReqCh chan any // name, [Id], or [netip.Addr]
	whoisRspCh chan *Registration
	verifiers  map[int]func([]byte) bool
}

func (sub *subscriber) config() error {
	depth := runtime.NumCPU()
	sub.whoisReqCh = make(chan any, depth)
	sub.whoisRspCh = make(chan *Registration, depth)
	sub.verifiers = make(map[int]func([]byte) bool)

	return sub.rest.config()
}

func (sub *subscriber) greetings(
	ctx context.Context, now time.Time, to Id, ap netip.AddrPort,
) error {
	m := mp.Get()
	defer mp.Put(m)
	m.Data = m.Data[:0]
	m.AddrPort = ap

	sign, err := FirstPrivSigFileSign()
	if err != nil {
		return err
	}
	m.Data, err = xnet.Attach(m.Data, sub.start)
	if err == nil {
		m.Data, err = xnet.Attach(m.Data, now.UnixMicro())
		if err == nil {
			m.Data = append(m.Data, sign(m.Data)...)
		}
	}
	m.Data = append(m.Data, MyLabel...)
	xlog.Trace.Print("tx hello ", to, ", ", now)
	return udpSend(ctx, m)
}

func (sub *subscriber) validateCheckinResponse(rsp *http.Response) error {
	var err error

	if rsp == nil {
		return xerrors.Invalid("checkin response")
	}

	vcsRev := xprogram.VcsRevision.String()
	regVcsRev := rsp.Header.Get(RestVcsRevision)
	if vcsRev != regVcsRev {
		return fmt.Errorf("upgrade to %s", regVcsRev)
	}

	start := rsp.Header.Get(RestUnixMicroStart)
	if len(start) == 0 {
		return xerrors.Unavailable("registry start time")
	}
	sub.start, err = strconv.ParseInt(start, 10, 64)
	return xerrors.Label(err, "registry start")
}

func (sub *subscriber) resolve(ctx context.Context, hn string) (
	netip.Addr, error,
) {
	var z netip.Addr
	if addr, err := netip.ParseAddr(hn); err == nil {
		return addr, err
	}
	ips, err := WaitForResolution(ctx, "ip", hn, time.Minute)
	if err != nil {
		return z, err
	}
	if len(ips) == 0 {
		return z, xerrors.Invalid(hn)
	}
	addr, ok := netip.AddrFromSlice(ips[0])
	if !ok {
		return z, xerrors.Invalid(hn)
	}
	return addr, nil
}

func (sub *subscriber) verifyHello(m *xnet.Msg, from Id) bool {
	if f, ok := sub.verifiers[from.Index()]; !ok {
		xlog.Trace.Println("can't verify", from)
	} else if data := TruncLabel(m.Data); !f(data) {
		xlog.Trace.Println("imposter", from)
	} else if len(data) < 2*8 {
		xlog.Trace.Println("incomplete hello from", from)
	} else if start := xnet.Int64(m.Data); start < sub.start {
		xlog.Errata.Println("FIXME out-of-data peer")
	} else if start > sub.start {
		xlog.Errata.Println("FIXME newer peer")
	} else {
		t := time.UnixMicro(xnet.Int64(m.Data[8:]))
		xlog.Trace.Printf("rx hello %v, %v", from, t)
		return true
	}
	return false
}

func (sub *subscriber) waitForExchange(
	ctx context.Context, namePort string,
) error {
	begin := time.Now()
	for {
		if xcontext.IsDone(ctx) {
			return ctx.Err()
		}
		err := sub.viaWhom(ctx, namePort)
		if err == nil {
			break
		} else if time.Now().Sub(begin) > time.Minute {
			return err
		} else if !IsNotFound(err) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			xlog.Info.Println("retry", namePort, "...")
		}
	}
	xlog.Info.Println("after", time.Now().Sub(begin), "found",
		sub.via.Name, "with", sub.via.Id, "at", sub.via.Service)
	return nil
}

func (sub *subscriber) viaWhom(ctx context.Context, namePort string) error {
	var addr netip.Addr

	name := namePort
	port := uint16(defaultServicePort)
	colon := strings.LastIndex(namePort, ":")
	if colon > 0 && colon < len(namePort)-1 {
		name = namePort[:colon]
		if _, err := fmt.Sscan(namePort[colon+1:], &port); err != nil {
			return err
		}
	}
	via, err := sub.whois(ctx, name)
	if err != nil {
		return err
	}
	sub.via = via
	sub.verifiers[sub.via.Id.Index()], err = sub.via.verifier()
	if err != nil {
		return err
	}
	switch {
	case sub.via.Service.IsValid():
	case len(sub.via.IPAddresses) > 0:
		addr, err = netip.ParseAddr(sub.via.IPAddresses[0])
	case len(sub.via.URIs) > 0:
		uri, err := url.Parse(sub.via.URIs[0])
		if err != nil {
			return err
		}
		if s := uri.Port(); len(s) > 0 {
			if _, err = fmt.Sscan(s, &port); err != nil {
				return err
			}
		}
		addr, err = sub.resolve(ctx, uri.Hostname())
	case len(sub.via.DNSNames) > 0:
		addr, err = sub.resolve(ctx, sub.via.DNSNames[0])
	default:
		addr, err = sub.resolve(ctx, name)
	}
	if err == nil {
		sub.via.Service = netip.AddrPortFrom(addr, port)
	}
	return err
}

func (sub *subscriber) whoisQueue(ctx context.Context, q any) bool {
	return xcontext.Queue(ctx, sub.whoisReqCh, q)
}

func (sub *subscriber) whoisService(ctx context.Context) {
	cn := sub.rest.crt.Subject.CommonName

	xlog.Info.Println("start", cn, "whois request service")
	defer xlog.Info.Println("stopped", cn, "whois request service")
	defer close(sub.whoisRspCh)

	for {
		select {
		case <-ctx.Done():
			return
		case q, ok := <-sub.whoisReqCh:
			if !ok {
				return
			}
			xlog.Trace.Println("whois", q)
			rsp, err := sub.whois(ctx, q)
			if err != nil {
				xlog.Errata.Print(err)
			} else {
				xcontext.Queue(ctx, sub.whoisRspCh, rsp)
			}
		}
	}
}
