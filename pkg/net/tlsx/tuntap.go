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

	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/net/tuntap"
	"github.com/platinasystems/goes/v2/pkg/os/page"
)

const (
	ttDown = false
	ttUp   = true
)

var MaxNonce = big.NewInt(math.MaxInt64)

// Path to iproute2 command if available.
var iproute2 string

func TunTap(
	ctx context.Context,
	path []string,
	args ...string,
) (err error) {
	const usage = `{{$path := join .Path " "}}{{/*
*/}}usage: {{$path}} [<options>] <exchange>
Open tap to named exchange or self @ given address.

<exchange>
	[<name>][@<dns|ip4|\[ip6\]>][:<port>]
{{SprintDefault .Flags}}`
	defer egress.Recovery(&err)
	var addr net.IP
	fs := flag.NewSilentFlagSet("tuntap")
	fs.TextVar(&addr, "a", addr, "static network address")
	randll := fs.Bool("r", false,
		"use random link address instead of hashed cert SKI")
	unit := fs.Uint("u", 0, "unit number")
	if flag.Search[bool]("complete") {
		return
	}
	err = fs.Parse(args)
	if err != nil {
		return
	}
	if flag.Search[bool]("help", fs) {
		err = style.Usage(usage, struct {
			Path  []string
			Flags *flag.FlagSet
		}{path, fs})
		return
	}
	if args = fs.Args(); len(args) == 0 {
		panic(ErrIncomplete)
	}

	ex := args[0]

	if s, err := exec.LookPath("ip"); err == nil {
		iproute2 = s
	}

	ha := netif.NewHardwareAddr()
	if *randll {
		if err = ha.Rand(); err != nil {
			panic(err)
		}
	} else {
		hash := fnv.New64()
		hash.Write(Self().Certificate.Leaf.SubjectKeyId)
		fmt.Fprint(hash, *unit)
		copy(ha, hash.Sum(nil))
		ha.Unicast()
	}

	network := 3
	if path[len(path)-1] == "tap" {
		if !tuntap.CanTAP {
			panic(ErrCantTap)
		}
		if tuntap.HasPI {
			panic(ErrHasPI)
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
		panic(err)
	}
	defer f.Close()

	ifname := f.Name()
	if err := ttAdmin(ifname, ttUp); err != nil {
		panic(err)
	}

	for ctx.Err() == nil {
		if err := ttSetState(ctx, ifname, ttDown); err != nil {
			panic(err)
		}

		tkey, err := ecdh.X25519().GenerateKey(rand.Reader)
		if err != nil {
			panic(err)
		}
		tnonce, err := rand.Int(rand.Reader, MaxNonce)
		if err != nil {
			panic(err)
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
					panic(err)
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
			panic(err)
		}

		got := new(strings.Builder)
		/*FIXME
		if err = Exec(ctx, tlsc, nil, got, join...); err != nil {
			tlsc.Close()
			panic(err)
		}
		*/
		prefix, err := netip.ParsePrefix(got.String())
		if err != nil {
			tlsc.Close()
			panic(err)
		}
		if err = ttSetPrefix(ifname, prefix); err != nil {
			tlsc.Close()
			panic(err)
		}
		if err = ttSetState(ctx, ifname, ttUp); err != nil {
			tlsc.Close()
			panic(err)
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
	return
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
	defer style.Recovery()
	pg := page.New()
	defer page.Free(pg)
	enc := lv.NewEncoder(write.With(ctx, tlsc))
	p := poll.WithReader(ctx, f)
	for {
		n, err := p.Read(pg)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				panic(err)
			}
			break
		}
		/*FIXME
		if n < tuntap.TapMin {
			panic(ErrTooShort)
			break
		}
		*/
		if _, err = enc.Write(pg[:n]); err != nil {
			if !errors.Is(err, context.Canceled) {
				panic(err)
			}
			break
		}
		// FIXME style.Println(host, "->", x, frame.NewEth(pg[:n]))
	}
}

func ttWrite(
	ctx context.Context,
	tlsc *tls.Conn,
	f *os.File,
	host, x string,
) {
	defer style.Recovery()

	pg := page.New()
	defer page.Free(pg)

	dec := lv.NewDecoder(poll.WithReader(ctx, tlsc))
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
		// FIXME style.Println(host, "<-", x, frame.NewEth(pg[:n]))
	}
}
