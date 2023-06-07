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
	"html/template"
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

	"github.com/platinasystems/goes/v2/pkg/container/slice"
	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/flag/flags"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/frame"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/greet"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/service"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/net/tuntap"
	"github.com/platinasystems/goes/v2/pkg/os/page"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

const Usage = `
usage: {{.Command}} [<options>] <exchange>
Open tap to named exchange or self @ given address.

<exchange>
	[<name>][@<dns|ip4|\[ip6\]>][:<port>]
{{print .Flags}}`

const (
	down = false
	up   = true
)

var (
	ErrIncomplete = errors.New("incomplete")
	ErrCantTap    = errors.New(runtime.GOOS + " can't TAP")
	ErrHasPI      = errors.New(runtime.GOOS + " has unwanted packet info")
	ErrTooShort   = errors.New("too short")
)

// Path to iproute2 command if available.
var iproute2 string

// The Tap interface address (aka. MAC) is a hash of the unit number (order of
// creation) and the subject-key-id of the Self certificate.
func Daemon(
	ctx context.Context,
	path []string,
	args ...string,
) error {
	var addr net.IP
	fs := flags.New()
	port := fs.Uint("p", 0, "non-zero service port")
	fs.TextVar(&addr, "a", addr, "static tap address")

	usage := func() error {
		return template.Must(template.New("usage").Parse(Usage[1:])).
			Execute(style.Plain.Notice.Writer(), struct {
				Command string
				Flags   flags.Flags
			}{
				strings.Join(path, " "),
				fs,
			})
	}

	switch path[1] {
	case "complete":
		return nil
	case "help":
		path = slice.Cut[string](path, 1, 1)
		path[1] = "start" // replace "daemon"
		return usage()
	default:
		if !program.IsKoApp() {
			style.System()
		}
	}

	err := fs.Parse(args)
	if err == flags.ErrHelp {
		return usage()
	} else if err != nil {
		return err
	}
	if args = fs.Args(); len(args) == 0 {
		return egress.Marked(ErrIncomplete)
	}
	ex := args[0]

	if !tuntap.CanTAP {
		return egress.Marked(ErrCantTap)
	}
	if tuntap.HasPI {
		return egress.Marked(ErrHasPI)
	}
	if s, err := exec.LookPath("ip"); err == nil {
		iproute2 = s
	}

	self, err := certs.Self.TLS()
	if err != nil {
		return egress.Marked(err)
	}

	hash := fnv.New64()
	hash.Write(self.Leaf.SubjectKeyId)

	cfg := &tuntap.Configuration{
		Unit:  0,
		IsTap: true,
		Link: tuntap.Link{
			HardwareAddr: make(net.HardwareAddr, 6),
		},
	}

	copy(cfg.Link.HardwareAddr, hash.Sum(nil))
	cfg.Link.HardwareAddr[0] &^= 1 // mask muti-cast

	join := []any{"join", ""}
	if addr.IsUnspecified() {
		join = join[:1]
	} else {
		join[1] = addr.String()
	}

	f, err := tuntap.New(cfg)
	if err != nil {
		return egress.Marked(err)
	}
	style.Noteln(ex, join)

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go routine(ctx, &wg, f, ex, join)
	if *port != 0 {
		wg.Add(1)
		go service.Routine(ctx, &wg, *port)
	}
	return nil
}

func routine(
	ctx context.Context,
	wg *sync.WaitGroup,
	f *os.File,
	ex string,
	join []any,
) {
	defer wg.Done()
	defer f.Close()
	defer style.ShortFile.Errata.Recovery()

	ifname := f.Name()
	if err := admin(ifname, up); err != nil {
		panic(err)
	}

	for ctx.Err() == nil {
		if err := setState(ctx, ifname, down); err != nil {
			panic(err)
		}

		conn, err := tlsx.Connect(ctx, ex)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if !errors.Is(err, fs.ErrNotExist) {
				if errors.Is(err, syscall.ECONNREFUSED) {
					panic(err)
				}
			}
			fsleep(ctx, f, 3*time.Second)
			continue
		}

		la := conn.LocalAddr().String()
		xname := dns0(conn)

		tlsc, err := greet.Server(ctx, ex, conn)
		if err != nil {
			panic(egress.Unmarked(err, conn.Close))
		}

		got := new(strings.Builder)
		if err = tlsx.Exec(ctx, tlsc, nil, got, join...); err != nil {
			panic(egress.Unmarked(err, tlsc.Close))
		}
		prefix, err := netip.ParsePrefix(got.String())
		if err != nil {
			panic(egress.Unmarked(err, tlsc.Close))
		}
		if err = setPrefix(ifname, prefix); err != nil {
			panic(egress.Unmarked(err, tlsc.Close))
		}
		if err = setState(ctx, ifname, up); err != nil {
			panic(egress.Unmarked(err, tlsc.Close))
		}

		cctx, cancel := context.WithCancel(ctx)

		var rwg sync.WaitGroup
		rwg.Add(1)
		go fread(cctx, &rwg, f, tlsc, la, xname)
		fwrite(ctx, tlsc, f, la, xname)
		cancel()
		rwg.Wait()
		tlsc.Close()
	}
}

func admin(ifname string, updown bool) error {
	if updown {
		return netif.Up(ifname)
	} else {
		return netif.Down(ifname)
	}
}

func setPrefix(ifname string, prefix netip.Prefix) error {
	bc := netip.IPv4Unspecified()
	// or bc := netip.AddrFrom4([4]byte{255, 255, 255, 255})
	return netif.Add(ifname, prefix, bc)
}

func setState(ctx context.Context, ifname string, updown bool) error {
	// FIXME w/ netlink
	if true || len(iproute2) == 0 {
		return nil
	}
	s := map[bool]string{
		false: "LOWERLAYERDOWN",
		true:  "UP",
	}[updown]
	out, err := exec.CommandContext(ctx, iproute2, "link", "set",
		ifname, "state", s).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s; %w", out, err)
	}
	return nil
}

func fsleep(ctx context.Context, f *os.File, dur time.Duration) {
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

func fread(
	ctx context.Context,
	wg *sync.WaitGroup,
	f *os.File,
	tlsc *tls.Conn,
	host, x string,
) {
	defer wg.Done()
	defer style.ShortFile.Errata.Recovery()
	pg := page.New()
	defer page.Free(pg)
	enc := lv.NewEncoder(write.With(ctx, tlsc))
	p := poll.With(ctx, f)
	for {
		n, err := p.Read(pg)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				panic(err)
			}
			break
		}
		if n < tuntap.TapMin {
			panic(ErrTooShort)
			break
		}
		if _, err = enc.Write(pg[:n]); err != nil {
			if !errors.Is(err, context.Canceled) {
				panic(err)
			}
			break
		}
		style.Println(host, "->", x, frame.NewEth(pg[:n]))
	}
}

func fwrite(
	ctx context.Context,
	tlsc *tls.Conn,
	f *os.File,
	host, x string,
) {
	defer style.ShortFile.Errata.Recovery()

	pg := page.New()
	defer page.Free(pg)

	dec := lv.NewDecoder(poll.With(ctx, tlsc))
	for {
		n, err := dec.Read(pg)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				panic(err)
			}
			break
		}
		if _, err = f.Write(pg[:n]); err != nil {
			if !errors.Is(err, context.Canceled) {
				panic(err)
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
