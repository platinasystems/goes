// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"

	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xsignal"
)

var exchange struct {
	sub map[int]*Subscriber

	fromVpnC <-chan *xnet.Msg
	toVpnC   chan<- *xnet.Msg
}

// Exchange is a UDP server that forwards ciphered packets between guest's.
func Exchange(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
Exchange ciphered packets between guests.

{{flags .}}`)

	defineRestFlags()
	enableTrace()
	enableQuiet()
	enableVerbose()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	alarm := make(chan os.Signal, 2)
	signal.Notify(alarm, xsignal.Alarm)
	defer signal.Stop(alarm)

	ctx, cancel := context.WithCancel(ctx)
	defer xlog.Trace.Println("stopped")
	defer wg.Wait()

	if err = restInit(); err != nil {
		return err
	}
	defer close(rest.whoisReqC)

	exchange.sub = make(map[int]*Subscriber)

	if err = restAssertVcsMatch(ctx); err != nil {
		return err
	}

	port, err := restExchangeCheckin(ctx)
	if err != nil {
		return err
	} else if port == 0 {
		return errors.New("unassigned port")
	}

	exchange.fromVpnC, exchange.toVpnC, err = startUDP(ctx, port)
	if err != nil {
		return err
	}
	defer close(exchange.toVpnC)

	wg.Go(func() { restWhoisService(ctx) })

	xlog.Trace.Println("start exchange", MyId)
	defer cancel()
	defer xlog.Trace.Println("stopping exchange", MyId, "...")

selection:
	for err == nil {
		select {
		case <-ctx.Done():
			break selection
		case <-alarm:
			xlog.Info = xlog.ToggleMute(xlog.Info)
			xlog.Trace = xlog.Mute(xlog.Trace)
		case err = <-rest.fault:
		case sub, ok := <-rest.whoisRspC:
			if !ok {
				break selection
			}
			exchange.sub[sub.Id.Index()] = sub
			xlog.Trace.Println("guest", sub)
		case m, ok := <-exchange.fromVpnC:
			if !ok {
				break selection
			}
			if len(m.Data) < SizeofLabel {
				mp.Put(m)
				xlog.Errata.Print("underrun")
			} else {
				exchangeFromVpn(ctx, m)
			}
		}
	}
	return err
}

func exchangeFromVpn(ctx context.Context, m *xnet.Msg) {
	tid, fid := ScanLabel(m.Data)
	ti, fi := tid.Index(), fid.Index()

	if from, fok := exchange.sub[fi]; !fok ||
		from.Id.Version() != fid.Version() {
		restQueueWhois(ctx, fi)
		xlog.Trace.Println("whois from", fi)
	} else if fid == tid {
		if from.helloIsOK(m) {
			from.ap = unmap4in6(m.AddrPort)
			if hello := newGreeting(0); hello != nil {
				xlog.Trace.Println("hello reply", from)
				hello.AddrPort = from.ap
				mp.Queue(ctx, exchange.toVpnC, hello)
			}
		}
	} else if to, tok := exchange.sub[ti]; !tok ||
		to.Id.Version() != tid.Version() {
		restQueueWhois(ctx, ti)
		xlog.Trace.Println("whois to", ti)
	} else if !to.ap.IsValid() {
		xlog.Trace.Println("dropped", from.name(), "-> unaddressed",
			to.name())
	} else {
		m.AddrPort = to.ap
		mp.Queue(ctx, exchange.toVpnC, m)
		xlog.Trace.Println("forward", from.name(), "->", to.name())
		return
	}
	mp.Put(m)
}
