// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
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
	"text/tabwriter"
	"time"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/netif"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"golang.org/x/sys/unix"
)

type RtMsghdr = unix.RtMsghdr

func Extract[T RtMsghdr](data []byte) (t *T, body, rem []byte) {
	l := xnet.SysctlMsgLen(data)
	t = Pointer[T](data)
	body = data[Sizeof(t):]
	rem = data[xnet.SysctlAlign(l):]
	return
}

var ExtractRtMsghdr = Extract[RtMsghdr]

func Pointer[T ~uint16 | RtMsghdr](data []byte) *T {
	return (*T)(unsafe.Pointer(&data[0]))
}

var PointerRtMsghdr = Pointer[RtMsghdr]

func Sizeof[T RtMsghdr](p *T) int {
	return int(unsafe.Sizeof(*p))
}

var SizeofRtMsghdr = Sizeof[RtMsghdr]

type link struct {
	line uint16
	ha   net.HardwareAddr
}

type netrt struct {
	index uint16
	_     uint16
	flags int32
	rmx   unix.RtMetrics
	addrs [unix.RTAX_MAX]any
}

func newNetRt(msg []byte) NetRt {
	rtm := PointerRtMsghdr(msg)
	body := msg[xnet.SysctlAlign(Sizeof(rtm)):]
	nrt := &netrt{
		index: rtm.Index,
		flags: rtm.Flags,
		rmx:   rtm.Rmx,
	}
	const min = xnet.SAMin
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
	tw := tabwriter.NewWriter(w, 9, 0, 1, ' ', tabwriter.AlignRight)
	defer tw.Flush()
	fmt.Fprint(tw, "recvpipe", "\t")
	fmt.Fprint(tw, "sendpipe", "\t")
	fmt.Fprint(tw, "ssthresh", "\t")
	fmt.Fprint(tw, "rtt", "\t")
	fmt.Fprint(tw, "rttvar", "\t")
	fmt.Fprint(tw, "hopcount", "\t")
	fmt.Fprint(tw, "mtu", "\t")
	fmt.Fprint(tw, "expire", "\t")
	fmt.Fprintln(tw)
	fmt.Fprint(tw, nrt.rmx.Recvpipe, "\t")
	fmt.Fprint(tw, nrt.rmx.Sendpipe, "\t")
	fmt.Fprint(tw, nrt.rmx.Ssthresh, "\t")
	fmt.Fprint(tw, nrt.rmx.Rtt, "\t")
	fmt.Fprint(tw, nrt.rmx.Rttvar, "\t")
	fmt.Fprint(tw, nrt.rmx.Hopcount, "\t")
	fmt.Fprint(tw, nrt.rmx.Mtu, "\t")
	if nrt.rmx.Expire == 0 {
		fmt.Fprint(tw, 0, "\t")
	} else {
		ut := time.Unix(int64(nrt.rmx.Expire), 0)
		expire := ut.Sub(time.Now()).Round(time.Second).Seconds()
		fmt.Fprint(tw, expire, "\t")
	}
	fmt.Fprintln(tw)
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
