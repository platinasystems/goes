// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tap

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/flag/flags"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/frame"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/net/tuntap"
	"github.com/platinasystems/goes/v2/pkg/os/page"
)

const (
	Down = false
	Up   = true
)

var ErrNoExchange = errors.New("missing <exchange>")

type T struct {
	Exchange string
	*os.File
	Prefix netip.Prefix
}

func (t *T) Configure(ctx context.Context, args []string) ([]string, error) {
	fs := flags.New()
	fs.TextVar(&t.Prefix, "prefix", t.Prefix,
		"ip/bits (default dynamic lease)")
	unit := fs.Uint("unit", 0, "interface suffix")
	err := fs.Parse(args)
	if err != nil {
		return []string{}, err
	}
	if args = fs.Args(); len(args) == 0 {
		return []string{}, ErrNoExchange
	}
	t.Exchange = args[0]
	args = args[1:]

	ha := net.HardwareAddr(certs.Self.SKI()[:6])
	ha[0] &^= 1

	if !tuntap.CanTAP {
		return args, fmt.Errorf("%s can't TAP", runtime.GOOS)
	}
	if tuntap.HasPI {
		return args, fmt.Errorf("%s's has unwanted packet info",
			runtime.GOOS)
	}

	if t.File, err = tuntap.New(&tuntap.Configuration{
		Unit:  *unit,
		IsTap: true,
		Link:  tuntap.Link{ha},
	}); err != nil {
		return args, err
	}

	if t.Prefix.IsValid() {
		if err = t.setPrefix(ctx); err != nil {
			t.Close()
			t.File = nil
			return args, err
		}
	}

	return args, err
}

func (t *T) Routine(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer t.File.Close()
	defer t.admin(ctx, Down)

	hf := fmt.Sprint(certs.Self.Name(), ":", t.File.Name())

	err := t.admin(ctx, Up)
	if err != nil {
		style.Error(err)
		return
	}

	for ctx.Err() == nil {
		if err = t.State(ctx, Down); err != nil {
			style.Error(err)
			return
		}

		conn, err := tlsx.Exchange.Connect(ctx, t.Exchange)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if !errors.Is(err, fs.ErrNotExist) {
				style.Error(err)
			}
			if errors.Is(err, syscall.ECONNREFUSED) {
				style.Error(err)
				return
			}
			t.sleep(ctx, 3*time.Second)
			continue
		}

		name := dns0(conn)

		got := new(strings.Builder)
		args := []any{"join"}
		if t.Prefix.IsValid() {
			args = append(args, t.Prefix.Addr())
		}
		if err = tlsx.Exec(ctx, conn, nil, got, args...); err != nil {
			conn.Close()
			style.Error(err)
			return
		}
		if !t.Prefix.IsValid() {
			t.Prefix, err = netip.ParsePrefix(got.String())
			if err == nil {
				err = t.setPrefix(ctx)
			}
			if err != nil {
				conn.Close()
				style.Error(err)
				return
			}
		}

		cctx, cancel := context.WithCancel(ctx)
		var rwg sync.WaitGroup

		if err = t.State(ctx, Up); err != nil {
			conn.Close()
			style.Error(err)
			return
		}

		rwg.Add(1)
		go t.reader(cctx, &rwg, conn, hf, name)
		t.writer(ctx, conn, hf, name)
		cancel()
		rwg.Wait()
		conn.Close()
	}
}

// FIXME change to a GO implementation of iproute2/ifconfig
func (t *T) admin(ctx context.Context, updown bool) error {
	s := map[bool]string{
		false: "down",
		true:  "up",
	}[updown]
	out, err := exec.CommandContext(ctx, "ip", "link", "set", t.Name(),
		s).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s; %w", out, err)
	}
	return nil
}

func (t *T) setPrefix(ctx context.Context) error {
	out, err := exec.CommandContext(ctx, "ip", "address", "add",
		t.Prefix.String(), "dev", t.Name()).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s; %w", out, err)
	}
	return nil
}

func (t *T) State(ctx context.Context, updown bool) error {
	s := map[bool]string{
		false: "LOWERLAYERDOWN",
		true:  "UP",
	}[updown]
	out, err := exec.CommandContext(ctx, "ip", "link", "set", t.Name(),
		"state", s).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s; %w", out, err)
	}
	return nil
}

func (t *T) sleep(ctx context.Context, dur time.Duration) {
	var wg sync.WaitGroup
	defer wg.Wait()

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	p := poll.With(cctx, t.File)

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

	timer := time.NewTimer(dur)
	for {
		select {
		case <-timer.C:
			return
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return
		}
	}
}

func (t *T) reader(
	ctx context.Context,
	wg *sync.WaitGroup,
	conn net.Conn,
	host, x string,
) {
	defer wg.Done()
	pg := page.New()
	defer page.Free(pg)
	enc := lv.NewEncoder(write.With(ctx, conn))
	p := poll.With(ctx, t.File)
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

func (t *T) writer(
	ctx context.Context,
	conn net.Conn,
	host, x string,
) {
	pg := page.New()
	defer page.Free(pg)
	dec := lv.NewDecoder(poll.With(ctx, conn))
	for {
		n, err := dec.Read(pg)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				style.Errorln(host, err)
			}
			break
		}
		if _, err = t.Write(pg[:n]); err != nil {
			if !errors.Is(err, context.Canceled) {
				style.Errorln(host, err)
			}
			break
		}
		style.Println(host, "<-", x, frame.Eth(pg[:n]))
	}
}

func dns0(conn net.Conn) string {
	if tlsc, ok := conn.(*tls.Conn); ok {
		cs := tlsc.ConnectionState()
		if len(cs.PeerCertificates) > 0 &&
			len(cs.PeerCertificates[0].DNSNames) > 0 {
			return cs.PeerCertificates[0].DNSNames[0]
		}
	}
	return "anonymous"
}
