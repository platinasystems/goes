// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netrt

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/netip"
	"time"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/integer"
	"github.com/platinasystems/goes/v2/pkg/netif"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xos/sysctl"
	"golang.org/x/sys/unix"
)

const RTM_VERSION = unix.RTM_VERSION

type RtMsghdr = unix.RtMsghdr
type RtMsghdr2 = unix.RtMsghdr2
type RtMetrics = unix.RtMetrics

func Pointer[T ~uint16 | RtMsghdr | RtMsghdr2](data []byte) *T {
	return (*T)(unsafe.Pointer(&data[0]))
}

var PointerRtMsghdr2 = Pointer[RtMsghdr2]

func Sizeof[T RtMsghdr | RtMsghdr2 | RtMetrics](p *T) int {
	return int(unsafe.Sizeof(*p))
}

func Extract[T RtMsghdr | RtMsghdr2 | RtMetrics](data []byte) (
	t *T, body, rem []byte,
) {
	if n, i := len(data), Sizeof(t); n >= i {
		t = (*T)(unsafe.Pointer(&data[0]))
		body = data[i:]
		i = sysctl.Align(sysctl.Len(data))
		if n < i {
			i = n
		}
		rem = data[i:]
	}
	return
}

type link struct {
	line uint16
	ha   net.HardwareAddr
}

type netrt struct {
	expire,
	index,
	state int
	flags,
	probes uint
	addrs [unix.RTAX_MAX]any
}

func newNetRt(rtm *RtMsghdr2, body []byte) *netrt {
	const min = xnet.SAMin
	nrt := new(netrt)
	integer.Assign(&nrt.index, rtm.Index)
	integer.Assign(&nrt.flags, rtm.Flags)
	integer.Assign(&nrt.expire, rtm.Rmx.Expire)
	integer.Assign(&nrt.state, rtm.Rmx.State)
	integer.Assign(&nrt.probes, rtm.Rmx.Pksent)
	for i := 0; i < len(nrt.addrs[:]) && len(body) > min; i++ {
		if (int(rtm.Addrs) & (1 << i)) == 0 {
			continue
		}
		sal := xnet.SALen(body)
		if sal == 0 {
			body = body[xnet.Sizeof32:]
			continue
		}
		if sal > len(body) {
			log.Printf("addr[%d] too long, %d", i, sal)
			break
		}
		switch family := xnet.SAFamily(body); family {
		case xnet.AF_INET:
			nrt.addrs[i] = xnet.SAIP4(body)
		case xnet.AF_INET6:
			nrt.addrs[i] = xnet.SAIP6(body)
		case xnet.AF_LINK:
			dl, dlbody, _ := xnet.SAExtractDataLink(body)
			_, address, _ := dl.NAS(dlbody)
			nrt.addrs[i] = &link{dl.Index, address}
		case 0xff:
			if v := nrt.addrs[unix.RTAX_DST]; v != nil {
				if addr, ok := v.(netip.Addr); ok {
					if addr.Is4() {
						nrt.addrs[i] = xnet.
							SAIP4(body)
					} else if addr.Is6() {
						nrt.addrs[i] = xnet.
							SAIP6(body)
					}
				}
			}
		default:
			log.Printf("addr[%d], family[%#x]", i, family)
		}
		body = body[xnet.Align32(sal):]
	}
	return nrt
}

func (nrt *netrt) Index() int  { return int(nrt.index) }
func (nrt *netrt) Flags() uint { return uint(nrt.flags) }

func vip(v any) netip.Addr {
	if ip, ok := v.(netip.Addr); ok {
		return ip
	}
	return netip.Addr{}
}

func (nrt *netrt) Dst() netip.Addr {
	return vip(nrt.addrs[unix.RTAX_DST])
}

func (nrt *netrt) GW() netip.Addr {
	return vip(nrt.addrs[unix.RTAX_GATEWAY])
}

func (nrt *netrt) Netmask() netip.Addr {
	v := FirstNonNil(nrt.addrs[unix.RTAX_NETMASK],
		nrt.addrs[unix.RTAX_GENMASK])
	return vip(v)
}

func (nrt *netrt) IFA() netip.Addr {
	return vip(nrt.addrs[unix.RTAX_IFA])
}

func (nrt *netrt) Line() int {
	if lnk, ok := nrt.addrs[unix.RTAX_GATEWAY].(*link); ok {
		return int(lnk.line)
	}
	return -1
}

func (nrt *netrt) HA() net.HardwareAddr {
	if lnk, ok := nrt.addrs[unix.RTAX_GATEWAY].(*link); ok {
		return lnk.ha
	}
	return net.HardwareAddr{}
}

func (nrt *netrt) Bits() int {
	if nm := vip(nrt.addrs[unix.RTAX_NETMASK]); nm.IsValid() {
		ones, _ := net.IPMask(nm.AsSlice()).Size()
		return ones
	}
	return 0
}

func (nrt *netrt) Expire() int  { return nrt.expire }
func (nrt *netrt) Probes() uint { return nrt.probes }
func (nrt *netrt) State() int   { return nrt.state }

func (nrt *netrt) Format(w fmt.State, verb rune) {
	for i, s := range []string{
		"destination", //	RTAX_DST	= 0x0
		"gateway",     //	RTAX_GATEWAY	= 0x1
		"netmask",     //	RTAX_NETMASK	= 0x2
		"genmask",     //	RTAX_GENMASK	= 0x3
		"ifp",         //	RTAX_IFP	= 0x4
		"ifa",         //	RTAX_IFA	= 0x5
		"author",      //	RTAX_AUTHOR	= 0x6
		"broadcast",   //	RTAX_BRD	= 0x7
	} {
		if v := nrt.addrs[i]; v != nil {
			fmt.Fprintf(w, "%11s: %s\n", s, rtname(v))
		}
	}
	name, err := netif.Name(context.TODO(), nrt.Index())
	fmt.Fprintf(w, "%11s: ", "interface")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	fmt.Fprintln(w, name)
	fmt.Fprintf(w, "%11s: <%s>\n", "flags", xnet.IFFNames(nrt.flags))
	fmt.Fprintf(w, "%11s: ", "expire")
	if nrt.expire == 0 {
		fmt.Fprint(w, 0, "\n")
	} else {
		ut := time.Unix(int64(nrt.expire), 0)
		expire := ut.Sub(time.Now()).Round(time.Second).Seconds()
		fmt.Fprint(w, expire, "\n")
	}
}

func rtname(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case netip.Addr:
		if t.IsLoopback() {
			return "loopback"
		}
		if t.IsUnspecified() {
			return "default"
		}
		return t.String()
	case *link:
		name, err := netif.Name(context.TODO(), int(t.line))
		if err != nil {
			return err.Error()
		} else if len(t.ha) == 0 {
			return name
		} else {
			return fmt.Sprintf("%s[%v]", name, t.ha)
		}
	default:
		return fmt.Sprintf("%#v", v)
	}
}
