// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tap

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"hash/fnv"
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
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/net/tuntap"
	"github.com/platinasystems/goes/v2/pkg/os/page"
)

const (
	Down = false
	Up   = true
)

type T struct {
	Exchange string
	*os.File
	Prefix   netip.Prefix
	iproute2 string
}

func (t *T) Configure(ctx context.Context, args []string) ([]string, error) {
	fs := flags.New()
	fs.TextVar(&t.Prefix, "p", t.Prefix,
		"prefix [ip/bits] (default dynamic lease)")
	uflag := fs.Uint("u", 0, "interface unit suffix")
	xflag := fs.String("x", certs.Self.Name(), "exchange")
	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}
	t.Exchange = *xflag

	if t.iproute2, err = exec.LookPath("ip"); err != nil {
		t.iproute2 = ""
		err = nil
	}

	ha := make(net.HardwareAddr, 6)
	if t.Prefix.IsValid() {
		h := fnv.New64()
		h.Write(t.Prefix.Addr().AsSlice())
		h.Write(certs.Self.TLS().Leaf.SubjectKeyId)
		copy(ha, h.Sum(nil))
	} else {
		copy(ha, certs.Self.TLS().Leaf.SubjectKeyId)
	}
	ha[0] &^= 1

	if !tuntap.CanTAP {
		return args, fmt.Errorf("%s can't TAP", runtime.GOOS)
	}
	if tuntap.HasPI {
		return args, fmt.Errorf("%s's has unwanted packet info",
			runtime.GOOS)
	}

	if t.File, err = tuntap.New(&tuntap.Configuration{
		Unit:  *uflag,
		IsTap: true,
		Link:  tuntap.Link{ha},
	}); err != nil {
		return nil, fmt.Errorf("new: %w", err)
	}

	if t.Prefix.IsValid() {
		if err = t.setPrefix(ctx); err != nil {
			t.Close()
			t.File = nil
			return nil, fmt.Errorf("prefix: %w", err)
		}
	}

	return fs.Args(), nil
}

func (t *T) Routine(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer t.File.Close()
	defer t.admin(ctx, Down)

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
				if errors.Is(err, syscall.ECONNREFUSED) {
					return
				}
			}
			t.sleep(ctx, 3*time.Second)
			continue
		}

		la := conn.LocalAddr().String()
		xname := dns0(conn)

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
		go t.reader(cctx, &rwg, conn, la, xname)
		t.writer(ctx, conn, la, xname)
		cancel()
		rwg.Wait()
		conn.Close()
	}
}

func (t *T) admin(ctx context.Context, updown bool) error {
	if updown {
		return netif.Up(t.Name())
	} else {
		return netif.Down(t.Name())
	}
}

func (t *T) setPrefix(ctx context.Context) error {
	bc := netip.IPv4Unspecified()
	// or bc := netip.AddrFrom4([4]byte{255, 255, 255, 255})
	return netif.Add(t.Name(), t.Prefix, bc)
}

func (t *T) State(ctx context.Context, updown bool) error {
	// FIXME w/ netlink
	if true || len(t.iproute2) == 0 {
		return nil
	}
	s := map[bool]string{
		false: "LOWERLAYERDOWN",
		true:  "UP",
	}[updown]
	out, err := exec.CommandContext(ctx, t.iproute2, "link", "set",
		t.Name(), "state", s).CombinedOutput()
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
		style.Println(host, "->", x, frame.NewEth(pg[:n]))
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
		style.Println(host, "<-", x, frame.NewEth(pg[:n]))
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
