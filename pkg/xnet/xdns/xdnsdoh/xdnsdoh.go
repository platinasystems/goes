// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Dns-Over-Http(s)
package xdnsdoh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/cert"
	"github.com/platinasystems/goes/v2/pkg/kvc"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
)

var client *http.Client
var config struct{ url, search string }

// only make one request at a time
var mutex sync.Mutex

const DefaultConfigFile = "doh"

var (
	ConfigFile = DefaultConfigFile
	ConfigFlag = xflag.Label{"doh", `
File w/in current or config directory containing line separated
assignments of Dns-Over-Http(s) “url” and “search” keywords.
No DOH if this flag is empty or default file doesn't exist.`[1:], &ConfigFile}
)

// If [ConfigFile] exists or given optional “url” parameter,
// this returns an [xdns.Asker] compatible function
// to forward binary encoded DNS requests to DOH server;
// otherwise, returns nil so that caller may establish fallback.
func Asker(skipVerify bool, url ...any) (xdns.Asker, error) {
	var err error
	if len(url) > 0 {
		config.url = fmt.Sprint(url...)
	} else if err = loadConfig(); err != nil {
		return nil, err
	} else if len(config.url) == 0 {
		return nil, nil
	}
	if skipVerify {
		cert.Verify = false
	}
	client, err = cert.NewHTTPClient()
	if err != nil {
		return nil, err
	}
	return ask, nil
}

// If [ConfigFile] exists and has an assigned “search” keyword,
// append this to the unqualified (no dot) string to form a
// Fully-Qualified-Domain-Name;
// otherwise, return unchanged.
func FQDN(s string) string {
	if len(config.search) == 0 || strings.IndexRune(s, '.') >= 0 {
		return s
	}
	if !strings.HasPrefix(config.search, ".") {
		s += "."
	}
	s += config.search
	if !strings.HasSuffix(s, ".") {
		s += "."
	}
	return s
}

// If [ConfigFile] exists,
// this returns a custom [net.Resolver]
// that has a custom [net.Resolver.Dial]
// that returns a custom [net.Conn]
// whose [net.Conn.Write] makes a DOH query with
// the binary DNS formatted slice
// and [net.Conn.Read] returns the response.
// Otherwise, return [net.DefaultResolver].
func Resolver() (*net.Resolver, error) {
	err := loadConfig()
	if err != nil {
		return nil, err
	}
	if len(config.url) == 0 {
		return net.DefaultResolver, nil
	}
	client, err = cert.NewHTTPClient()
	if err != nil {
		return nil, err
	}
	return &net.Resolver{
		PreferGo: true,
		Dial:     dial,
	}, nil
}

func ask(ctx context.Context, b []byte) ([]byte, error) {
	mutex.Lock()
	defer mutex.Unlock()
	r := bytes.NewReader(b)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		config.url, r)
	if err != nil {
		return b[:0], err
	}
	req.Header.Set("Content-Type", "application/dns-message")
	rsp, err := client.Do(req)
	if err != nil {
		return b[:0], err
	}
	if rsp == nil {
		return b[:0], errors.New("nil response")
	}
	defer rsp.Body.Close()
	n, err := rsp.Body.Read(b[:cap(b)])
	if err != nil {
		n = 0
	}
	b = b[:n]
	if rsp.StatusCode == http.StatusOK {
	} else if n == 0 {
		err = errors.New(rsp.Status)
	} else {
		err = fmt.Errorf("%s, %s", rsp.Status, b)
		b = b[:0]
	}
	return b, err
}

func dial(ctx context.Context, nw, s string) (net.Conn, error) {
	var lap netip.AddrPort
	rap, err := netip.ParseAddrPort(s)
	if err != nil {
		return nil, err
	}
	if a := rap.Addr(); a.Is4In6() {
		rap = netip.AddrPortFrom(a.Unmap(), rap.Port())
	}
	if rap.Addr().Is6() {
		lap = netip.AddrPortFrom(netip.IPv4Unspecified(), 0)
	} else {
		lap = netip.AddrPortFrom(netip.IPv4Unspecified(), 0)
	}
	if strings.HasPrefix(nw, "tcp") {
		la := net.TCPAddrFromAddrPort(lap)
		ra := net.TCPAddrFromAddrPort(rap)
		return newTCP(ctx, la, ra), nil
	}
	la := net.UDPAddrFromAddrPort(lap)
	ra := net.UDPAddrFromAddrPort(rap)
	return newUDP(ctx, la, ra), nil
}

// If [ConfigFile] is an empty string or is [DefaultConfigFile]
// and that doesn't exist, this returns nil and [config.url] will be empty.
func loadConfig() error {
	if len(ConfigFile) == 0 {
		return nil
	}
	fn := ConfigFile
	_, err := os.Stat(fn)
	if err != nil {
		if strings.IndexRune(fn, os.PathSeparator) >= 0 {
			return err
		}
		fn = xmain.ConfigFile(ConfigFile)
		if _, err = os.Stat(fn); err != nil {
			if ConfigFile == DefaultConfigFile {
				return nil
			}
			return err
		}
	}
	err = kvc.RangeFile(fn,
		func(s string) []string {
			s = strings.TrimSpace(s)
			if len(s) == 0 || []rune(s)[0] == '#' {
				return nil
			}
			args := strings.Fields(s)
			for i, arg := range args {
				if len(arg) == 0 || []rune(arg)[0] == '#' {
					args = args[:i]
					break
				}
			}
			return args
		},
		func(lno int, key string, values []string) error {
			if len(values) > 0 {
				switch key {
				case "search":
					config.search = values[0]
				case "url":
					config.url = values[0]
				}
			}
			return nil
		})
	if err != nil {
		return err
	}
	if len(config.url) == 0 {
		return xerrors.Label(xerrors.Incomplete("url"), fn)
	}
	return nil
}

type tcpConn struct {
	dohConn
	// accumulate writes until prefaced length
	buf []byte
	n   int
}

func newTCP(ctx context.Context, la, ra net.Addr) net.Conn {
	c := new(tcpConn)
	c.init(ctx, la, ra)
	c.buf = xdnsmessage.MakeBuffer()
	return c
}

func (c *tcpConn) Read(b []byte) (int, error) {
	n, err := c.read(b[2:])
	if err == nil {
		xnet.ByteOrder.PutUint16(b, uint16(n))
		n += 2
	} else {
		n = 0
	}
	return n, err
}

func (c *tcpConn) Write(b []byte) (int, error) {
	if c.n == 0 {
		c.n = int(xnet.ByteOrder.Uint16(b))
		c.buf = append(c.buf[:0], b[2:]...)
		if len(c.buf) < c.n {
			return len(b), nil
		}
	} else {
		c.buf = append(c.buf, b...)
		if len(c.buf) < c.n {
			return len(b), nil
		}
	}
	_, err := c.write(c.buf[:c.n])
	c.n = 0
	return len(b), err
}

type udpConn struct {
	dohConn
}

func newUDP(ctx context.Context, la, ra net.Addr) interface {
	net.Conn
	net.PacketConn
} {
	c := new(udpConn)
	c.init(ctx, la, ra)
	return c
}

func (c *udpConn) Read(b []byte) (int, error) {
	return c.read(b)
}

func (c *udpConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, err := c.read(b)
	return n, c.ra, err
}

func (c *udpConn) Write(b []byte) (int, error) {
	return c.write(b)
}

func (c *udpConn) WriteTo(b []byte, ra net.Addr) (int, error) {
	return c.write(b)
}

type dohConn struct {
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
	ans    chan any // buf or error
	la, ra net.Addr
	mt     struct {
		// Mutexed Timeout
		sync.RWMutex
		time.Duration
	}
}

func (c *dohConn) init(ctx context.Context, la, ra net.Addr) {
	c.ctx, c.cancel = context.WithCancel(ctx)
	c.ans = make(chan any, 1)
	c.ra = ra
	c.ra = ra
}

func (c *dohConn) Close() error {
	c.cancel()
	return nil
}

func (c *dohConn) LocalAddr() net.Addr  { return c.la }
func (c *dohConn) RemoteAddr() net.Addr { return c.ra }

func (c *dohConn) SetDeadline(t time.Time) error {
	c.SetReadDeadline(t)
	c.SetWriteDeadline(t)
	return nil
}

func (c *dohConn) SetReadDeadline(t time.Time) error {
	// do not timeout
	return nil
}

func (c *dohConn) SetWriteDeadline(t time.Time) (err error) {
	c.mt.Lock()
	defer c.mt.Unlock()
	if t.IsZero() {
		c.mt.Duration = 0
	} else if now := time.Now(); t.After(now) {
		c.mt.Duration = t.Sub(now)
	} else {
		err = os.ErrDeadlineExceeded
	}
	return
}

func (c *dohConn) ask(buf []byte) {
	buf, err := ask(c.ctx, buf)
	if err != nil {
		xdnsmessage.FreeBuffer(buf)
		c.ans <- err
	} else {
		c.ans <- buf
	}
}

func (c *dohConn) dump(prefix string, buf []byte) error {
	m := xdnsmessage.NewMessage()
	defer m.Free()
	err := m.UnmarshalBinary(buf)
	if err == nil {
		fmt.Println(prefix, m)
	}
	return err
}

func (c *dohConn) read(b []byte) (int, error) {
	select {
	case <-c.ctx.Done():
		return 0, c.ctx.Err()
	case ans, ok := <-c.ans:
		if !ok || ans == nil {
			return 0, io.EOF
		}
		err, iserr := ans.(error)
		if iserr {
			return 0, err
		}
		buf, isbuf := ans.([]byte)
		if !isbuf {
			return 0, fmt.Errorf("unexpected %T", ans)
		}
		n := copy(b, buf)
		xdnsmessage.FreeBuffer(buf)
		return n, nil
	}
}

func (c *dohConn) timeout() time.Duration {
	c.mt.RLock()
	defer c.mt.RUnlock()
	return c.mt.Duration
}

func (c *dohConn) write(b []byte) (int, error) {
	err := c.ctx.Err()
	if err != nil {
		c.once.Do(func() { close(c.ans) })
		return 0, err
	}
	client.Timeout = c.timeout()
	buf := xdnsmessage.MakeBuffer()
	n := copy(buf, b)
	buf = buf[:n]
	go c.ask(buf)
	return n, nil
}
