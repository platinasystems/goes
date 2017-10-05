// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"
	"net"
	"syscall"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

// ip link COMMAND type geneve { id | vni } ID remote ADDR
//	[ ttl TTL ]
//	[ tos TOS ]
//	[ flowlabel LABEL ]
//	[ dstport PORT ]
//	[ [no-]external ]
//	[ [no-]udpcsum ]
//	[ [no-]udp6zerocsumtx ]
//	[ [no-]udp6zerocsumrx ]
func (m *mod) parseTypeGeneve() error {
	var s string
	var u8 uint8
	var u16 uint16
	var u32 uint32
	var err error
	m.args = m.opt.Parms.More(m.args, []string{"id", "vni"})
	s = m.opt.Parms.ByName["id"]
	if len(s) == 0 {
		return fmt.Errorf("missing id")
	}
	if _, err = fmt.Sscan(s, &u32); err != nil {
		return fmt.Errorf("vni: %q %v", s, err)
	} else if u32 >= 1<<24 {
		return fmt.Errorf("vni: %q %v", s, syscall.ERANGE)
	}
	m.tinfo = append(m.tinfo, rtnl.Attr{rtnl.IFLA_GENEVE_ID,
		rtnl.Uint32Attr(u32)})
	m.args = m.opt.Parms.More(m.args, []string{"id", "vni"})
	s = m.opt.Parms.ByName["remote"]
	if len(s) == 0 {
		return fmt.Errorf("missing remote")
	}
	if addr := net.ParseIP(s); addr == nil {
		return fmt.Errorf("remote: %q invalid", s)
	} else if ip4 := addr.To4(); ip4 != nil {
		m.tinfo = append(m.tinfo, rtnl.Attr{rtnl.IFLA_GENEVE_REMOTE,
			rtnl.BytesAttr(ip4)})
	} else {
		m.tinfo = append(m.tinfo, rtnl.Attr{rtnl.IFLA_GENEVE_REMOTE6,
			rtnl.BytesAttr(addr.To16())})
	}
	for _, x := range []struct {
		names []string
		t     uint16
	}{
		{[]string{"ttl", "hoplimit"}, rtnl.IFLA_GENEVE_TTL},
		{[]string{"tos", "dsfield"}, rtnl.IFLA_GENEVE_TOS},
	} {
		m.args = m.opt.Parms.More(m.args, x.names)
		s = m.opt.Parms.ByName[x.names[0]]
		if len(s) == 0 || s == "inherit" {
			continue
		}
		if _, err = fmt.Sscan(s, &u8); err != nil {
			return fmt.Errorf("%s: %q %v", x.names[0], s, err)
		}
		m.tinfo = append(m.tinfo, rtnl.Attr{x.t, rtnl.Uint8Attr(u8)})
	}
	m.args = m.opt.Parms.More(m.args, "flowlabel")
	if s = m.opt.Parms.ByName["flowlabel"]; len(s) > 0 {
		var u32 uint32
		if _, err = fmt.Sscan(s, &u32); err != nil {
			return fmt.Errorf("flowlabel: %q %v", s, err)
		}
		m.tinfo = append(m.tinfo, rtnl.Attr{rtnl.IFLA_GENEVE_LABEL,
			rtnl.Be32Attr(u32)})
	}
	m.args = m.opt.Parms.More(m.args, "dstport")
	if s = m.opt.Parms.ByName["dstport"]; len(s) > 0 {
		if _, err = fmt.Sscan(s, &u16); err != nil {
			return fmt.Errorf("dstport: %q %v", s, err)
		}
		m.tinfo = append(m.tinfo, rtnl.Attr{rtnl.IFLA_GENEVE_PORT,
			rtnl.Be16Attr(u16)})
	}
	m.args = m.opt.Flags.More(m.args,
		[]string{"external", "+external"},
		[]string{"no-external", "-external"},
	)
	if m.opt.Flags.ByName["external"] {
		m.tinfo = append(m.tinfo,
			rtnl.Attr{rtnl.IFLA_GENEVE_COLLECT_METADATA,
				rtnl.Uint8Attr(1)})
	} else if m.opt.Flags.ByName["no-external"] {
		m.tinfo = append(m.tinfo,
			rtnl.Attr{rtnl.IFLA_GENEVE_COLLECT_METADATA,
				rtnl.Uint8Attr(0)})
	}
	for _, x := range []struct {
		set   []string
		unset []string
		t     uint16
	}{
		{
			[]string{"udpcsum", "+udpcsum"},
			[]string{"no-udpcsum", "-udpcsum"},
			rtnl.IFLA_VXLAN_UDP_CSUM,
		},
		{
			[]string{"udp6zerocsumtx", "+udp6zerocsumtx"},
			[]string{"no-udp6zerocsumtx", "-udp6zerocsumtx"},
			rtnl.IFLA_VXLAN_UDP_ZERO_CSUM6_TX,
		},
		{
			[]string{"udp6zerocsumrx", "+udp6zerocsumrx"},
			[]string{"no-udp6zerocsumrx", "-udp6zerocsumrx"},
			rtnl.IFLA_VXLAN_UDP_ZERO_CSUM6_RX,
		},
	} {
		m.args = m.opt.Flags.More(m.args, x.set, x.unset)
		if m.opt.Flags.ByName[x.set[0]] {
			m.tinfo = append(m.tinfo, rtnl.Attr{x.t,
				rtnl.Uint8Attr(1)})
		} else if m.opt.Flags.ByName[x.set[0]] {
			m.tinfo = append(m.tinfo, rtnl.Attr{x.t,
				rtnl.Uint8Attr(0)})
		}
	}
	return nil
}
