// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"flag"
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet"
)

var exchange struct {
	sub map[int]*Subscriber
}

// Exchange is a UDP server that forwards ciphered packets between guest's.
func Exchange(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
Exchange ciphered packets between guests.

{{flags .}}`)

	definePort()
	defineRestFlags()
	enableTrace()
	enableQuiet()
	enableVerbose()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer xlog.Trace.Println("stopped")
	defer wg.Wait()

	if err = restInit(); err != nil {
		return err
	}

	exchange.sub = make(map[int]*Subscriber)

	if err = restExchangeCheckin(ctx); err != nil {
		return err
	}

	if err = udpInit(vpnPort); err != nil {
		return err
	}
	defer udp.Close()
	wg.Go(udpStream)

	wg.Go(func() { xlog.AlarmHandler(ctx) })

	defer close(rest.whoisReqC)
	wg.Go(func() { restWhoisService(ctx) })

	svc := fmt.Sprintf("%v @ %v", MyId, udp.LocalAddr())
	xlog.Trace.Println("start", svc)
	defer xlog.Trace.Println("stopping", svc, "...")

	defer cancel()

selection:
	for err == nil {
		select {
		case <-ctx.Done():
			break selection
		case err = <-rest.fault:
		case sub, ok := <-rest.whoisRspC:
			if !ok {
				break selection
			}
			exchange.sub[sub.Id.Index()] = sub
			xlog.Trace.Println("guest", sub)
		case m, ok := <-udpC:
			if !ok {
				break selection
			}
			if len(m.Data) < SizeofLabel {
				xlog.Errata.Print("underrun")
			} else {
				exchangeFromUDP(ctx, m)
			}
			mp.Put(m)
		}
	}
	return err
}

func exchangeFromUDP(ctx context.Context, m *xnet.Msg) {
	tid, fid := ScanLabel(m.Data)

	fi := fid.Index()
	from, fok := exchange.sub[fi]
	if !fok || from.Id.Version() != fid.Version() {
		restQueueWhois(ctx, fi)
		xlog.Trace.Println("whois from", fi)
		return
	}

	if fid == tid {
		from.hellohello(ctx, m)
		return
	}

	ti := tid.Index()
	to, tok := exchange.sub[ti]
	if !tok || to.Id.Version() != tid.Version() {
		restQueueWhois(ctx, ti)
		xlog.Trace.Println("whois to", ti)
		return
	}
	if !to.via.IsValid() {
		xlog.Trace.Println("dropped from", from.name(),
			"to unaddressed", to.name())
		return
	}
	udp.WriteToUDPAddrPort(m.Data, to.via)
	xlog.Trace.Println("forward from", from.name(), "to", to.name())
}
