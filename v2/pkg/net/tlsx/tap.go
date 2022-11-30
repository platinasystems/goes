// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/frame"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/cert"
	"github.com/platinasystems/goes/v2/pkg/net/tuntap"
	"github.com/platinasystems/goes/v2/pkg/os/page"
)

func Tap(ctx context.Context, ex string, unit uint) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	if !tuntap.CanTAP {
		return fmt.Errorf("%s can't TAP", runtime.GOOS)
	}
	if tuntap.HasPI {
		return fmt.Errorf("%s's TAP includes unwanted packet info",
			runtime.GOOS)
	}

	addr, cfg, err := AddrCfg(ex)
	if err != nil {
		return err
	}

	host := cfg.Certificates[0].Leaf.DNSNames[0]

	ha := net.HardwareAddr(cert.Value().Leaf.SubjectKeyId[:6])
	ha[0] &^= 1

	f, err := tuntap.New(&tuntap.Configuration{
		Unit:  unit,
		IsTap: true,
		Link:  tuntap.Link{ha},
	})
	if err != nil {
		return err
	}
	defer f.Close()

	hf := fmt.Sprint(host, ":", f.Name())

	var dl net.Dialer

	for ctx.Err() == nil {
		tapcarrier(ctx, f, false)
		c, err := dl.DialContext(ctx, addr.Network(), addr.String())
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			if !errors.Is(err, fs.ErrNotExist) {
				style.Error(err)
			}
			if errors.Is(err, syscall.ECONNREFUSED) {
				return err
			}
			tapsleep(ctx, f, 3*time.Second)
			continue
		}

		cl := tls.Client(c, cfg)

		if err = cl.HandshakeContext(ctx); err != nil {
			cl.Close()
			if ctx.Err() == nil {
				return err
			}
			return nil
		}

		x := dns0(cl)

		got := new(strings.Builder)
		if err = Req(ctx, cl, nil, got, "join"); err != nil {
			cl.Close()
			if ctx.Err() != nil {
				return nil
			}
			continue
		}

		tapcarrier(ctx, f, true)

		cctx, cancel := context.WithCancel(ctx)
		wg.Add(1)
		go tapreader(cctx, &wg, cl, hf, x, f)
		tapwriter(ctx, cl, hf, x, f)
		cancel()
	}

	return nil
}

func tapcarrier(ctx context.Context, f *os.File, on bool) {
	// FIXME
}

func tapsleep(ctx context.Context, f *os.File, dur time.Duration) {
	var wg sync.WaitGroup
	defer wg.Wait()

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	p := poll.With(cctx, f)

	wg.Add(1)
	go func() {
		defer wg.Done()
		pg := page.New()
		defer page.Free(pg)
		for {
			if _, err := p.Read(pg); err != nil {
				break
			}
		}
	}()

	t := time.NewTimer(dur)
	for {
		select {
		case <-t.C:
			return
		case <-ctx.Done():
			if !t.Stop() {
				<-t.C
			}
			return
		}
	}
}

func tapreader(
	ctx context.Context,
	wg *sync.WaitGroup,
	tlsc *tls.Conn,
	host, x string,
	f *os.File,
) {
	defer wg.Done()
	pg := page.New()
	defer page.Free(pg)
	enc := lv.NewEncoder(write.With(ctx, tlsc))
	p := poll.With(ctx, f)
	for {
		n, err := p.Read(pg)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				style.Errorln(host, err)
			}
			break
		}
		if n < tuntap.TapMin {
			style.Error(host, "too short")
			break
		}
		if _, err = enc.Write(pg[:n]); err != nil {
			if !errors.Is(err, context.Canceled) {
				style.Errorln(host, "->", x, err)
			}
			break
		}
		style.Println(host, "->", x, frame.Eth(pg[:n]))
	}
}

func tapwriter(
	ctx context.Context,
	tlsc *tls.Conn,
	host, x string,
	f *os.File,
) {
	pg := page.New()
	defer page.Free(pg)
	dec := lv.NewDecoder(poll.With(ctx, tlsc))
	for {
		n, err := dec.Read(pg)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				style.Errorln(host, err)
			}
			break
		}
		if _, err = f.Write(pg[:n]); err != nil {
			if !errors.Is(err, context.Canceled) {
				style.Errorln(host, err)
			}
			break
		}
		style.Println(host, "<-", x, frame.Eth(pg[:n]))
	}
}

func dns0(c *tls.Conn) string {
	if cs := c.ConnectionState(); len(cs.PeerCertificates) > 0 &&
		len(cs.PeerCertificates[0].DNSNames) > 0 {
		return cs.PeerCertificates[0].DNSNames[0]
	}
	return "anonymous"
}
