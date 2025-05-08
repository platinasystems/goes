// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bind

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	named_conf "github.com/platinasystems/goes/v2/pkg/bind/named-conf"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsdb"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
	"github.com/platinasystems/goes/v2/pkg/xsync"
)

const NamedDefaultConf = "/etc/named.conf"

type namedListener interface {
	net.Listener
	SetDeadline(time.Time) error
}

var namedWG xsync.WaitGroup

func Named(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [[-flags]
Mimic BIND9's Internet domain name daemon.

{{flags .}}`)

	defineNamedFlags()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	var conf named_conf.Conf

	opt := map[string]string{
		"cert":      "/etc/named.crt",
		"key":       "/etc/named.key",
		"endpoints": "/dns/query",
	}

	if named_V {
		version := "(unavailable)"
		if mm := xprogram.MainModule(); mm != nil {
			version = mm.Version
		}
		fmt.Println(version)
		return nil
	}

	if named_v {
		verbose = xlog.Unmute(mutable)
		verbose.Println("start named")
		defer verbose.Println("stopped named")
	} else if named_q {
		errata = xlog.Mute(mutable)
	}

	if len(named_c) > 0 {
		conf, err = named_conf.NewConf(named_c)
		if err != nil {
			if !os.IsNotExist(err) ||
				named_c != NamedDefaultConf {
				return err
			}
		}
	}
	if named_C {
		fmt.Print(conf)
		return nil
	}

	for _, s := range strings.Split(named_T, ",") {
		eq := strings.Index(s, "=")
		if eq < 0 {
			opt[s] = "true"
		} else {
			opt[s[:eq]] = s[eq+1:]
		}
	}

	for _, s := range strings.Split(opt["endpoints"], ",") {
		http.HandleFunc(s, namedHttpHandler)
	}

	ctx, cancel := context.WithCancel(ctx)

	namedWG.Go(func() { xdnsdb.Server(ctx, verbose) })

	if len(named_Z) > 0 {
		for _, fn := range strings.Split(named_Z, ",") {
			err = xdnsdb.Include(ctx, named_z, fn)
			if err != nil {
				cancel()
				namedWG.Wait()
				return err
			}
		}
	}

	host := ":"
	tcpNW := "tcp"
	udpNW := "udp"
	if named_4 {
		host = "0.0.0.0:"
		tcpNW = "tcp4"
		udpNW = "udp4"
	} else if named_6 {
		host = "[::]:"
		tcpNW = "tcp6"
		udpNW = "udp6"
	}

	for _, s := range strings.Split(named_p, ",") {
		if strings.HasPrefix(s, "http=") {
			laddr := host + strings.TrimPrefix(s, "http=")
			srv := &http.Server{Addr: laddr}
			namedWG.Go(func() { namedHttpShutdown(ctx, srv) })
			namedWG.Go(func() { namedHttpListenAndServe(srv) })
		} else if strings.HasPrefix(s, "https=") {
			laddr := host + strings.TrimPrefix(s, "https=")
			srv := &http.Server{Addr: laddr}
			namedWG.Go(func() { namedHttpShutdown(ctx, srv) })
			namedWG.Go(func() {
				namedHttpListenAndServe(srv,
					opt["cert"], opt["key"])
			})
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
			xln, ok := ln.(namedListener)
			if !ok {
				return errors.New("can't set deadline")
			}
			namedWG.Go(func() { namedTcpAccept(ctx, xln) })
		} else {
			laddr := ":" + strings.TrimPrefix(s, "dns=")
			if opt["notcp"] != "true" {
				ln, err := net.Listen(tcpNW, laddr)
				if err != nil {
					return err
				}
				defer ln.Close()
				xln, ok := ln.(namedListener)
				if !ok {
					return errors.New("can't set deadline")
				}
				namedWG.Go(func() { namedTcpAccept(ctx, xln) })
			}
			udpConn, err := net.ListenPacket(udpNW, laddr)
			if err != nil {
				return err
			}
			defer udpConn.Close()
			mch := make(chan *xdnsmessage.Message, 4)
			namedWG.Go(func() {
				namedUdpReceive(ctx, udpConn, mch)
			})
			namedWG.Go(func() {
				namedUdpService(ctx, udpConn, mch)
			})
		}
	}

	namedWG.Wait()
	return err
}

var (
	named_4 = false
	named_6 = false
	named_C = false
	named_T = ""
	named_V = false
	named_Z = ""
	named_c = NamedDefaultConf
	named_p = "53"
	named_q = false
	named_v = false
	named_z = "."
)

func defineNamedFlags() {
	xflag.Define(&named_4, "4", `Only service IPv4 host addresses.`)
	xflag.Define(&named_6, "6", `Only service IPv6 host addresses.`)
	xflag.Define(&named_C, "C", `Print configuration and exit.`)
	xflag.Define(&named_T, "T", `
Commas separated “<key>[=<value>]” options.  e.g.
    -T notcp,key=/etc/named.key,cert=/etc/named.crt`[1:])
	xflag.Define(&named_V, "V", `Print version and exit.`)
	xflag.Define(&named_Z, "Z", `
Comma separated zone files instead of or in addition to configuation.`[1:])
	xflag.Define(&named_c, "c", `
Absolute path name of configuration file.`[1:])
	xflag.Define(&named_p, "p", `
Comma separated ports on which the server will listen for queries.
If value is of the form “<portnum> or “dns=<portnum>”, the server will
listen for DNS queries on the numbered port. If value is of the form
“tls=<portnum>”, the server will listen for TLS queries on portnum;
the default is 853.  If value is of the form “https=<portnum>”,
the server will listen for HTTPS queries on portnum; the default is 443.
If value is of the form “http=<portnum>”, the server will listen for
HTTP queries on portnum; the default is 80.`[1:])
	xflag.Define(&named_q, "q", `Quiet logging.`)
	xflag.Define(&named_v, "v", `Verbose logging.`)
	xflag.Define(&named_z, "z", `Default zone.`)
}

func namedAnswer(req *xdnsmessage.Message) *xdnsmessage.Message {
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
func namedHttpHandler(rsp http.ResponseWriter, req *http.Request) {
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
	rspm := namedAnswer(reqm)
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

func namedHttpListenAndServe(srv *http.Server, fns ...string) {
	var err error
	defer verbose.Println("stopped http", srv.Addr, "service:", err)
	verbose.Println("start http", srv.Addr, "service")
	if len(fns) == 2 {
		err = srv.ListenAndServeTLS(fns[0], fns[1])
	} else {
		err = srv.ListenAndServe()
	}
}

func namedHttpShutdown(ctx context.Context, srv *http.Server) {
	const timeout = 3 * time.Second
	defer verbose.Print("http", srv.Addr, "done")
	<-ctx.Done()
	cctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	verbose.Print("shutdown http", srv.Addr, "...")
	srv.Shutdown(cctx)
}

func namedTcpAccept(ctx context.Context, xln namedListener) {
	var err error
	la := xln.Addr()
	defer verbose.Println("stopped tcp", la, "accept:", err)
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
		namedWG.Go(func() { namedTcpService(ctx, conn) })
	}
}

func namedTcpService(ctx context.Context, conn net.Conn) {
	ra := conn.RemoteAddr()
	defer verbose.Println("stopped tcp", ra, "service")
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
		ans := namedAnswer(req)
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

func namedUdpReceive(
	ctx context.Context,
	conn net.PacketConn,
	ch chan<- *xdnsmessage.Message,
) {
	defer verbose.Println("stopped udp receiver")
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

func namedUdpService(
	ctx context.Context,
	conn net.PacketConn,
	ch <-chan *xdnsmessage.Message,
) {
	defer verbose.Println("stopped udp message service")
	verbose.Println("start udp message service")
	data := make([]byte, 4<<10, 4<<10)
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-ch:
			ans := namedAnswer(req)
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
