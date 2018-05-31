/* Copyright(c) 2018 Platina Systems, Inc.
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms and conditions of the GNU General Public License,
 * version 2, as published by the Free Software Foundation.
 *
 * This program is distributed in the hope it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
 * more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * this program; if not, write to the Free Software Foundation, Inc.,
 * 51 Franklin St - Fifth Floor, Boston, MA 02110-1301 USA.
 *
 * The full GNU General Public License is included in this distribution in
 * the file called "COPYING".
 *
 * Contact Information:
 * sw@platina.com
 * Platina Systems, 3180 Del La Cruz Blvd, Santa Clara, CA 95054
 */
package xeth

import (
	"fmt"
	"net"
	"unsafe"
)

func IsMsg(b []byte) bool {
	if len(b) < SizeofMsg {
		return false
	}
	msg := (*Msg)(unsafe.Pointer(&b[0]))
	return msg.Z64 == 0 && msg.Z32 == 0 && msg.Z16 == 0 && msg.Z8 == 0
}

func (msg *MsgStat) String() string {
	var stat string
	kind := Kind(msg.Kind)
	switch kind {
	case XETH_MSG_KIND_LINK_STAT:
		stat = LinkStat(msg.Index).String()
	case XETH_MSG_KIND_ETHTOOL_STAT:
		stat = EthtoolStat(msg.Index).String()
	default:
		stat = "unknown"
	}
	return fmt.Sprint(kind, " ", (*Ifname)(&msg.Ifname), " ", stat, " ",
		msg.Count)
}

func (msg *MsgEthtoolFlags) String() string {
	return fmt.Sprint(Kind(msg.Kind), " ", (*Ifname)(&msg.Ifname), " ",
		EthtoolFlagBits(msg.Flags))
}

func (msg *MsgEthtoolSettings) String() string {
	return fmt.Sprint(Kind(msg.Kind), " ", (*Ifname)(&msg.Ifname),
		"\n\tspeed: ", Mbps(msg.Speed),
		"\n\tduplex: ", Duplex(msg.Duplex),
		"\n\tport: ", Port(msg.Port),
		"\n\tautoneg: ", Autoneg(msg.Autoneg),
		"\n\tsupported:",
		(*EthtoolLinkModeBits)(&msg.Link_modes_supported),
		"\n\tadvertising:",
		(*EthtoolLinkModeBits)(&msg.Link_modes_advertising),
		"\n\tpartner:",
		(*EthtoolLinkModeBits)(&msg.Link_modes_lp_advertising),
	)
}

func (msg *MsgCarrier) String() string {
	return fmt.Sprint(Kind(msg.Kind), " ", (*Ifname)(&msg.Ifname), " ",
		CarrierFlag(msg.Flag))
}

func (msg *MsgSpeed) String() string {
	return fmt.Sprint(Kind(msg.Kind), " ", (*Ifname)(&msg.Ifname), " ",
		Mbps(msg.Mbps))
}

func (msg *MsgIfindex) String() string {
	return fmt.Sprint(Kind(msg.Kind),
		" ", (*Ifname)(&msg.Ifname),
		"\n\tindex: ", msg.Ifindex,
		"\n\tnet: ", fmt.Sprintf("%#x", msg.Net),
	)
}

func (msg *MsgIfa) String() string {
	return fmt.Sprint(Kind(msg.Kind), " ", (*Ifname)(&msg.Ifname), " ",
		AddressEvent(msg.Event), " ", msg.IPNet())
}

func (msg *MsgIfa) IsAdd() bool { return Event(msg.Event) == NETDEV_UP }
func (msg *MsgIfa) IsDel() bool { return Event(msg.Event) == NETDEV_DOWN }

func (msg *MsgIfa) IPNet() *net.IPNet {
	ipBuf := make([]byte, 4)
	maskBuf := make([]byte, 4)
	*(*uint32)(unsafe.Pointer(&ipBuf[0])) = msg.Address
	*(*uint32)(unsafe.Pointer(&maskBuf[0])) = msg.Mask
	return &net.IPNet{net.IP(ipBuf), net.IPMask(maskBuf)}
}
