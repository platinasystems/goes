// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// DNS Over HTTPS
package xdnsdoh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/cert"
	"github.com/platinasystems/goes/v2/pkg/kvc"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
)

type Config struct {
	URL, Search string
}

var GetConfig = sync.OnceValues(func() (cfg Config, err error) {
	fn := xmain.ConfigFile("doh")
	err = kvc.RangeFile(fn, func(s string) []string {
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
	}, func(lno int, key string, values []string) error {
		if len(values) < 0 {
			return xerrors.Label(xerrors.Incomplete(lno), fn)
		}
		switch key {
		case "search":
			cfg.Search = values[0]
		case "url":
			cfg.URL = values[0]
		default:
			return xerrors.Label(xerrors.Invalid(lno), fn)
		}
		return nil
	})
	return
})

// Return successful [netip.ParseAddrPort];
// otherwise, [LookupNetIP] and [net.DefaultResolver.LookupPort]
// the respective host and port segments returned from [net.SplitHostPort](s);
// then recombine as [netip.AddrPort] list.
// The returned ports are zero if (s) didn't have a “:<port>” suffix.
func LookupAddrPort(ctx context.Context, nw, s string) (
	aps []netip.AddrPort, err error,
) {
	is4nw := strings.HasSuffix(nw, "4")
	is6nw := strings.HasSuffix(nw, "6")
	ap, err := netip.ParseAddrPort(s)
	if err == nil {
		aps = append(aps, ap)
		return
	}

	var hs, ps string
	if strings.Count(s, ":") != 1 && strings.Count(s, "]:") != 1 {
		hs = s
	} else if hs, ps, err = net.SplitHostPort(s); err != nil {
		return
	}

	var port int
	if len(ps) > 0 {
		port, err = net.DefaultResolver.LookupPort(ctx, nw, ps)
		if err != nil {
			return
		}
	}
	addrs, err := LookupNetIP(ctx, hs)
	if err != nil {
		return
	}
	for _, addr := range addrs {
		if (is4nw && addr.Is6()) || (is6nw && addr.Is4()) {
			continue
		}
		aps = append(aps, netip.AddrPortFrom(addr, uint16(port)))
	}
	if len(aps) == 0 {
		err = xerrors.NotFound(s)
	}
	return
}

// Like [net.Resolver.LookupAddr] but with [netip.Addr]
// instead of string parameter.
func LookupName(ctx context.Context, addr netip.Addr) ([]string, error) {
	var names []string
	doh, err := New()
	if err == nil {
		names, err = doh.LookupName(ctx, addr)
	} else if errors.Is(err, fs.ErrNotExist) {
		names, err = net.DefaultResolver.LookupAddr(ctx, addr.String())
	}
	if err == nil && len(names) == 0 {
		err = xerrors.NotFound(addr)
	}
	return names, err
}

// Return successful [netip.ParseAddr];
// otherwise, returned lookup of the configured URL,
// or if that's unconfigured,
// the result of [net.DefaultResolver.LookupNetIP].
// A nil-error result will always return at least one [netip.Addr].
func LookupNetIP(ctx context.Context, s string) ([]netip.Addr, error) {
	var (
		addr  netip.Addr
		addrs []netip.Addr
		doh   *DOH
		err   error
	)
	if addr, err = netip.ParseAddr(s); err == nil {
		if addr.Is4In6() {
			addr = addr.Unmap()
		}
		addrs = []netip.Addr{addr}
	} else if doh, err = New(); err == nil {
		addrs, err = doh.LookupNetIP(ctx, s)
	} else if errors.Is(err, fs.ErrNotExist) {
		addrs, err = net.DefaultResolver.LookupNetIP(ctx, "ip", s)
	}
	if err == nil && len(addrs) == 0 {
		err = xerrors.NotFound(s)
	} else {
		for i, addr := range addrs {
			if addr.Is4In6() {
				addrs[i] = addr.Unmap()
			}
		}
	}
	return addrs, err
}

type DOH struct {
	*http.Client
	buf []byte
	url,
	search string
}

// Create [http.Client] of [URL] to ask DOH queries.
var New = sync.OnceValues(func() (*DOH, error) {
	cfg, err := GetConfig()
	if err != nil {
		return nil, err
	}
	httpc, err := cert.NewHTTPClient()
	if err != nil {
		return nil, err
	}
	return NewClient(httpc, cfg.URL, cfg.Search), nil
})

// Use [http.Client] to ask DOH queries from “url”.
func NewClient(cl *http.Client, url, search string) *DOH {
	return &DOH{
		Client: cl,
		buf:    xdnsmessage.MakeBuffer(),
		url:    url,
		search: search,
	}
}

func (doh *DOH) Ask(ctx context.Context, b []byte) ([]byte, error) {
	r := bytes.NewReader(b)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, doh.url, r)
	if err != nil {
		return b[:0], err
	}
	req.Header.Set("Content-Type", "application/dns-message")
	rsp, err := doh.Do(req)
	if err != nil {
		return b[:0], err
	}
	if rsp == nil {
		return b[:0], errors.New("nil")
	}
	defer rsp.Body.Close()
	if rsp.StatusCode == http.StatusOK {
		buf := new(bytes.Buffer)
		io.Copy(buf, rsp.Body)
		return buf.Bytes(), nil
	}
	sb := new(strings.Builder)
	io.Copy(sb, rsp.Body)
	if sb.Len() == 0 {
		return b[:0], errors.New(rsp.Status)
	}
	return b[:0], fmt.Errorf("%s, %s", rsp.Status, sb.String())
}

func (doh *DOH) RecursiveLookup(
	ctx context.Context,
	us xdnsmessage.UniqueString,
	c xdnsmessage.Class,
	t xdnsmessage.Type,
) (answers []xdnsmessage.WireResource, err error) {
	var rsp xdnsmessage.Message
	q := xdnsmessage.NewQuery(true, us, c, t)
	doh.buf, err = q.AppendTo(doh.buf[:0])
	if err == nil {
		if doh.buf, err = doh.Ask(ctx, doh.buf); err == nil {
			if err = rsp.UnmarshalBinary(doh.buf); err == nil {
				if rsp.ID == q.ID {
					answers = rsp.Answers
				} else {
					err = fmt.Errorf("rsp id %d != req %d",
						rsp.ID, q.ID)
				}
			}
		}
	}
	return
}

// Like [net.Resolver.LookupAddr] but with [netip.Addr]
// instead of string parameter.
func (doh *DOH) LookupName(ctx context.Context, addr netip.Addr) (
	names []string, err error,
) {
	const (
		class   = xdnsmessage.ClassINET
		typePTR = xdnsmessage.TypePTR
	)
	name := xdnsmessage.Reverse(addr)
	us := xdnsmessage.MakeUniqueString(name)
	ans, err := doh.RecursiveLookup(ctx, us, class, typePTR)
	if err != nil {
		return
	}
	for _, a := range ans {
		names = append(names, a.String())
	}
	return
}

// Like [net.Resolver.LookupNetIP] w/o the “network” parameter.
func (doh *DOH) LookupNetIP(ctx context.Context, name string) (
	addr []netip.Addr, err error,
) {
	const (
		class    = xdnsmessage.ClassINET
		typeA    = xdnsmessage.TypeA
		typeAAAA = xdnsmessage.TypeAAAA
	)
	if !strings.Contains(name, ".") {
		if len(doh.search) == 0 {
			err = xerrors.Incomplete(name)
			return
		}
		if !strings.HasPrefix(doh.search, ".") {
			name += "."
		}
		name += doh.search
	}
	if !strings.HasSuffix(name, ".") {
		name += "."
	}
	us := xdnsmessage.MakeUniqueString(name)
	ans, err := doh.RecursiveLookup(ctx, us, class, typeA)
	if err != nil {
		return
	}
	for _, a := range ans {
		if a.Type() == typeA {
			res := a.Resource().(xdnsmessage.TypeAResource)
			addr = append(addr, res.Addr)
		}
	}
	ans, err = doh.RecursiveLookup(ctx, us, class, typeAAAA)
	if err == nil {
		for _, a := range ans {
			if a.Type() == typeAAAA {
				x := a.Resource().(xdnsmessage.TypeAAAAResource)
				addr = append(addr, x.Addr)
			}
		}
	}
	return
}
