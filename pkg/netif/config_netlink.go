// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netif

import (
	"context"
	"fmt"
	"net"

	"github.com/platinasystems/goes/v2/pkg/integer"
	"github.com/platinasystems/goes/v2/pkg/netlink"
	"github.com/platinasystems/goes/v2/pkg/netlink/iflink"
	"github.com/platinasystems/goes/v2/pkg/netlink/rtnetlink"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xnet"
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

var ConfigFlag = map[string]uint{
	"up":     iflink.IFF_UP,
	"-arp":   iflink.IFF_NOARP,
	"no-arp": iflink.IFF_NOARP,
	"down":   iflink.IFF_UP,
	"arp":    iflink.IFF_NOARP,
}

var ConfigAttr = map[string]uint{
	"carrier":      iflink.IFLA_CARRIER,
	"protodown":    iflink.IFLA_PROTO_DOWN,
	"-protodown":   iflink.IFLA_PROTO_DOWN,
	"no-protodown": iflink.IFLA_PROTO_DOWN,
	"-carrier":     iflink.IFLA_CARRIER,
	"no-carrier":   iflink.IFLA_CARRIER,
	"protoup":      iflink.IFLA_PROTO_DOWN,
	"link-netnsid": iflink.IFLA_LINK_NETNSID,
	"mtu":          iflink.IFLA_MTU,
	"numtxqueues":  iflink.IFLA_NUM_TX_QUEUES,
	"numrxqueues":  iflink.IFLA_NUM_RX_QUEUES,
	"txqueuelen":   iflink.IFLA_TXQLEN,
	"master":       iflink.IFLA_MASTER,
	"vrf":          iflink.IFLA_MASTER,
	"address":      iflink.IFLA_ADDRESS,
	"ether":        iflink.IFLA_ADDRESS,
	"lladdr":       iflink.IFLA_ADDRESS,
	"mode":         iflink.IFLA_LINKMODE,
	"state":        iflink.IFLA_OPERSTATE,
}

func (nif *NetIf) Config(ctx context.Context, args []string) error {
	nl, err := netlink.Open()
	if err != nil {
		return err
	}
	defer nl.Close()

	if err = nif.refresh(ctx, nl); err != nil {
		return xerrors.Mark(err)
	}

	req, msg := netlink.Expand[netlink.MsgHdr](nil)
	req.Type = rtnetlink.RTM_NEWLINK
	req.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_ACK
	ifinfo, msg := netlink.Expand[rtnetlink.IfInfoMsg](msg)
	ifinfo.Family = xnet.AF_UNSPEC
	ifinfo.Index = int32(nif.Index)
	ifinfo.Change = 0
	ifinfo.Flags = uint32(nif.Flags)

	for len(args) > 0 {
		change := ConfigFlag[args[0]]
		ifla := ConfigAttr[args[0]]
		switch parameter := ConfigParameter[args[0]]; parameter {
		case UnknownParameter:
			return xerrors.Invalid(args[0])
		case TrueFlagParameter:
			integer.Set(&ifinfo.Change, change)
			integer.Set(&ifinfo.Flags, change)
			args = args[1:]
		case FalseFlagParameter:
			integer.Set(&ifinfo.Change, change)
			integer.Reset(&ifinfo.Flags, change)
			args = args[1:]
		case TrueAttrParameter:
			msg = netlink.CatAttr(msg, ifla, uint8(1))
			args = args[1:]
		case FalseAttrParameter:
			msg = netlink.CatAttr(msg, ifla, uint8(0))
			args = args[1:]
		case Uint32Parameter:
			if len(args) < 2 {
				return xerrors.Incomplete(args[0])
			}
			var v uint32
			_, err = fmt.Sscan(args[1], &v)
			if err != nil {
				return xerrors.Label(err, args[1])
			}
			msg = netlink.CatAttr(msg, ifla, v)
			args = args[2:]
		case Int32Parameter:
			if len(args) < 2 {
				return xerrors.Incomplete(args[0])
			}
			var v int32
			_, err = fmt.Sscan(args[1], &v)
			if err != nil {
				return xerrors.Label(err, args[1])
			}
			msg = netlink.CatAttr(msg, ifla, v)
			args = args[2:]
		case HardwareAddrParameter:
			if len(args) < 2 {
				return xerrors.Incomplete(args[0])
			}
			v, err := net.ParseMAC(args[1])
			if err != nil {
				return xerrors.Label(err, args[1])
			}
			msg = netlink.CatBytesAttr(msg, ifla, v)
			args = args[2:]
		case StringParameter:
			if len(args) < 2 {
				return xerrors.Incomplete(args[0])
			}
			msg = netlink.CatStringAttr(msg, ifla, args[1])
			args = args[2:]
		case LinkModeParameter:
			if len(args) < 2 {
				return xerrors.Incomplete(args[0])
			}
			mode := iflink.IfLinkModeByName(args[1])
			if mode == iflink.INVALID_IF_LINK_MODE {
				return xerrors.Invalid(args[1])
			}
			msg = netlink.CatAttr(msg, ifla, mode)
			args = args[2:]
		case StateParameter:
			if len(args) < 2 {
				return xerrors.Incomplete(args[0])
			}
			op := iflink.IfOperByName(args[1])
			if op == iflink.INVALID_IF_OPER {
				return xerrors.Invalid(args[1])
			}
			msg = netlink.CatAttr(msg, ifla, op)
			args = args[2:]
		default:
			return fmt.Errorf("%q %w", args[0], ErrInvalid)
		}
	}
	if err = nl.Request(msg); err == nil {
		err = nl.Wait(ctx, req.SEQ)
	}
	return err
}

func (nif *NetIf) refresh(ctx context.Context, nl *netlink.NL) error {
	hdr, req := netlink.ExpandMsgHdr(nil)
	hdr.Type = rtnetlink.RTM_GETLINK
	hdr.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_ACK
	ifinfo, req := netlink.ExpandIfInfoMsg(req)
	ifinfo.Family = xnet.AF_UNSPEC
	ifinfo.Index = int32(nif.Index)
	if err := nl.Request(req); err != nil {
		return err
	}
	seq := hdr.SEQ
	for {
		rsp, data, err := nl.Next(ctx)
		if err != nil {
			return err
		} else if rsp.SEQ != seq {
			continue
		} else if rsp.Type == netlink.NLMSG_DONE {
			return nil
		} else if rsp.Type == netlink.NLMSG_ERROR {
			e, _ := netlink.ExtractMsgErr(data)
			return e.Err()
		} else if rsp.Type != rtnetlink.RTM_NEWLINK {
			continue
		}
		return nif.ifinfo(data)
	}
}
