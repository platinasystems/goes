// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netif

import (
	"fmt"
	"net"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/net/netlink"
)

const ConfigParameters = `
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
}

func (nif *Netif) Config(args []string) error {
	nl, err := netlink.Open()
	if err != nil {
		return err
	}
	defer nl.Close()

	if err = nif.refresh(nl); err != nil {
		return egress.Marked(err)
	}

	req, msg := netlink.Expand[netlink.NlMsghdr](nil)
	req.Type = netlink.RTM_NEWLINK
	req.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_ACK
	ifinfo, msg := netlink.Expand[netlink.IfInfomsg](msg)
	ifinfo.Family = netlink.AF_UNSPEC
	ifinfo.Index = int32(nif.Index)
	ifinfo.Change = 0
	ifinfo.Flags = uint32(nif.Flags)

	for len(args) > 0 {
		change := uint32(ConfigFlag[args[0]])
		ifla := ConfigAttr[args[0]]
		switch parameter := ConfigParameter[args[0]]; parameter {
		case UnknownParameter:
			return egress.Markf("%q %w", args[0], ErrInvalid)
		case TrueFlagParameter:
			ifinfo.Change |= change
			ifinfo.Flags |= change
			args = args[1:]
		case FalseFlagParameter:
			ifinfo.Change |= change
			ifinfo.Flags &^= change
			args = args[1:]
		case TrueAttrParameter:
			msg = netlink.CatAttr(msg, ifla, uint8(1))
			args = args[1:]
		case FalseAttrParameter:
			msg = netlink.CatAttr(msg, ifla, uint8(0))
			args = args[1:]
		case Uint32Parameter:
			if len(args) < 2 {
				return egress.Markf("%q %w",
					args[0], ErrIncomplete)
			}
			var v uint32
			_, err = fmt.Sscan(args[1], &v)
			if err != nil {
				return egress.Markf("%q %w", args[1], err)
			}
			msg = netlink.CatAttr(msg, ifla, v)
			args = args[2:]
		case Int32Parameter:
			if len(args) < 2 {
				return egress.Markf("%q %w",
					args[0], ErrIncomplete)
			}
			var v int32
			_, err = fmt.Sscan(args[1], &v)
			if err != nil {
				return egress.Markf("%q %w", args[1], err)
			}
			msg = netlink.CatAttr(msg, ifla, v)
			args = args[2:]
		case HardwareAddrParameter:
			if len(args) < 2 {
				return egress.Markf("%q %w",
					args[0], ErrIncomplete)
			}
			v, err := net.ParseMAC(args[1])
			if err != nil {
				return egress.Markf("%q %w", args[1], err)
			}
			msg = netlink.CatBytesAttr(msg, ifla, v)
			args = args[2:]
		case StringParameter:
			if len(args) < 2 {
				return egress.Markf("%q %w",
					args[0], ErrIncomplete)
			}
			msg = netlink.CatStringAttr(msg, ifla, args[1])
			args = args[2:]
		case LinkModeParameter:
			if len(args) < 2 {
				return egress.Markf("%q %w",
					args[0], ErrIncomplete)
			}
			mode, ok := netlink.IfLinkModeByName[args[1]]
			if !ok {
				return egress.Markf("%q %w",
					args[1], ErrInvalid)
			}
			msg = netlink.CatAttr(msg, ifla, mode)
			args = args[2:]
		case StateParameter:
			if len(args) < 2 {
				return egress.Markf("%q %w",
					args[0], ErrIncomplete)
			}
			op, ok := netlink.IfOperByName[args[1]]
			if !ok {
				return egress.Markf("%q %w",
					args[1], ErrInvalid)
			}
			msg = netlink.CatAttr(msg, ifla, op)
			args = args[2:]
		default:
			return fmt.Errorf("%q %w", args[0], ErrInvalid)
		}
	}
	if err = nl.Request(msg); err == nil {
		err = nl.Wait(req.Seq)
	}
	return err
}

func (nif *Netif) refresh(nl *netlink.Netlink) error {
	req, msg := netlink.ExpandNlMsghdr(nil)
	req.Type = netlink.RTM_GETLINK
	req.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_ACK
	ifinfo, msg := netlink.ExpandIfInfomsg(msg)
	ifinfo.Family = netlink.AF_UNSPEC
	ifinfo.Index = int32(nif.Index)
	if err := nl.Request(msg); err != nil {
		return err
	}
	for {
		rsp, data, err := nl.Next()
		if err != nil {
			return err
		} else if rsp.Seq != req.Seq {
			continue
		} else if rsp.Type == netlink.NLMSG_DONE {
			return nil
		} else if rsp.Type == netlink.NLMSG_ERROR {
			return netlink.ExtractError(data)
		} else if rsp.Type != netlink.RTM_NEWLINK {
			continue
		}
		return nif.ifinfo(data)
	}
}
