// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"hash/fnv"
	"log"
	"math"
	"math/big"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/goes/v2/pkg/context/ctxparm"
	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/net/frame"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/net/tuntap"
	"github.com/platinasystems/goes/v2/pkg/os/page"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

const (
	ttDown = false
	ttUp   = true
)

var MaxNonce = big.NewInt(math.MaxInt64)
var EthHeader = frame.Header[frame.ETH]
var PiHeader = frame.Header[frame.TunPI]
var EthDump = func(...any) {}

// Path to iproute2 command if available.
var iproute2 string

const TunTapUsageTemplate = `
usage: {{.Path}} [<options>] <exchange>
Open tap to named exchange or self @ given address.

<exchange>
	[<name>][@<dns|ip4|\[ip6\]>][:<port>]
{{.Flag}}`

func TunTapUsageData(ctx context.Context) any {
	return struct{ Path, Flag string }{
		strings.Join(ctxparm.Strings.In(ctx), " "),
		ctxparm.SprintFlagsIn(ctx),
	}
}

func TunTap(ctx context.Context, args ...string) error {
	var addr net.IP
	if *complete.Help {
		return nil
	}
	flags := usage.NewFlags("tuntap")
	ctx = ctxparm.Flags.With(ctx, flags)
	flags.TextVar(&addr, "a", addr, "static network address")
	randll := flags.Bool("r", false,
		"use random link address instead of hashed cert SKI")
	unit := flags.Uint("u", 0, "unit number")
	err := flags.Parse(args)
	if err != nil {
		return err
	}
	if *usage.Help {
		return usage.Error(TunTapUsageTemplate[1:],
			TunTapUsageData(ctx))
	}
	if args = flags.Args(); len(args) == 0 {
		return ErrIncomplete
	}

	ex := args[0]

	if s, err := exec.LookPath("ip"); err == nil {
		iproute2 = s
	}

	ha := netif.NewHardwareAddr()
	if *randll {
		if err = ha.Rand(); err != nil {
			return err
		}
	} else {
		hash := fnv.New64()
		hash.Write(Self().Certificate.Leaf.SubjectKeyId)
		fmt.Fprint(hash, *unit)
		copy(ha, hash.Sum(nil))
		ha.Unicast()
	}

	network := 3
	path := ctxparm.Strings.In(ctx)
	if path[len(path)-1] == "tap" {
		if !tuntap.CanTAP {
			return ErrCantTap
		}
		if tuntap.HasPI {
			return ErrHasPI
		}
		network = 2
	}

	const (
		persist = false
		owner   = -1
		group   = -1
	)
	f, err := tuntap.New(*unit, network == 2, persist, owner, group, ha)
	if err != nil {
		return err
	}
	defer f.Close()

	ifname := f.Name()
	if err = ttAdmin(ifname, ttUp); err != nil {
		return err
	}

	for ctx.Err() == nil {
		if err = ttSetState(ctx, ifname, ttDown); err != nil {
			return err
		}

		tkey, err := ecdh.X25519().GenerateKey(rand.Reader)
		if err != nil {
			return err
		}
		tnonce, err := rand.Int(rand.Reader, MaxNonce)
		if err != nil {
			return err
		}

		confirmation, gcm, snonce, err := ttReserve(ctx, ex, network,
			tkey, tnonce, ha)

		_ = confirmation // FIXME
		_ = gcm
		_ = snonce

		conn, err := Connect(ctx, ex)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			if !errors.Is(err, os.ErrNotExist) {
				if errors.Is(err, syscall.ECONNREFUSED) {
					return err
				}
			}
			ttSleep(ctx, f, 3*time.Second)
			continue
		}

		la := conn.LocalAddr().String()
		xname := DNS0(conn)

		tlsc, err := greetServer(ctx, ex, conn)
		if err != nil {
			conn.Close()
			return err
		}

		got := new(strings.Builder)
		/*FIXME
		if err = Exec(ctx, tlsc, nil, got, join...); err != nil {
			tlsc.Close()
			return err
		}
		*/
		prefix, err := netip.ParsePrefix(got.String())
		if err != nil {
			tlsc.Close()
			return err
		}
		if err = ttSetPrefix(ifname, prefix); err != nil {
			tlsc.Close()
			return err
		}
		if err = ttSetState(ctx, ifname, ttUp); err != nil {
			tlsc.Close()
			return err
		}

		cctx, cancel := context.WithCancel(ctx)

		var rwg sync.WaitGroup
		rwg.Add(1)
		go ttRead(cctx, &rwg, f, tlsc, la, xname)
		ttWrite(ctx, tlsc, f, la, xname)
		cancel()
		rwg.Wait()
		tlsc.Close()
	}
	return nil
}

func ttAdmin(ifname string, up bool) error {
	if up {
		return netif.Up(ifname)
	} else {
		return netif.Down(ifname)
	}
}

func ttReserve(
	ctx context.Context,
	ex string,
	network int,
	tkey *ecdh.PrivateKey,
	tnonce *big.Int,
	ha netif.HardwareAddr,
) (
	confirmation string,
	gcm cipher.AEAD,
	snonce *big.Int,
	err error,
) {
	err = FIXME
	return
}

func ttSetPrefix(ifname string, prefix netip.Prefix) error {
	// bc := netip.IPv4Unspecified()
	// or bc := netip.AddrFrom4([4]byte{255, 255, 255, 255})
	// return netif.Add(ifname, prefix, bc)
	return FIXME
}

func ttSetState(ctx context.Context, ifname string, up bool) error {
	// FIXME w/ netlink
	if true || len(iproute2) == 0 {
		return nil
	}
	s := map[bool]string{
		false: "LOWERLAYERDOWN",
		true:  "UP",
	}[up]
	out, err := exec.CommandContext(ctx, iproute2, "link", "set",
		ifname, "state", s).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s; %w", out, err)
	}
	return nil
}

func ttSleep(ctx context.Context, f *os.File, dur time.Duration) {
	var wg sync.WaitGroup
	defer wg.Wait()

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	p := poll.WithReader(cctx, f)

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

func ttRead(
	ctx context.Context,
	wg *sync.WaitGroup,
	f *os.File,
	tlsc *tls.Conn,
	host, x string,
) {
	defer wg.Done()
	pg := page.New()
	defer page.Free(pg)
	enc := lv.NewEncoder(write.With(ctx, tlsc))
	p := poll.WithReader(ctx, f)
	pi := PiHeader(pg)
	for {
		n, err := p.Read(pg)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Print(err)
			}
			break
		}
		if n < frame.Sizeof(pi) {
			log.Print(ErrTooShort)
			break
		}
		if _, err = enc.Write(pg[:n]); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Print(err)
			}
			break
		}
		EthDump(host, "->", x, EthHeader(pg[:n]))
	}
}

func ttWrite(
	ctx context.Context,
	tlsc *tls.Conn,
	f *os.File,
	host, x string,
) {
	pg := page.New()
	defer page.Free(pg)

	dec := lv.NewDecoder(poll.WithReader(ctx, tlsc))
	for {
		n, err := dec.Read(pg)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Print(err)
			}
			break
		}
		if _, err = f.Write(pg[:n]); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Print(err)
			}
			break
		}
		EthDump(host, "<-", x, EthHeader(pg[:n]))
	}
}
