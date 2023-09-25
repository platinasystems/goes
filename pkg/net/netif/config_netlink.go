// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netif

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/net/netlink"
)

const ConfigParameters = `
Config Parameters
  {add | alias} <prefix>
	Add network <prefix> to interface.
  {delete | -alias} <prefix>
	Remove network <prefix> from interface.
  {change | replace} <prefix>
	Change or replace the current prefix with the preceding <prefix>.
  peer <address>
  	Set <address> of the interface's point-to-point peer. 
  up | down
	Enable/Disable named interface.
  arp | -arp
	Enable/Disable Address Resolution Protocol.
  mtu <number>
	Set maximum transmission unit (octets).
  mode { default, dormant }
  state { not-present, down, lower-down, testing, dormant, up }
  ether <xx:xx:xx:xx:xx:xx>
	Set the named interfaces's 6-byte link-level address from colon
	sepearated, hex digits.
`

var ConfigParameter = map[string]Parameter{
	"up":           TrueFlagParameter,
	"-arp":         TrueFlagParameter,
	"no-arp":       TrueFlagParameter,
	"down":         FalseFlagParameter,
	"arp":          FalseFlagParameter,
	"carrier":      TrueAttrParameter,
	"protodown":    TrueAttrParameter,
	"-protodown":   TrueAttrParameter,
	"no-protodown": TrueAttrParameter,
	"-carrier":     FalseAttrParameter,
	"no-carrier":   FalseAttrParameter,
	"protoup":      FalseAttrParameter,
	"link-netnsid": Uint32Parameter,
	"mtu":          Uint32Parameter,
	"numtxqueues":  Uint32Parameter,
	"numrxqueues":  Uint32Parameter,
	"txqueuelen":   Uint32Parameter,
	"master":       Int32Parameter,
	"vrf":          Int32Parameter,
	"address":      HardwareAddrParameter,
	"ether":        HardwareAddrParameter,
	"lladdr":       HardwareAddrParameter,
	"mode":         LinkModeParameter,
	"state":        StateParameter,
	"add":          PrefixParameter,
	"alias":        PrefixParameter,
	"del":          PrefixParameter,
	"-alias":       PrefixParameter,
	"change":       PrefixParameter,
	"replace":      PrefixParameter,
	"peer":         AddrParameter,
}

var ConfigFlag = map[string]IFF{
	"up":     IFF_UP,
	"-arp":   IFF_NOARP,
	"no-arp": IFF_NOARP,
	"down":   IFF_UP,
	"arp":    IFF_NOARP,
}

var ConfigAttr = map[string]uint16{
	"carrier":      netlink.IFLA_CARRIER,
	"protodown":    netlink.IFLA_PROTO_DOWN,
	"-protodown":   netlink.IFLA_PROTO_DOWN,
	"no-protodown": netlink.IFLA_PROTO_DOWN,
	"-carrier":     netlink.IFLA_CARRIER,
	"no-carrier":   netlink.IFLA_CARRIER,
	"protoup":      netlink.IFLA_PROTO_DOWN,
	"link-netnsid": netlink.IFLA_LINK_NETNSID,
	"mtu":          netlink.IFLA_MTU,
	"numtxqueues":  netlink.IFLA_NUM_TX_QUEUES,
	"numrxqueues":  netlink.IFLA_NUM_RX_QUEUES,
	"txqueuelen":   netlink.IFLA_TXQLEN,
	"master":       netlink.IFLA_MASTER,
	"vrf":          netlink.IFLA_MASTER,
	"address":      netlink.IFLA_ADDRESS,
	"ether":        netlink.IFLA_ADDRESS,
	"lladdr":       netlink.IFLA_ADDRESS,
	"mode":         netlink.IFLA_LINKMODE,
	"state":        netlink.IFLA_OPERSTATE,
	"add":          netlink.IFA_LOCAL,
	"del":          netlink.IFA_LOCAL,
	"change":       netlink.IFA_LOCAL,
	"replace":      netlink.IFA_LOCAL,
	"peer":         netlink.IFA_ADDRESS,
}

func (nif *Netif) Config(args ...string) error {
	nl, err := netlink.Open()
	if err != nil {
		return err
	}
	defer nl.Close()

	if err = nif.refresh(nl); err != nil {
		return egress.Marked(err)
	}

	ifahdr, ifareq := netlink.Expand[netlink.NlMsghdr](nil)
	ifahdr.Type = netlink.RTM_NEWADDR
	ifahdr.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_ACK
	ifa, ifareq := netlink.Expand[netlink.IfAddrmsg](ifareq)
	ifa.Family = netlink.AF_UNSPEC
	ifa.Index = uint32(nif.Index)
	baseIfareqLen := len(ifareq)

	iflhdr, iflreq := netlink.Expand[netlink.NlMsghdr](nil)
	iflhdr.Type = netlink.RTM_NEWLINK
	iflhdr.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_ACK
	ifinfo, iflreq := netlink.Expand[netlink.IfInfomsg](iflreq)
	ifinfo.Family = netlink.AF_UNSPEC
	ifinfo.Index = int32(nif.Index)
	ifinfo.Change = 0
	ifinfo.Flags = uint32(nif.Flags)
	baseIflreqLen := len(iflreq)

	for len(args) > 0 {
		change := uint32(ConfigFlag[args[0]])
		ifla := ConfigAttr[args[0]]
		switch parameter := ConfigParameter[args[0]]; parameter {
		case UnknownParameter:
			return fmt.Errorf("%q %w", args[0], ErrInvalid)
		case TrueFlagParameter:
			ifinfo.Change |= change
			ifinfo.Flags |= change
			args = args[1:]
		case FalseFlagParameter:
			ifinfo.Change |= change
			ifinfo.Flags &^= change
			args = args[1:]
		case TrueAttrParameter:
			iflreq = netlink.CatAttr(iflreq, ifla, uint8(1))
			args = args[1:]
		case FalseAttrParameter:
			iflreq = netlink.CatAttr(iflreq, ifla, uint8(0))
			args = args[1:]
		case Uint32Parameter:
			var v uint32
			if len(args) < 2 {
				return ErrIncomplete
			} else if _, err = fmt.Sscan(args[1], &v); err != nil {
				return fmt.Errorf("%q %w", args[1], err)
			}
			iflreq = netlink.CatAttr(iflreq, ifla, v)
			args = args[2:]
		case Int32Parameter:
			var v int32
			if len(args) < 2 {
				return ErrIncomplete
			} else if _, err = fmt.Sscan(args[1], &v); err != nil {
				return fmt.Errorf("%q %w", args[1], err)
			}
			iflreq = netlink.CatAttr(iflreq, ifla, v)
			args = args[2:]
		case HardwareAddrParameter:
			var v net.HardwareAddr
			if len(args) < 2 {
				return ErrIncomplete
			} else if v, err = net.ParseMAC(args[1]); err != nil {
				return fmt.Errorf("%q %w", args[1], err)
			}
			iflreq = netlink.CatBytesAttr(iflreq, ifla, v)
			args = args[2:]
		case StringParameter:
			if len(args) < 2 {
				return ErrIncomplete
			}
			iflreq = netlink.CatStringAttr(iflreq, ifla, args[1])
			args = args[2:]
		case LinkModeParameter:
			if len(args) < 2 {
				return ErrIncomplete
			}
			mode, ok := netlink.IfLinkModeByName[args[1]]
			if !ok {
				return fmt.Errorf("%q %w", args[1], ErrInvalid)
			}
			iflreq = netlink.CatAttr(iflreq, ifla, mode)
			args = args[2:]
		case StateParameter:
			if len(args) < 2 {
				return ErrIncomplete
			}
			op, ok := netlink.IfOperByName[args[1]]
			if !ok {
				return fmt.Errorf("%q %w", args[1], ErrInvalid)
			}
			iflreq = netlink.CatAttr(iflreq, ifla, op)
			args = args[2:]
		case PrefixParameter:
			if len(args) < 2 {
				return ErrIncomplete
			}
			prefix, err := netip.ParsePrefix(args[1])
			if err != nil {
				return fmt.Errorf("%q %w", args[1], err)
			}
			if prefix.Addr().Is4() {
				ifa.Family = netlink.AF_INET
			} else if prefix.Addr().Is6() {
				ifa.Family = netlink.AF_INET6
			} else {
				return fmt.Errorf("%q %w", args[0], ErrInvalid)
			}
			v := prefix.Addr().AsSlice()
			ifareq = netlink.CatBytesAttr(ifareq, ifla, v)
			ifa.Prefixlen = uint8(prefix.Bits())
			switch args[0] {
			case "add":
				ifahdr.Type = netlink.RTM_NEWADDR
				ifahdr.Flags |= netlink.NLM_F_CREATE
				ifahdr.Flags |= netlink.NLM_F_EXCL
			case "del":
				ifahdr.Type = netlink.RTM_DELADDR
			case "change":
				ifahdr.Type = netlink.RTM_NEWADDR
				ifahdr.Flags |= netlink.NLM_F_REPLACE
			case "replace":
				ifahdr.Type = netlink.RTM_NEWADDR
				ifahdr.Flags |= netlink.NLM_F_CREATE
				ifahdr.Flags |= netlink.NLM_F_EXCL
			}
			args = args[2:]
		case AddrParameter:
			if len(args) < 2 {
				return ErrIncomplete
			}
			peer, err := netip.ParseAddr(args[1])
			if err != nil {
				return fmt.Errorf("%q %w", args[1], err)
			}
			v := peer.AsSlice()
			ifareq = netlink.CatBytesAttr(ifareq, ifla, v)
			args = args[2:]
		default:
			return fmt.Errorf("%q %w", args[0], ErrInvalid)
		}
	}
	if ifahdr.Type != 0 && len(ifareq) > baseIfareqLen {
		if err = nl.Request(ifareq, nil); err != nil {
			return egress.Marked(err)
		}
	}
	if ifinfo.Change != 0 || len(iflreq) > baseIflreqLen {
		if err = nl.Request(iflreq, nil); err != nil {
			return egress.Marked(err)
		}
	}
	return nil
}

func (nif *Netif) refresh(nl *netlink.Netlink) error {
	iflhdr, iflreq := netlink.ExpandNlMsghdr(nil)
	iflhdr.Type = netlink.RTM_GETLINK
	iflhdr.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_ACK
	ifinfo, iflreq := netlink.ExpandIfInfomsg(iflreq)
	ifinfo.Family = netlink.AF_UNSPEC
	ifinfo.Index = int32(nif.Index)
	return nl.Request(iflreq, nif.update)
}
