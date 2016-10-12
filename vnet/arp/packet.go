package arp

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip4"

	"bytes"
	"fmt"
	"unsafe"
)

type Opcode uint16

const (
	Request Opcode = 1 + iota
	Reply
	ReverseRequest
	ReverseReply
)

var opcodeStrings = [...]string{
	Request:        "request",
	Reply:          "reply",
	ReverseRequest: "reverse-request",
	ReverseReply:   "reverse-reply",
}

type L2Type uint16

const (
	L2TypeEthernet L2Type = 1
)

var l2TypeStrings = [...]string{
	L2TypeEthernet: "ethernet",
}

func (x Opcode) String() string { return elib.StringerWithFormat(opcodeStrings[:], int(x), "0x%x") }
func (x L2Type) String() string { return elib.StringerWithFormat(l2TypeStrings[:], int(x), "0x%x") }

type Header struct {
	L2Type          vnet.Uint16
	L3Type          vnet.Uint16
	NL2AddressBytes uint8
	NL3AddressBytes uint8
	Opcode          vnet.Uint16
}

func (h *Header) GetOpcode() Opcode        { return Opcode(h.Opcode.ToHost()) }
func (h *Header) GetL2Type() L2Type        { return L2Type(h.L2Type.ToHost()) }
func (h *Header) GetL3Type() ethernet.Type { return ethernet.Type(h.L3Type.ToHost()) }

func (x Opcode) FromHost() vnet.Uint16 { return vnet.Uint16(x).FromHost() }
func (x L2Type) FromHost() vnet.Uint16 { return vnet.Uint16(x).FromHost() }

// Typical case for arp: ip4 over ethernet.
type EthernetIp4Addr struct {
	Ethernet ethernet.Address
	Ip4      ip4.Address
}

type HeaderEthernetIp4 struct {
	Header
	Addrs [2]EthernetIp4Addr
}

const HeaderEthernetIp4Bytes = 8 + 6 + 4

func (h *HeaderEthernetIp4) String() (s string) {
	s = fmt.Sprintf("%s, l2/l3 type/size %s/%d %s/%d, %s/%s -> %s/%s",
		h.GetOpcode().String(),
		h.GetL2Type().String(), h.Header.NL2AddressBytes,
		h.GetL3Type().String(), h.Header.NL3AddressBytes,
		&h.Addrs[0].Ethernet, &h.Addrs[0].Ip4,
		&h.Addrs[1].Ethernet, &h.Addrs[1].Ip4)
	return
}

// Implement vnet.Header interface.
func (h *HeaderEthernetIp4) Len() uint                      { return HeaderEthernetIp4Bytes }
func (h *HeaderEthernetIp4) Finalize(l []vnet.PacketHeader) {}
func (h *HeaderEthernetIp4) Write(b *bytes.Buffer) {
	type t struct{ data [unsafe.Sizeof(*h)]byte }
	i := (*t)(unsafe.Pointer(h))
	b.Write(i.data[:])
}
func (h *HeaderEthernetIp4) Read(b []byte) vnet.PacketHeader {
	return (*HeaderEthernetIp4)(vnet.Pointer(b))
}

func GetHeader(r *vnet.Ref) *HeaderEthernetIp4      { return (*HeaderEthernetIp4)(r.Data()) }
func GetPacketHeader(r *vnet.Ref) vnet.PacketHeader { return GetHeader(r) }
