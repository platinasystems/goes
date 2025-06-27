// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"encoding/binary"
	"fmt"
	"io"
	"net/netip"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netpdu"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

const (
	VPN_P_HELLO = 0x400 + iota
	_
	_
	_
	_
)

const (
	VPN_P_IP  = netph.ETH_P_IP
	VPN_P_IP6 = netph.ETH_P_IPV6
	VPNSize   = netph.TunPISize
)

var zaddr netip.Addr

var PacketName = map[uint16]string{
	VPN_P_HELLO: "vpn-hello",
}

type PDU []byte
type HelloPDU []byte

func (pdu PDU) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "vpn")
	h, d, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, " ", err)
		return
	}
	switch h.Proto {
	case VPN_P_HELLO:
		fmt.Fprint(w, netpdu.Mark, HelloPDU(d))
	case VPN_P_IP:
		fmt.Fprint(w, netpdu.Mark, netpdu.IP(d))
	case VPN_P_IP6:
		fmt.Fprint(w, netpdu.Mark, netpdu.IP6(d))
	default:
		fmt.Fprintf(w, " proto %#x", h.Proto)
	}
}

func (pdu PDU) Parse() (netph.TunPI, []byte, error) {
	return netph.Parse[netph.TunPI](pdu)
}

func (pdu HelloPDU) Format(w fmt.State, verb rune) {
	var t int64
	fmt.Fprint(w, "hello ")
	if _, err := xnet.Remove(pdu, &t); err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprint(w, time.UnixMicro(t))
	}
}

func WriteAddrTo(w io.Writer, addr netip.Addr) (int64, error) {
	a16 := addr.As16()
	n, err := w.Write(a16[:])
	return int64(n), err
}

func WriteAddrPortTo(w io.Writer, ap netip.AddrPort) (int64, error) {
	n, err := WriteAddrTo(w, ap.Addr())
	if err == nil {
		err = binary.Write(w, binary.BigEndian, ap.Port())
		if err == nil {
			n += 2
		}
	}
	return n, err
}
