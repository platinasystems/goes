// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsdb"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

var (
	mutable = log.New(os.Stdout, "", log.Lshortfile)
	errata  = xlog.Unmute(mutable)
	verbose = xlog.Mute(mutable)
)

func Named(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [[-flags]
Mimic BIND9's Internet domain name daemon.

{{flags .}}`)
	addCommandLineFlags()
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if flags.V {
		version := "(unavailable)"
		if mm := xprogram.MainModule(); mm != nil {
			version = mm.Version
		}
		fmt.Println(version)
		return nil
	}
	if flags.v {
		verbose = xlog.Unmute(verbose)
		verbose.Println("start named")
		defer verbose.Println("stopped named")
	} else if flags.q {
		errata = xlog.Mute(errata)
	}
	if len(flags.c) > 0 {
		if f, err := os.Open(flags.c); err == nil {
			// FIXME parse config
			f.Close()
		} else if !os.IsNotExist(err) || flags.c != defaultNamedConfig {
			return xerrors.Mark(err)
		}
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go xdnsdb.Routine(ctx, wg, verbose)

	if len(flags.Z) > 0 {
		for _, fn := range strings.Split(flags.Z, ",") {
			err = xdnsdb.Include(ctx, flags.z, fn)
			if err != nil {
				return err
			}
		}
	}
	if flags.C {
		xdnsdb.Dump(os.Stdout)
		return nil
	}
	if flags.p != 0 {
		nw := "udp"
		laddr := fmt.Sprintf(":%d", flags.p)
		if flags.ip4only {
			nw = "udp4"
		} else if flags.ip6only {
			nw = "udp6"
		}
		pc, err := net.ListenPacket(nw, laddr)
		if err != nil {
			return err
		}
		ch := make(chan *xdnsmessage.Message, 4)
		wg.Add(1)
		go pktrcv(ctx, wg, pc, ch)
		wg.Add(1)
		go pktsvc(ctx, wg, pc, ch)
	}
	if flags.P != 0 {
		// FIXME DOH
	}

	wg.Wait()
	return nil
}

func pktrcv(
	ctx context.Context,
	wg *sync.WaitGroup,
	conn net.PacketConn,
	ch chan<- *xdnsmessage.Message,
) {
	verbose.Println("start pktrcv")
	defer verbose.Println("stopped pktrcv")
	defer wg.Done()
	var n int
	var a net.Addr
	data := make([]byte, 4<<10, 4<<10)
	for err := ctx.Err(); err == nil; err = ctx.Err() {
		err = conn.SetReadDeadline(time.Now().Add(time.Second))
		if err != nil {
			errata.Print(err)
		} else if n, a, err = conn.ReadFrom(data); err == nil {
			m := xdnsmessage.NewMessage()
			err = m.UnmarshalBinary(data[:n])
			if err != nil {
				errata.Print(err)
				m.Free()
			} else {
				m.Addr = a
				ch <- m
			}
		} else if errors.Is(err, os.ErrDeadlineExceeded) {
			err = nil
		} else {
			errata.Print(err)
		}
	}
}

func pktsvc(
	ctx context.Context,
	wg *sync.WaitGroup,
	conn net.PacketConn,
	ch <-chan *xdnsmessage.Message,
) {
	verbose.Println("start pktsvc")
	defer verbose.Println("stopped pktsvc")
	defer wg.Done()
	data := make([]byte, 4<<10, 4<<10)
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-ch:
			ans := answer(req)
			if b, err := ans.AppendTo(data[:0]); err != nil {
				errata.Print(err)
			} else if _, err = conn.
				WriteTo(b, ans.Addr); err != nil {
				errata.Print(err)
			}
			req.Free()
			ans.Free()
		}
	}
}

func answer(req *xdnsmessage.Message) *xdnsmessage.Message {
	now := time.Now()
	rsp := xdnsmessage.NewMessage()
	rsp.Addr = req.Addr
	rsp.ID = req.ID
	rsp.HF = xdnsmessage.HFResponse
	rsp.OpCode = req.OpCode
	rsp.Questions = req.Questions
	if len(req.Questions) == 0 {
		rsp.RCode = xdnsmessage.RCodeFormatError
	} else if req.OpCode != xdnsmessage.OpCodeQuery {
		rsp.RCode = xdnsmessage.RCodeNotImplemented
	} else {
		name := req.Questions[0].Name
		if entries := xdnsdb.Get(name); len(entries) == 0 {
			rsp.RCode = xdnsmessage.RCodeNameError
		} else {
			rsp.RCode = xdnsmessage.RCodeSuccess
			c := req.Questions[0].Class
			t := req.Questions[0].Type
			for _, entry := range entries {
				if entry.Class != c &&
					c != xdnsmessage.ClassANY {
					continue
				}
				if entry.Type() != t &&
					t != xdnsmessage.TypeANY {
					continue
				}
				if now.After(entry.Deadline) {
					continue
				}
				dur := entry.Deadline.Sub(now)
				r := entry.TypedResource
				rsp.Answers = append(rsp.Answers,
					xdnsmessage.WireResource{
						Name:          name,
						Duration:      dur,
						Class:         entry.Class,
						TypedResource: r,
					})
			}
		}
	}
	return rsp
}
