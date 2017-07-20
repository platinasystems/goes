// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netlink

import (
	"bytes"
	"fmt"
	"io"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/platinasystems/go/internal/accumulate"
	"github.com/platinasystems/go/internal/indent"
)

type Byter interface {
	Bytes() []byte
}

type Message interface {
	// The Message Closer returns itself to its respective pool.
	io.Closer

	MsgHeader() *Header
	MsgLen() int
	MsgType() MsgType
	Nsid() *int

	// The Message Reader stores itself in a given byte buffer.
	io.Reader

	// The Message Stringer returns itself as a formatted string.
	fmt.Stringer

	// The Message Writer loads itself from a given byte buffer.
	io.Writer

	// The Message WriterTo string formats itself into the given writer.
	io.WriterTo
}

type multiliner interface {
	multiline()
}

type Runer interface {
	Rune() rune
}

func StringOf(wt io.WriterTo) string {
	buf := pool.Bytes.Get().(*bytes.Buffer)
	defer repool(buf)
	wt.WriteTo(buf)
	return buf.String()
}

type Header struct {
	Len      uint32
	Type     MsgType
	Flags    HeaderFlags
	Sequence uint32
	Pid      uint32
}

const SizeofHeader = 4 + SizeofMsgType + SizeofHeaderFlags + 4 + 4

func (h *Header) MsgHeader() *Header { return h }
func (h *Header) MsgType() MsgType   { return h.Type }

// Roundup netlink message length for proper alignment.
func (h *Header) MsgLen() int {
	return (int(h.Len) + NLMSG_ALIGNTO - 1) & ^(NLMSG_ALIGNTO - 1)
}

func (h *Header) Read(b []byte) (int, error) {
	if len(b) < SizeofHeader {
		return 0, syscall.EOVERFLOW
	}
	*(*Header)(unsafe.Pointer(&b[0])) = *h
	return SizeofHeader, nil
}

func (h *Header) String() string {
	return StringOf(h)
}

func (h *Header) Write(b []byte) (int, error) {
	if len(b) < SizeofHeader {
		return 0, syscall.EOVERFLOW
	}
	*h = *(*Header)(unsafe.Pointer(&b[0]))
	return SizeofHeader, nil
}

func (h *Header) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprintln(acc, "len:", h.Len)
	fmt.Fprintln(acc, "seq:", h.Sequence)
	fmt.Fprintln(acc, "pid:", h.Pid)
	if h.Flags != 0 {
		fmt.Fprintln(acc, "flags:", h.Flags)
	}
	return acc.Tuple()
}

func NewGenMessage() *GenMessage {
	m := pool.GenMessage.Get().(*GenMessage)
	runtime.SetFinalizer(m, (*GenMessage).Close)
	m.nsid = DefaultNsid
	m.Header.Len = SizeofGenMessage
	return m
}

type GenMessage struct {
	nsid int
	Header
	Genmsg
}

const SizeofGenMessage = SizeofHeader + SizeofGenmsg

type Genmsg struct {
	AddressFamily
}

const SizeofGenmsg = SizeofAddressFamily

func (m *GenMessage) Close() error {
	runtime.SetFinalizer(m, nil)
	repool(m)
	return nil
}

func (m *GenMessage) Nsid() *int { return &m.nsid }

func (m *GenMessage) Read(b []byte) (int, error) {
	if len(b) < SizeofGenMessage {
		return 0, syscall.EOVERFLOW
	}
	n, _ := m.Header.Read(b)
	*(*Genmsg)(unsafe.Pointer(&b[n])) = m.Genmsg
	n += SizeofGenmsg
	return n, nil
}

func (m *GenMessage) String() string { return StringOf(m) }

func (m *GenMessage) Write(b []byte) (int, error) {
	if len(b) < SizeofGenMessage {
		return 0, syscall.EOVERFLOW
	}
	n, _ := m.Header.Write(b)
	m.Genmsg = *(*Genmsg)(unsafe.Pointer(&b[n]))
	n += SizeofGenmsg
	return n, nil
}

func (m *GenMessage) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, m.Header.Type, ":\n")
	indent.Increase(acc)
	if m.nsid != DefaultNsid {
		fmt.Fprintln(acc, "nsid:", m.nsid)
	}
	m.Header.WriteTo(acc)
	fmt.Fprintln(acc, "family:", m.AddressFamily)
	indent.Decrease(acc)
	return acc.Tuple()
}

type NoopMessage struct {
	nsid int
	Header
}

const SizeofNoopMessage = SizeofHeader

func NewNoopMessage() *NoopMessage {
	m := pool.NoopMessage.Get().(*NoopMessage)
	runtime.SetFinalizer(m, (*NoopMessage).Close)
	m.nsid = DefaultNsid
	return m
}

func (m *NoopMessage) Close() error {
	runtime.SetFinalizer(m, nil)
	repool(m)
	return nil
}

func (m *NoopMessage) Nsid() *int { return &m.nsid }

func (m *NoopMessage) Read(b []byte) (int, error) {
	if len(b) < SizeofNoopMessage {
		return 0, syscall.EOVERFLOW
	}
	return m.Header.Read(b)
}

func (m *NoopMessage) String() string { return StringOf(m) }

func (m *NoopMessage) Write(b []byte) (int, error) {
	if len(b) < SizeofNoopMessage {
		return 0, syscall.EOVERFLOW
	}
	return m.Header.Write(b)
}

func (m *NoopMessage) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, MessageType(m.Header.Type), ":\n")
	indent.Increase(acc)
	if m.nsid != DefaultNsid {
		fmt.Fprintln(acc, "nsid:", m.nsid)
	}
	m.Header.WriteTo(acc)
	indent.Decrease(acc)
	return acc.Tuple()
}

type DoneMessage struct {
	nsid int
	Header
}

const SizeofDoneMessage = SizeofHeader

func NewDoneMessage() *DoneMessage {
	m := pool.DoneMessage.Get().(*DoneMessage)
	runtime.SetFinalizer(m, (*DoneMessage).Close)
	m.nsid = DefaultNsid
	return m
}

func (m *DoneMessage) Close() error {
	runtime.SetFinalizer(m, nil)
	repool(m)
	return nil
}

func (m *DoneMessage) Nsid() *int { return &m.nsid }

func (m *DoneMessage) Read(b []byte) (int, error) {
	if len(b) < SizeofDoneMessage {
		return 0, syscall.EOVERFLOW
	}
	return m.Header.Read(b)
}

func (m *DoneMessage) String() string { return StringOf(m) }

func (m *DoneMessage) Write(b []byte) (int, error) {
	if len(b) < SizeofDoneMessage {
		return 0, syscall.EOVERFLOW
	}
	return m.Header.Write(b)
}

func (m *DoneMessage) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, MessageType(m.Header.Type), ":\n")
	indent.Increase(acc)
	if m.nsid != DefaultNsid {
		fmt.Fprintln(acc, "nsid:", m.nsid)
	}
	m.Header.WriteTo(acc)
	indent.Decrease(acc)
	return acc.Tuple()
}

type ErrorMessage struct {
	nsid int
	Header
	Errormsg
}

const SizeofErrorMessage = SizeofHeader + SizeofErrormsg

type Errormsg struct {
	// Unix errno for error.
	Errno int32
	// Header for message with error.
	Req Header
}

const SizeofErrormsg = 4 + SizeofHeader

func NewErrorMessage() *ErrorMessage {
	m := pool.ErrorMessage.Get().(*ErrorMessage)
	runtime.SetFinalizer(m, (*ErrorMessage).Close)
	m.nsid = DefaultNsid
	m.Header.Type = NLMSG_ERROR
	m.Header.Len = SizeofErrormsg
	return m
}

func (m *ErrorMessage) Close() error {
	runtime.SetFinalizer(m, nil)
	repool(m)
	return nil
}

func (m *ErrorMessage) Nsid() *int { return &m.nsid }

func (m *ErrorMessage) Read(b []byte) (int, error) {
	if len(b) < SizeofErrorMessage {
		return 0, syscall.EOVERFLOW
	}
	n, _ := m.Header.Read(b)
	*(*Errormsg)(unsafe.Pointer(&b[n])) = m.Errormsg
	n += SizeofErrormsg
	return n, nil
}

func (m *ErrorMessage) String() string { return StringOf(m) }

func (m *ErrorMessage) Write(b []byte) (int, error) {
	if len(b) < SizeofErrorMessage {
		return 0, syscall.EOVERFLOW
	}
	n, _ := m.Header.Write(b)
	m.Errormsg = *(*Errormsg)(unsafe.Pointer(&b[n]))
	n += SizeofErrormsg
	return n, nil
}

func (m *ErrorMessage) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, m.Header.Type, ":\n")
	indent.Increase(acc)
	if m.nsid != DefaultNsid {
		fmt.Fprintln(acc, "nsid:", m.nsid)
	}
	m.Header.WriteTo(acc)
	fmt.Fprintln(acc, "error:", syscall.Errno(-m.Errno))
	fmt.Fprintln(acc, "req:", m.Req.Type)
	indent.Increase(acc)
	m.Req.WriteTo(acc)
	indent.Decrease(acc)
	indent.Decrease(acc)
	return acc.Tuple()
}

type IfInfoMessage struct {
	nsid int
	Header
	IfInfomsg
	Attrs [IFLA_MAX]Attr
}

const SizeofIfInfoMessage = SizeofHeader + SizeofIfInfomsg

type IfInfomsg struct {
	Family   uint8
	_        uint8
	L2IfType uint16
	Index    uint32
	Flags    IfInfoFlags
	Change   IfInfoFlags
}

const SizeofIfInfomsg = 16

func NewIfInfoMessage() *IfInfoMessage {
	m := pool.IfInfoMessage.Get().(*IfInfoMessage)
	runtime.SetFinalizer(m, (*IfInfoMessage).Close)
	m.nsid = DefaultNsid
	return m
}

func (m *IfInfoMessage) Close() error {
	runtime.SetFinalizer(m, nil)
	closeAttrs(m.Attrs[:])
	repool(m)
	return nil
}

func (m *IfInfoMessage) Nsid() *int { return &m.nsid }

func (m *IfInfoMessage) Read(b []byte) (int, error) {
	if len(b) < SizeofIfInfoMessage {
		return 0, syscall.EOVERFLOW
	}
	n, _ := m.Header.Read(b)
	*(*IfInfomsg)(unsafe.Pointer(&b[n])) = m.IfInfomsg
	n += SizeofIfInfomsg
	n += AttrVec(m.Attrs[:]).Set(b[n:])
	return n, nil
}

func (m *IfInfoMessage) String() string { return StringOf(m) }

func (m *IfInfoMessage) Write(b []byte) (int, error) {
	if len(b) < SizeofIfInfoMessage {
		return 0, syscall.EOVERFLOW
	}
	n, _ := m.Header.Write(b)
	m.IfInfomsg = *(*IfInfomsg)(unsafe.Pointer(&b[n]))
	n += SizeofIfInfomsg
	for n < len(b) {
		a, v, next := nextAttr(b, n)
		n = next
		k := IfInfoAttrKind(a.Kind())
		switch k {
		case IFLA_IFNAME, IFLA_QDISC:
			m.Attrs[k] = StringAttrBytes(v[:len(v)-1])
		case IFLA_MTU, IFLA_LINK, IFLA_MASTER,
			IFLA_WEIGHT,
			IFLA_NET_NS_PID, IFLA_NET_NS_FD, IFLA_LINK_NETNSID,
			IFLA_EXT_MASK, IFLA_PROMISCUITY,
			IFLA_NUM_TX_QUEUES, IFLA_NUM_RX_QUEUES, IFLA_TXQLEN,
			IFLA_GSO_MAX_SEGS, IFLA_GSO_MAX_SIZE,
			IFLA_CARRIER_CHANGES,
			IFLA_GROUP:
			m.Attrs[k] = Uint32AttrBytes(v)
		case IFLA_CARRIER, IFLA_LINKMODE, IFLA_PROTO_DOWN:
			m.Attrs[k] = Uint8Attr(v[0])
		case IFLA_OPERSTATE:
			m.Attrs[k] = IfOperState(v[0])
		case IFLA_STATS:
			m.Attrs[k] = NewLinkStatsBytes(v)
		case IFLA_STATS64:
			m.Attrs[k] = NewLinkStats64Bytes(v)
		case IFLA_AF_SPEC:
			m.Attrs[k] = parse_af_spec(v)
		case IFLA_ADDRESS, IFLA_BROADCAST:
			m.Attrs[k] = afAddr(AF_UNSPEC, v)
		case IFLA_LINKINFO:
			m.Attrs[k] = parse_link_info(v)
		case IFLA_MAP:
		default:
			if k < IFLA_MAX {
				m.Attrs[k] = NewHexStringAttrBytes(v)
			} else {
				return n, fmt.Errorf("%#v: unknown attr", k)
			}
		}
	}
	return n, nil
}

func (m *IfInfoMessage) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, MessageType(m.Header.Type), ":\n")
	indent.Increase(acc)
	if m.nsid != DefaultNsid {
		fmt.Fprintln(acc, "nsid:", m.nsid)
	}
	m.Header.WriteTo(acc)
	fmt.Fprintln(acc, "index:", m.Index)
	fmt.Fprintln(acc, "family:", AddressFamily(m.Family))
	fmt.Fprintln(acc, "type:", L2IfType(m.L2IfType))
	fmt.Fprintln(acc, "ifinfo flags:", m.IfInfomsg.Flags)
	if m.Change != 0 {
		if m.Change == ^IfInfoFlags(0) { // means everything changed
			fmt.Fprintln(acc, "changed flags: everything")
		} else {
			fmt.Fprintln(acc, "changed flags:", m.Change)
		}
	}
	fprintAttrs(acc, ifInfoAttrKindNames, m.Attrs[:])
	indent.Decrease(acc)
	return acc.Tuple()
}

type IfAddrMessage struct {
	nsid int
	Header
	IfAddrmsg
	Attrs [IFA_MAX]Attr
}

const SizeofIfAddrMessage = SizeofHeader + SizeofIfAddrmsg

type IfAddrmsg struct {
	Family    AddressFamily
	Prefixlen uint8
	Flags     uint8
	Scope     uint8
	Index     uint32
}

const SizeofIfAddrmsg = 8

func NewIfAddrMessage() *IfAddrMessage {
	m := pool.IfAddrMessage.Get().(*IfAddrMessage)
	runtime.SetFinalizer(m, (*IfAddrMessage).Close)
	m.nsid = DefaultNsid
	return m
}

func (m *IfAddrMessage) Close() error {
	runtime.SetFinalizer(m, nil)
	closeAttrs(m.Attrs[:])
	repool(m)
	return nil
}

func (m *IfAddrMessage) Nsid() *int { return &m.nsid }

func (m *IfAddrMessage) Read(b []byte) (int, error) {
	if len(b) < SizeofIfAddrMessage {
		return 0, syscall.EOVERFLOW
	}
	n, _ := m.Header.Read(b)
	*(*IfAddrmsg)(unsafe.Pointer(&b[n])) = m.IfAddrmsg
	n += SizeofIfAddrmsg
	n += AttrVec(m.Attrs[:]).Set(b[n:])
	return n, nil
}

func (m *IfAddrMessage) String() string { return StringOf(m) }

func (m *IfAddrMessage) Write(b []byte) (int, error) {
	if len(b) < SizeofIfAddrMessage {
		return 0, syscall.EOVERFLOW
	}
	n, _ := m.Header.Write(b)
	m.IfAddrmsg = *(*IfAddrmsg)(unsafe.Pointer(&b[n]))
	n += SizeofIfAddrmsg
	for n < len(b) {
		a, v, next := nextAttr(b, n)
		n = next
		k := IfAddrAttrKind(a.Kind())
		switch k {
		case IFA_LABEL:
			m.Attrs[k] = StringAttrBytes(v[:len(v)-1])
		case IFA_FLAGS:
			m.Attrs[k] = IfAddrFlagAttrBytes(v)
		case IFA_CACHEINFO:
			m.Attrs[k] = NewIfAddrCacheInfoBytes(v)
		case IFA_ADDRESS, IFA_LOCAL, IFA_BROADCAST, IFA_ANYCAST,
			IFA_MULTICAST:
			m.Attrs[k] = afAddr(AddressFamily(m.Family), v)
		default:
			if k < IFA_MAX {
				m.Attrs[k] = NewHexStringAttrBytes(v)
			} else {
				return n, fmt.Errorf("%#v: unknown attr", k)
			}
		}
	}
	return n, nil
}

func (m *IfAddrMessage) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, MessageType(m.Header.Type), ":\n")
	indent.Increase(acc)
	if m.nsid != DefaultNsid {
		fmt.Fprintln(acc, "nsid:", m.nsid)
	}
	m.Header.WriteTo(acc)
	fmt.Fprintln(acc, "index:", m.Index)
	fmt.Fprintln(acc, "family:", AddressFamily(m.Family))
	fmt.Fprintln(acc, "prefix:", m.Prefixlen)
	fmt.Fprintln(acc, "ifaddr flags:", m.IfAddrmsg.Flags)
	fmt.Fprintln(acc, "scope:", RtScope(m.Scope))
	fprintAttrs(acc, ifAddrAttrKindNames, m.Attrs[:])
	indent.Decrease(acc)
	return acc.Tuple()
}

type RouteMessage struct {
	nsid int
	Header
	Rtmsg
	Attrs [RTA_MAX]Attr
}

const SizeofRouteMessage = SizeofHeader + SizeofRtmsg

type Rtmsg struct {
	Family     AddressFamily
	DstLen     uint8
	SrcLen     uint8
	Tos        uint8
	Table      RouteTableKind
	Protocol   RouteProtocol
	Scope      RtScope
	RouteType  RouteType
	RouteFlags RouteFlags
}

const SizeofRtmsg = 12

func NewRouteMessage() *RouteMessage {
	m := pool.RouteMessage.Get().(*RouteMessage)
	runtime.SetFinalizer(m, (*RouteMessage).Close)
	m.nsid = DefaultNsid
	return m
}

func (m *RouteMessage) Close() error {
	runtime.SetFinalizer(m, nil)
	closeAttrs(m.Attrs[:])
	repool(m)
	return nil
}

func (m *RouteMessage) Nsid() *int { return &m.nsid }

func (m *RouteMessage) Read(b []byte) (int, error) {
	if len(b) < SizeofRouteMessage {
		return 0, syscall.EOVERFLOW
	}
	n, _ := m.Header.Read(b)
	*(*Rtmsg)(unsafe.Pointer(&b[n])) = m.Rtmsg
	n += SizeofRtmsg
	n += AttrVec(m.Attrs[:]).Set(b[n:])
	return n, nil
}

func (m *RouteMessage) String() string { return StringOf(m) }

func (m *RouteMessage) Write(b []byte) (int, error) {
	if len(b) < SizeofRouteMessage {
		return 0, syscall.EOVERFLOW
	}
	n, _ := m.Header.Write(b)
	m.Rtmsg = *(*Rtmsg)(unsafe.Pointer(&b[n]))
	n += SizeofRtmsg
	for n < len(b) {
		a, v, next := nextAttr(b, n)
		n = next
		k := RouteAttrKind(a.Kind())
		switch k {
		case RTA_DST, RTA_SRC, RTA_PREFSRC, RTA_GATEWAY:
			m.Attrs[k] = afAddr(AddressFamily(m.Family), v)
		case RTA_TABLE, RTA_IIF, RTA_OIF, RTA_PRIORITY, RTA_FLOW:
			m.Attrs[k] = Uint32AttrBytes(v)
		case RTA_ENCAP_TYPE:
			m.Attrs[k] = LwtunnelEncapType(v[0])
		case RTA_ENCAP:
			m.Attrs[k] = StringAttrBytes(v[:])
		case RTA_PREF:
			m.Attrs[k] = Uint8Attr(v[0])
		case RTA_CACHEINFO:
			m.Attrs[k] = NewRtaCacheInfoBytes(v)
		default:
			if k < RTA_MAX {
				m.Attrs[k] = NewHexStringAttrBytes(v)
			} else {
				return n, fmt.Errorf("%#v: unknown attr", k)
			}
		}
	}
	if a := m.Attrs[RTA_ENCAP_TYPE]; a != nil {
		switch a.(LwtunnelEncapType) {
		case LWTUNNEL_ENCAP_IP:
			m.Attrs[RTA_ENCAP] = parse_lwtunnel_ip4_encap([]byte(m.Attrs[RTA_ENCAP].(StringAttr)))
		case LWTUNNEL_ENCAP_IP6:
			m.Attrs[RTA_ENCAP] = parse_lwtunnel_ip6_encap([]byte(m.Attrs[RTA_ENCAP].(StringAttr)))
		default:
			panic("not yet")
		}
	}
	return n, nil
}

func (m *RouteMessage) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, MessageType(m.Header.Type), ":\n")
	indent.Increase(acc)
	if m.nsid != DefaultNsid {
		fmt.Fprintln(acc, "nsid:", m.nsid)
	}
	m.Header.WriteTo(acc)
	fmt.Fprintln(acc, "family:", AddressFamily(m.Family))
	fmt.Fprintln(acc, "srclen:", m.SrcLen)
	fmt.Fprintln(acc, "dstlen:", m.DstLen)
	fmt.Fprintln(acc, "tos:", m.Tos)
	fmt.Fprintln(acc, "table:", m.Table)
	fmt.Fprintln(acc, "protocol:", m.Protocol)
	fmt.Fprintln(acc, "scope:", m.Scope)
	fmt.Fprintln(acc, "type:", m.RouteType)
	if m.RouteFlags != 0 {
		fmt.Fprintln(acc, "route flags:", m.RouteFlags)
	}
	fprintAttrs(acc, routeAttrKindNames, m.Attrs[:])
	indent.Decrease(acc)
	return acc.Tuple()
}

type NeighborMessage struct {
	nsid int
	Header
	Ndmsg
	Attrs [NDA_MAX]Attr
}

const SizeofNeighborMessage = SizeofHeader + SizeofNdmsg

type Ndmsg struct {
	Family AddressFamily
	_      [3]uint8
	Index  uint32
	State  NeighborState
	Flags  uint8
	Type   RouteType
}

const SizeofNdmsg = 12

func NewNeighborMessage() *NeighborMessage {
	m := pool.NeighborMessage.Get().(*NeighborMessage)
	runtime.SetFinalizer(m, (*NeighborMessage).Close)
	m.nsid = DefaultNsid
	return m
}

func (m *NeighborMessage) AttrBytes(kind NeighborAttrKind) []byte {
	return m.Attrs[kind].(Byter).Bytes()
}

func (m *NeighborMessage) Close() error {
	runtime.SetFinalizer(m, nil)
	closeAttrs(m.Attrs[:])
	repool(m)
	return nil
}

func (m *NeighborMessage) Nsid() *int { return &m.nsid }

func (m *NeighborMessage) Read(b []byte) (int, error) {
	if len(b) < SizeofNeighborMessage {
		return 0, syscall.EOVERFLOW
	}
	n, _ := m.Header.Read(b)
	*(*Ndmsg)(unsafe.Pointer(&b[n])) = m.Ndmsg
	n += SizeofNdmsg
	n += AttrVec(m.Attrs[:]).Set(b[n:])
	return n, nil
}

func (m *NeighborMessage) String() string { return StringOf(m) }

func (m *NeighborMessage) Write(b []byte) (int, error) {
	if len(b) < SizeofNeighborMessage {
		return 0, syscall.EOVERFLOW
	}
	n, _ := m.Header.Write(b)
	m.Ndmsg = *(*Ndmsg)(unsafe.Pointer(&b[n]))
	n += SizeofNdmsg
	for n < len(b) {
		a, v, next := nextAttr(b, n)
		n = next
		k := NeighborAttrKind(a.Kind())
		switch k {
		case NDA_DST:
			m.Attrs[k] = afAddr(AddressFamily(m.Family), v)
		case NDA_LLADDR:
			m.Attrs[k] = afAddr(AF_UNSPEC, v)
		case NDA_CACHEINFO:
			m.Attrs[k] = NewNdaCacheInfoBytes(v)
		case NDA_PROBES, NDA_VNI, NDA_IFINDEX, NDA_MASTER,
			NDA_LINK_NETNSID:
			m.Attrs[k] = Uint32AttrBytes(v)
		case NDA_VLAN:
			m.Attrs[k] = Uint16AttrBytes(v)
		default:
			if k < NDA_MAX {
				m.Attrs[k] = NewHexStringAttrBytes(v)
			} else {
				return 0, fmt.Errorf("%#v: unknown attr", k)
			}
		}
	}
	return n, nil
}

func (m *NeighborMessage) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, MessageType(m.Header.Type), ":\n")
	indent.Increase(acc)
	if m.nsid != DefaultNsid {
		fmt.Fprintln(acc, "nsid:", m.nsid)
	}
	m.Header.WriteTo(acc)
	fmt.Fprintln(acc, "index:", m.Index)
	fmt.Fprintln(acc, "address family:", AddressFamily(m.Family))
	fmt.Fprintln(acc, "type:", RouteType(m.Ndmsg.Type))
	fmt.Fprintln(acc, "state:", NeighborState(m.State))
	if m.Ndmsg.Flags != 0 {
		fmt.Fprintln(acc, "neighbor flags:", m.Ndmsg.Flags)
	}
	fprintAttrs(acc, neighborAttrKindNames, m.Attrs[:])
	indent.Decrease(acc)
	return acc.Tuple()
}

type NetnsMessage struct {
	nsid int
	Header
	Netnsmsg
	Attrs [NETNSA_MAX]Attr
}

const SizeofNetnsMessage = SizeofHeader + SizeofNetnsmsg

type Netnsmsg struct {
	AddressFamily
	_ [NetnsPad]byte
}

const SizeofNetnsmsg = SizeofAddressFamily + NetnsPad
const NetnsPad = 3

func NewNetnsMessage() *NetnsMessage {
	m := pool.NetnsMessage.Get().(*NetnsMessage)
	runtime.SetFinalizer(m, (*NetnsMessage).Close)
	m.nsid = DefaultNsid
	return m
}

func (m *NetnsMessage) NSID() int32 {
	return m.Attrs[NETNSA_NSID].(Int32Attr).Int()
}

func (m *NetnsMessage) PID() uint32 {
	return m.Attrs[NETNSA_PID].(Uint32Attr).Uint()
}

func (m *NetnsMessage) FD() uint32 {
	return m.Attrs[NETNSA_FD].(Uint32Attr).Uint()
}

func (m *NetnsMessage) AttrBytes(kind NetnsAttrKind) []byte {
	return m.Attrs[kind].(Byter).Bytes()
}

func (m *NetnsMessage) Close() error {
	runtime.SetFinalizer(m, nil)
	closeAttrs(m.Attrs[:])
	repool(m)
	return nil
}

func (m *NetnsMessage) Nsid() *int { return &m.nsid }

func (m *NetnsMessage) Read(b []byte) (int, error) {
	if len(b) < SizeofNetnsMessage {
		return 0, syscall.EOVERFLOW
	}
	n, _ := m.Header.Read(b)
	*(*Netnsmsg)(unsafe.Pointer(&b[n])) = m.Netnsmsg
	n += SizeofNetnsmsg
	n += AttrVec(m.Attrs[:]).Set(b[n:])
	return n, nil
}

func (m *NetnsMessage) String() string { return StringOf(m) }

func (m *NetnsMessage) Write(b []byte) (int, error) {
	if len(b) < SizeofNetnsMessage {
		return 0, syscall.EOVERFLOW
	}
	n, _ := m.Header.Write(b)
	m.Netnsmsg = *(*Netnsmsg)(unsafe.Pointer(&b[n]))
	n += SizeofNetnsmsg
	for n < len(b) {
		a, v, next := nextAttr(b, n)
		n = next
		k := NetnsAttrKind(a.Kind())
		switch k {
		case NETNSA_NONE:
		case NETNSA_NSID:
			m.Attrs[k] = Int32AttrBytes(v)
		case NETNSA_PID, NETNSA_FD:
			m.Attrs[k] = Uint32AttrBytes(v)
		default:
			return n, fmt.Errorf("%#v: unknown attr", k)
		}
	}
	return n, nil
}

func (m *NetnsMessage) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, m.Header.Type, ":\n")
	indent.Increase(acc)
	if m.nsid != DefaultNsid {
		fmt.Fprintln(acc, "nsid:", m.nsid)
	}
	m.Header.WriteTo(acc)
	fmt.Fprintln(acc, "family:", m.AddressFamily)
	fprintAttrs(acc, netnsAttrKindNames, m.Attrs[:])
	indent.Decrease(acc)
	return acc.Tuple()
}
