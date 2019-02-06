// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package macvlan

import (
	"fmt"
	"net"
	"strings"

	"github.com/platinasystems/goes/cmd/ip/link/add/internal/options"
	"github.com/platinasystems/goes/cmd/ip/link/add/internal/request"
	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/goes/internal/nl"
	"github.com/platinasystems/goes/internal/nl/rtnl"
)

type Command string

func (c Command) String() string { return string(c) }

func (c Command) Usage() string {
	return fmt.Sprint("ip link add type ", c, " mode MODE [ OPTION ]...")
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "add a macvlan or macvtap link",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
MACV TYPES
	macvlan, macvtap
		macvlan just creates a virtual interface, while macvtap also
		make a character device /dev/tapX to be used just like tuntap.

MODES
	private
		Do not allow communication between macvlan instances on the
		same physical interface, even if the external switch supports
		hairpin mode.

	vepa	Virtual Ethernet Port Aggregator mode. Data from one macvlan
		instance to the other on the same phys‐ ical interface is
		transmitted over the physical inter‐ face.  Either the attached
		switch needs to support hair‐ pin mode, or there must be a
		TCP/IP router forwarding the packets in order to allow
		communication. This is the default mode.

	bridge	In bridge mode, all endpoints are directly connected to each
		other, communication is not redirected through the physical
		interface's peer.

	passthru [ nopromisc ]
		This mode gives more power to a single endpoint, usually in
		macvtap mode. It is not allowed for more than one endpoint on
		the same physical interface. All traffic will be forwarded to
		this end‐ point, allowing virtio guests to change MAC address
		or set promiscuous mode in order to bridge the interface or
		create vlan interfaces on top of it. By default, this mode
		forces the underlying interface into promiscuous mode. Passing
		the nopromisc flag prevents this, so the promisc flag may be
		controlled using standard tools.

	source [ macaddr { { add | del } LLADDR | set LLADDRS | flush } ]

		LLADDRS := LLADDDR[,LLADDR]...

SEE ALSO
	ip link add type man TYPE || ip link add type TYPE -man
	ip link man add || ip link add -man
	man ip || ip -man`,
	}
}

func (c Command) Main(args ...string) error {
	var info nl.Attrs

	opt, args := options.New(args)
	args = opt.Flags.More(args,
		[]string{
			"no-promisc", "-promisc",
			"no-promiscuous", "-promiscuous",
		},
		"macaddr",
		"flush",
	)
	opt.Parms.More(args,
		"mode",
		"add",
		"del",
		"set",
	)

	sock, err := nl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	sr := nl.NewSockReceiver(sock)

	if err = rtnl.MakeIfMaps(sr); err != nil {
		return err
	}

	add, err := request.New(opt, args)
	if err != nil {
		return err
	}

	s := opt.Parms.ByName["mode"]
	if len(s) == 0 {
		return fmt.Errorf("missing mode")
	}
	mode, found := map[string]uint32{
		"private":  rtnl.MACVLAN_MODE_PRIVATE,
		"vepa":     rtnl.MACVLAN_MODE_VEPA,
		"bridge":   rtnl.MACVLAN_MODE_BRIDGE,
		"passthru": rtnl.MACVLAN_MODE_PASSTHRU,
		"source":   rtnl.MACVLAN_MODE_SOURCE,
	}[s]
	if !found {
		return fmt.Errorf("mode: %q unknown", s)
	}
	info = append(info, nl.Attr{rtnl.IFLA_MACVLAN_MODE,
		nl.Uint32Attr(mode)})
	switch mode {
	case rtnl.MACVLAN_MODE_PASSTHRU:
		if opt.Flags.ByName["no-promisc"] {
			info = append(info, nl.Attr{rtnl.IFLA_MACVLAN_FLAGS,
				nl.Uint16Attr(rtnl.MACVLAN_FLAG_NOPROMISC)})
		}
	case rtnl.MACVLAN_MODE_SOURCE:
		if !opt.Flags.ByName["macaddr"] {
			// skip
		} else if opt.Flags.ByName["flush"] {
			info = append(info, nl.Attr{
				rtnl.IFLA_MACVLAN_MACADDR_MODE,
				nl.Uint32Attr(rtnl.MACVLAN_MACADDR_FLUSH)})
		} else if s = opt.Parms.ByName["add"]; len(s) > 0 {
			mac, err := net.ParseMAC(s)
			if err != nil {
				return fmt.Errorf("LLADDR: %q %v", s, err)
			}
			info = append(info, nl.Attr{
				rtnl.IFLA_MACVLAN_MACADDR_MODE,
				nl.Uint32Attr(rtnl.MACVLAN_MACADDR_ADD)})
			info = append(info, nl.Attr{rtnl.IFLA_MACVLAN_MACADDR,
				nl.BytesAttr(mac)})
		} else if s = opt.Parms.ByName["del"]; len(s) > 0 {
			mac, err := net.ParseMAC(s)
			if err != nil {
				return fmt.Errorf("LLADDR: %q %v", s, err)
			}
			info = append(info, nl.Attr{
				rtnl.IFLA_MACVLAN_MACADDR_MODE,
				nl.Uint32Attr(rtnl.MACVLAN_MACADDR_DEL)})
			info = append(info, nl.Attr{rtnl.IFLA_MACVLAN_MACADDR,
				nl.BytesAttr(mac)})
		} else if s = opt.Parms.ByName["set"]; len(s) > 0 {
			var macs nl.Attrs
			for _, smac := range strings.Split(s, ",") {
				mac, err := net.ParseMAC(smac)
				if err != nil {
					return fmt.Errorf("LLADDR: %q %v",
						smac, err)
				}
				macs = append(macs, nl.Attr{
					rtnl.IFLA_MACVLAN_MACADDR,
					nl.BytesAttr(mac)})
			}
			info = append(info, nl.Attr{
				rtnl.IFLA_MACVLAN_MACADDR_MODE,
				nl.Uint32Attr(rtnl.MACVLAN_MACADDR_SET)})
			info = append(info,
				nl.Attr{rtnl.IFLA_MACVLAN_MACADDR_DATA, macs})
		}
	}

	add.Attrs = append(add.Attrs, nl.Attr{rtnl.IFLA_LINKINFO, nl.Attrs{
		nl.Attr{rtnl.IFLA_INFO_KIND, nl.KstringAttr(c)},
		nl.Attr{rtnl.IFLA_INFO_DATA, info},
	}})
	req, err := add.Message()
	if err == nil {
		err = sr.UntilDone(req, nl.DoNothing)
	}
	return err
}
