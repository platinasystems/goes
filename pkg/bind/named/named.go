// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	named_conf "github.com/platinasystems/goes/v2/pkg/bind/named-conf"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsdb"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

var errata, verbose xlog.WritePrinter
var conf named_conf.Conf
var opt = map[string]string{
	"cert":      "/etc/named.crt",
	"key":       "/etc/named.key",
	"endpoints": "/dns/query",
}

type xlistener interface {
	net.Listener
	SetDeadline(time.Time) error
}

func Named(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
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
		conf, err = named_conf.NewConf(flags.c)
		if err != nil {
			if !os.IsNotExist(err) || flags.c != defaultNamedConf {
				return err
			}
		}
	}
	if flags.C {
		fmt.Print(conf)
		return nil
	}

	for _, s := range strings.Split(flags.T, ",") {
		eq := strings.Index(s, "=")
		if eq < 0 {
			opt[s] = "true"
		} else {
			opt[s[:eq]] = s[eq+1:]
		}
	}

	for _, s := range strings.Split(opt["endpoints"], ",") {
		http.HandleFunc(s, httpHandler)
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

	host := ":"
	tcpNW := "tcp"
	udpNW := "udp"
	if flags.ip4only {
		host = "0.0.0.0:"
		tcpNW = "tcp4"
		udpNW = "udp4"
	} else if flags.ip6only {
		host = "[::]:"
		tcpNW = "tcp6"
		udpNW = "udp6"
	}

	for _, s := range strings.Split(flags.p, ",") {
		if strings.HasPrefix(s, "http=") {
			laddr := host + strings.TrimPrefix(s, "http=")
			srv := &http.Server{Addr: laddr}
			wg.Add(1)
			go httpShutdown(ctx, wg, srv)
			wg.Add(1)
			go httpListenAndServe(wg, srv)
		} else if strings.HasPrefix(s, "https=") {
			laddr := host + strings.TrimPrefix(s, "https=")
			srv := &http.Server{Addr: laddr}
			wg.Add(1)
			go httpShutdown(ctx, wg, srv)
			wg.Add(1)
			go httpListenAndServe(wg, srv, opt["cert"], opt["key"])
		} else if strings.HasPrefix(s, "tls=") {
			laddr := ":" + strings.TrimPrefix(s, "tls=")
			ca, err := tls.LoadX509KeyPair(opt["cert"], opt["key"])
			if err != nil {
				return err
			}
			cfg := &tls.Config{
				MinVersion:   tls.VersionTLS13,
				ClientAuth:   tls.RequestClientCert,
				Certificates: []tls.Certificate{ca},
			}
			ln, err := tls.Listen(tcpNW, laddr, cfg)
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
		} else {
			laddr := ":" + strings.TrimPrefix(s, "dns=")
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

// https://datatracker.ietf.org/doc/html/rfc8484
func httpHandler(rsp http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	var reqb []byte
	var err error
	switch req.Method {
	case http.MethodGet:
		dnsq := req.URL.Query().Get("dns")
		reqb, err = base64.StdEncoding.DecodeString(dnsq)
	case http.MethodPost:
		reqb, err = io.ReadAll(req.Body)
	default:
		rsp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err != nil {
		rsp.WriteHeader(http.StatusBadRequest)
		return
	}
	reqm := xdnsmessage.NewMessage()
	defer reqm.Free()
	if err = reqm.UnmarshalBinary(reqb); err != nil {
		rsp.WriteHeader(http.StatusBadRequest)
		return
	}
	rspm := answer(reqm)
	defer rspm.Free()
	rsp.Header().Set("Content-Type", "application/dns-message")
	// FIXME Message needs a WriteTo
	rspb, err := rspm.AppendTo(make([]byte, 0, 4<<10))
	if err != nil {
		errata.Print(err)
		rsp.WriteHeader(http.StatusInternalServerError)
	} else if _, err = rsp.Write(rspb); err != nil {
		errata.Print(err)
	}
}

func httpListenAndServe(wg *sync.WaitGroup, srv *http.Server, fns ...string) {
	var err error
	defer wg.Done()
	defer verbose.Println("stopped http", srv.Addr, "service:", err)
	verbose.Println("start http", srv.Addr, "service")
	if len(fns) == 2 {
		err = srv.ListenAndServeTLS(fns[0], fns[1])
	} else {
		err = srv.ListenAndServe()
	}
}

func httpShutdown(
	ctx context.Context,
	wg *sync.WaitGroup,
	srv *http.Server,
) {
	const timeout = 3 * time.Second
	defer wg.Done()
	defer verbose.Print("http", srv.Addr, "done")
	<-ctx.Done()
	cctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	verbose.Print("shutdown http", srv.Addr, "...")
	srv.Shutdown(cctx)
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
		n, err := conn.Read(data[:2])
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			} else if !errors.Is(err, io.EOF) {
				errata.Print(err)
			}
			break
		}
		if n != 2 {
			errata.Print("underrun")
			break
		}
		err = conn.SetReadDeadline(time.Time{})
		if err != nil {
			errata.Print(err)
			break
		}
		n = int(binary.BigEndian.Uint16(data))
		n, err = conn.Read(data[:n])
		if err != nil {
			errata.Print(err)
			break
		}
		if err = req.UnmarshalBinary(data[:n]); err != nil {
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
