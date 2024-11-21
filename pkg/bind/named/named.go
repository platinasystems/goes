// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named

import (
	"context"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
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

var errata, verbose xlog.WritePrinter
var opt map[string]string

type xlistener interface {
	net.Listener
	SetDeadline(time.Time) error
}

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

	mutable := log.New(os.Stdout, "", log.Lshortfile)
	if flags.v {
		verbose = xlog.Unmute(mutable)
		verbose.Println("start named")
		defer verbose.Println("stopped named")
	} else if flags.q {
		errata = xlog.Mute(mutable)
	}

	if len(flags.c) > 0 {
		if f, err := os.Open(flags.c); err == nil {
			// FIXME parse config
			f.Close()
		} else if !os.IsNotExist(err) || flags.c != defaultNamedConf {
			return xerrors.Mark(err)
		}
	}

	opt = make(map[string]string)
	for _, s := range strings.Split(flags.T, ",") {
		eq := strings.Index(s, "=")
		if eq < 0 {
			opt[s] = "true"
		} else {
			opt[s[:eq]] = s[eq+1:]
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	wg := new(sync.WaitGroup)
	defer func() {
		cancel()
		wg.Wait()
	}()

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

	for _, s := range strings.Split(flags.p, ",") {
		if strings.HasPrefix(s, "tls=") {
			return xerrors.Unsupported("tls")
		} else if strings.HasPrefix(s, "https=") {
			return xerrors.FIXME("https")
		} else if strings.HasPrefix(s, "http=") {
			return xerrors.FIXME("http")
		} else {
			laddr := ":" + strings.TrimPrefix(s, "dns=")
			tcpNW := "tcp"
			udpNW := "udp"
			if flags.ip4only {
				tcpNW = "tcp4"
				udpNW = "udp4"
			} else if flags.ip6only {
				tcpNW = "tcp6"
				udpNW = "udp6"
			}
			if opt["notcp"] != "true" {
				ln, err := net.Listen(tcpNW, laddr)
				if err != nil {
					return err
				}
				defer ln.Close()
				xln, ok := ln.(xlistener)
				if !ok {
					return errors.New("can't set deadline")
				}
				wg.Add(1)
				go tcpAccept(ctx, wg, xln)
			}
			udpConn, err := net.ListenPacket(udpNW, laddr)
			if err != nil {
				return err
			}
			defer udpConn.Close()
			mch := make(chan *xdnsmessage.Message, 4)
			wg.Add(1)
			go udpReceive(ctx, wg, udpConn, mch)
			wg.Add(1)
			go udpService(ctx, wg, udpConn, mch)
		}
	}

	wg.Wait()
	return err
}

func tcpAccept(
	ctx context.Context,
	wg *sync.WaitGroup,
	xln xlistener,
) {
	var err error
	la := xln.Addr()
	defer verbose.Println("stopped tcp", la, "accept:", err)
	defer wg.Done()
	verbose.Println("start tcp", la, "accept")
	for {
		if err = ctx.Err(); err != nil {
			break
		}
		err = xln.SetDeadline(time.Now().Add(time.Second))
		if err != nil {
			break
		}
		conn, err := xln.Accept()
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			}
			break
		}
		wg.Add(1)
		go tcpService(ctx, wg, conn)
	}
}

func tcpService(
	ctx context.Context,
	wg *sync.WaitGroup,
	conn net.Conn,
) {
	ra := conn.RemoteAddr()
	defer verbose.Println("stopped tcp", ra, "service")
	defer wg.Done()
	defer conn.Close()
	data := make([]byte, 4<<10, 4<<10)
	req := xdnsmessage.NewMessage()
	defer req.Free()
	verbose.Println("start tcp", ra, "service")
	for {
		err := ctx.Err()
		if err != nil {
			break
		}
		err = conn.SetReadDeadline(time.Now().Add(time.Second))
		if err != nil {
			errata.Print(err)
			break
		}
		n, err := conn.Read(data)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			} else if !errors.Is(err, io.EOF) {
				errata.Print(err)
			}
			break
		}
		if hn := binary.BigEndian.Uint16(data); n != 2+int(hn) {
			errata.Print("underrun")
			break
		}
		if err = req.UnmarshalBinary(data[2:n]); err != nil {
			errata.Print(err)
			break
		}
		ans := answer(req)
		b, err := ans.AppendTo(data[2:2])
		if err != nil {
			errata.Print(err)
			ans.Free()
			break
		}
		n = len(b)
		binary.BigEndian.PutUint16(data, uint16(n))
		if _, err = conn.Write(data[:2+n]); err != nil {
			errata.Print(err)
		}
		ans.Free()
	}
}

func udpReceive(
	ctx context.Context,
	wg *sync.WaitGroup,
	conn net.PacketConn,
	ch chan<- *xdnsmessage.Message,
) {
	defer verbose.Println("stopped udp receiver")
	defer wg.Done()
	defer close(ch)
	data := make([]byte, 4<<10, 4<<10)
	verbose.Println("start udp receiver")
	for {
		err := ctx.Err()
		if err != nil {
			break
		}
		err = conn.SetReadDeadline(time.Now().Add(time.Second))
		if err != nil {
			errata.Print(err)
			break
		}
		n, a, err := conn.ReadFrom(data)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			}
			errata.Print(err)
			break
		}
		req := xdnsmessage.NewMessage()
		if err = req.UnmarshalBinary(data[:n]); err != nil {
			errata.Print(err)
			req.Free()
		} else {
			req.Addr = a
			ch <- req
		}
	}
}

func udpService(
	ctx context.Context,
	wg *sync.WaitGroup,
	conn net.PacketConn,
	ch <-chan *xdnsmessage.Message,
) {
	defer verbose.Println("stopped udp message service")
	defer wg.Done()
	verbose.Println("start udp message service")
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
	ans := xdnsmessage.NewMessage()
	ans.Addr = req.Addr
	ans.ID = req.ID
	ans.HF = xdnsmessage.HFResponse
	ans.OpCode = req.OpCode
	ans.Questions = req.Questions
	if len(req.Questions) == 0 {
		ans.RCode = xdnsmessage.RCodeFormatError
	} else if req.OpCode != xdnsmessage.OpCodeQuery {
		ans.RCode = xdnsmessage.RCodeNotImplemented
	} else {
		name := req.Questions[0].Name
		if entries := xdnsdb.Get(name); len(entries) == 0 {
			ans.RCode = xdnsmessage.RCodeNameError
		} else {
			ans.RCode = xdnsmessage.RCodeSuccess
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
				ans.Answers = append(ans.Answers,
					xdnsmessage.WireResource{
						Name:          name,
						Duration:      dur,
						Class:         entry.Class,
						TypedResource: r,
					})
			}
		}
	}
	return ans
}
