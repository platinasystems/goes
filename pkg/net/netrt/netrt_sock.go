// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netrt

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"sync/atomic"
	"syscall"
	"unicode"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/net/netioctl"
	"github.com/platinasystems/goes/v2/pkg/net/sockaddr"
	"github.com/platinasystems/goes/v2/pkg/os/page"
	"github.com/platinasystems/goes/v2/pkg/syscall/af"
)

type appendAddrFunc func(context.Context, *flag.FlagSet, []byte) ([]byte, error)

var seq atomic.Int32

func Add(ctx context.Context, fs *flag.FlagSet) error {
	_, err := rtreq(ctx, fs, syscall.RTM_ADD)
	return err
}

func Change(ctx context.Context, fs *flag.FlagSet) error {
	_, err := rtreq(ctx, fs, syscall.RTM_CHANGE)
	return err
}

func Delete(ctx context.Context, fs *flag.FlagSet) error {
	_, err := rtreq(ctx, fs, syscall.RTM_DELETE)
	return err
}

func Flush(ctx context.Context, fs *flag.FlagSet) error {
	return FIXME
}

func Get(ctx context.Context, fs *flag.FlagSet) (
	NetRt, error,
) {
	nrt, err := rtreq(ctx, fs, syscall.RTM_GET)
	return nrt, err
}

func Monitor(ctx context.Context) (Streamer, error) {
	return nil, FIXME
}

func rtreq(ctx context.Context, fs *flag.FlagSet, cmd uint8) (NetRt, error) {
	var err error
	msg := page.New()
	rtm := Pointer[syscall.RtMsghdr](msg)
	msg = msg[:Sizeof(rtm)]
	rtm.Type = cmd
	rtm.Version = syscall.RTM_VERSION
	rtm.Seq = seq.Add(1)
	setFlags(rtm, fs, cmd)
	for _, f := range []appendAddrFunc{
		appendDstGatewayNetmask,
		appendGenmask,
		appendIfp,
		appendIfa,
		appendAuthor,
		appendBrd,
	} {
		if msg, err = f(ctx, fs, msg); err != nil {
			return nil, err
		}
	}
	rtm = PointerRtMsghdr(msg)
	rtm.Msglen = uint16(len(msg))
	sock, err := af.OpenRoute()
	if err != nil {
		return nil, egress.Marked(err)
	}
	defer af.Close(sock)
	/* FIXME darwin doesn't have SO_SETFIB
	if fib := flag.Search[int]("F", fs); fib >= 0 {
		if err = os.NewSyscallError("SO_SETFIB", syscall.
			SetsockoptInt(int(sock), syscall.SOL_SOCKET,
				syscall.SO_SETFIB, fib)); err != nil {
			return nil, err
		}
	}
	*/
	if _, err = af.Write(sock, msg); err != nil {
		switch {
		case errors.Is(err, syscall.ESRCH):
			return nil, ErrSRCH
		case errors.Is(err, syscall.EBUSY):
			return nil, ErrBUSY
		case errors.Is(err, syscall.ENOBUFS):
			return nil, ErrNOBUFS
		case errors.Is(err, syscall.EADDRINUSE):
			return nil, ErrADDRINUSE
		case errors.Is(err, syscall.EEXIST):
			return nil, ErrEXIST
		default:
			return nil, egress.Markf("%w\n%#v", err, rtm)
		}
	} else if cmd != syscall.RTM_GET {
		return nil, nil
	}
	msg = msg[:cap(msg)]
	n, err := af.Read(sock, msg)
	if err != nil {
		return nil, egress.Marked(err)
	}
	return newNetRt(msg[:n]), nil
}

func rtrmx[T int32 | uint32](rmx *T, inits *uint32, v uint, rtv uint32) {
	if v > 0 {
		*(rmx) = T(v)
		*(inits) |= rtv
	}
}

func appendDstGatewayNetmask(
	ctx context.Context, fs *flag.FlagSet, msg []byte,
) ([]byte, error) {
	var (
		err   error
		ok    bool
		addrs []net.IPAddr
		addr  netip.Addr
		mask  net.IPMask
	)
	bits := -1
	s := flag.Search[string]("dst", fs)
	if len(s) == 0 {
		if fs.NArg() == 0 {
			return msg, ErrNoDst
		}
		s = fs.Arg(0)
	}
	if slash := strings.Index(s, "/"); slash > 0 {
		if _, err = fmt.Sscan(s[slash+1:], &bits); err != nil {
			return msg, egress.Markf("%q %w", s[slash+1:], err)
		}
		s = s[:slash]
	}
	if s == "default" {
		if flag.Search[bool]("6", fs) {
			addr = netip.IPv6Unspecified()
		} else {
			addr = netip.IPv4Unspecified()
		}
	} else if s == "::" {
		addr = netip.IPv6Unspecified()
	} else if isnumeric(s) {
		if addr, err = netip.ParseAddr(s); err != nil {
			return msg, egress.Markf("%q %w", s, err)
		}
	} else if addrs, err = net.DefaultResolver.
		LookupIPAddr(ctx, s); err != nil {
		return msg, egress.Markf("%q %w", s, err)
	} else if addr, ok = netip.AddrFromSlice(addrs[0].IP); !ok {
		return msg, egress.Markf("%v invalid", addrs[0].IP)
	}
	if bits < 0 {
	} else if addr.Is6() {
		mask = net.CIDRMask(bits, 128)
	} else {
		mask = net.CIDRMask(bits, 32)
	}
	msg = sockaddr.Append(msg, addr)
	PointerRtMsghdr(msg).Addrs |= 1 << syscall.RTAX_DST

	if s = flag.Search[string]("gateway", fs); len(s) == 0 {
		if fs.NArg() > 1 {
			s = fs.Arg(1)
		}
	}
	if len(s) > 0 {
		if flag.Search[bool]("interface", fs) {
			nif := netif.Named(s)
			if nif == nil {
				return msg, egress.Markf("%q not found", s)
			}
			msg = sockaddr.AppendDl(msg,
				uint16(nif.Index),
				uint8(nif.Type.(netioctl.IFT)),
				nif.Name,
				nif.HardwareAddr,
				[]byte{})
			PointerRtMsghdr(msg).Addrs |= 1 << syscall.RTAX_GATEWAY
		} else if isnumeric(s) {
			if addr, err = netip.ParseAddr(s); err != nil {
				return msg, egress.Markf("%q %w", s, err)
			}
			msg = sockaddr.Append(msg, addr)
			PointerRtMsghdr(msg).Addrs |= 1 << syscall.RTAX_GATEWAY
		} else if addrs, err := net.DefaultResolver.
			LookupIPAddr(ctx, s); err != nil {
			return msg, egress.Markf("%q %w", s, err)
		} else if addr, ok = netip.AddrFromSlice(addrs[0].IP); !ok {
			return msg, egress.Markf("%v invalid", addrs[0].IP)
		} else {
			msg = sockaddr.Append(msg, addr)
			PointerRtMsghdr(msg).Addrs |= 1 << syscall.RTAX_GATEWAY
		}
	}

	if len(mask) == 0 {
		if addr.Is6() {
			bits = flag.Search[int]("prefixlen", fs)
			if bits >= 0 {
				mask = net.CIDRMask(int(bits), 128)
			}
		} else {
			s = flag.Search[string]("mask", fs)
			if len(s) == 0 && fs.NArg() > 2 {
				s = fs.Arg(2)
			}
			if len(s) > 0 {
				if addr, err = netip.ParseAddr(s); err != nil {
					return msg, egress.Markf("%q %w", s, err)
				}
				mask = net.IPMask(addr.AsSlice())
			}
		}
	}
	if len(mask) > 0 {
		if addr, ok = netip.AddrFromSlice(mask); ok {
			msg = sockaddr.Append(msg, addr)
			PointerRtMsghdr(msg).Addrs |= 1 << syscall.RTAX_NETMASK
		}
	}
	return msg, nil
}

func appendGenmask(
	ctx context.Context, fs *flag.FlagSet, msg []byte,
) ([]byte, error) {
	s := flag.Search[string]("genmask", fs)
	if len(s) == 0 {
		return msg, nil
	}
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return msg, egress.Markf("%q %w", s, err)
	}
	msg = sockaddr.Append(msg, addr)
	PointerRtMsghdr(msg).Addrs |= 1 << syscall.RTAX_GENMASK
	return msg, nil
}

func appendIfp(
	ctx context.Context, fs *flag.FlagSet, msg []byte,
) ([]byte, error) {
	s := flag.Search[string]("ifp", fs)
	if len(s) == 0 {
		return msg, nil
	}
	nif := netif.Named(s)
	if nif == nil {
		return msg, egress.Markf("%q not found", s)
	}
	msg = sockaddr.AppendDl(msg,
		uint16(nif.Index),
		uint8(nif.Type.(netioctl.IFT)),
		nif.Name,
		nif.HardwareAddr,
		[]byte{})
	PointerRtMsghdr(msg).Addrs |= 1 << syscall.RTAX_IFP
	return msg, nil
}

func appendIfa(
	ctx context.Context, fs *flag.FlagSet, msg []byte,
) ([]byte, error) {
	s := flag.Search[string]("ifa", fs)
	if len(s) == 0 {
		return msg, nil
	}
	return msg, nil
}

func appendAuthor(
	ctx context.Context, fs *flag.FlagSet, msg []byte,
) ([]byte, error) {
	s := flag.Search[string]("author", fs)
	if len(s) == 0 {
		return msg, nil
	}
	// redirect ?
	return msg, nil
}

func appendBrd(
	ctx context.Context, fs *flag.FlagSet, msg []byte,
) ([]byte, error) {
	// broadcast || point-to-point peer
	return msg, nil
}

func isnumeric(s string) bool {
	r := []rune(s)[0]
	return unicode.IsNumber(r) || r == ':'
}
