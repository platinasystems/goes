// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package netlink

import (
	"bytes"
	"fmt"
	"io"
	"runtime"
	"syscall"

	"unsafe"

	"github.com/platinasystems/go/accumulate"
	"github.com/platinasystems/go/indent"
)

type Byter interface {
	Bytes() []byte
}

type Message interface {
	netlinkMessage()
	MsgType() MsgType
	io.Closer
	Parse([]byte)
	fmt.Stringer
	TxAdd(*Socket)
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

const SizeofHeader = 16

func (h *Header) String() string {
	return StringOf(h)
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
	return acc.N, acc.Err
}
func (h *Header) MsgType() MsgType { return h.Type }

// AFMessage is a generic message depending only on address family.
type GenMessage struct {
	Header
	AddressFamily
}

const SizeofGenMessage = SizeofHeader + SizeofGenmsg
const SizeofGenmsg = SizeofAddressFamily

func NewGenMessage() *GenMessage {
	m := pool.GenMessage.Get().(*GenMessage)
	runtime.SetFinalizer(m, (*GenMessage).Close)
	return m
}

func NewGenMessageBytes(b []byte) *GenMessage {
	m := NewGenMessage()
	m.Parse(b)
	return m
}

func (m *GenMessage) netlinkMessage() {}
func (m *GenMessage) Close() error {
	runtime.SetFinalizer(m, nil)
	repool(m)
	return nil
}
func (m *GenMessage) Parse(b []byte) {
	p := (*GenMessage)(unsafe.Pointer(&b[0]))
	m.Header = p.Header
	m.AddressFamily = p.AddressFamily
}
func (m *GenMessage) String() string {
	return StringOf(m)
}
func (m *GenMessage) TxAdd(s *Socket) {
	b := s.TxAddReq(&m.Header, SizeofGenmsg)
	p := (*GenMessage)(unsafe.Pointer(&b[0]))
	p.AddressFamily = m.AddressFamily
}
func (m *GenMessage) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, m.Header.Type, ":\n")
	indent.Increase(acc)
	m.Header.WriteTo(acc)
	fmt.Fprintln(acc, "family:", m.AddressFamily)
	indent.Decrease(acc)
	return acc.N, acc.Err
}

type NoopMessage struct {
	Header
}

const SizeofNoopMessage = SizeofHeader

func NewNoopMessage() *NoopMessage {
	m := pool.NoopMessage.Get().(*NoopMessage)
	runtime.SetFinalizer(m, (*NoopMessage).Close)
	return m
}

func NewNoopMessageBytes(b []byte) *NoopMessage {
	m := NewNoopMessage()
	m.Parse(b)
	return m
}

func (m *NoopMessage) netlinkMessage() {}
func (m *NoopMessage) Close() error {
	runtime.SetFinalizer(m, nil)
	repool(m)
	return nil
}
func (m *NoopMessage) Parse(b []byte) {
	*m = *(*NoopMessage)(unsafe.Pointer(&b[0]))
}
func (m *NoopMessage) String() string {
	return StringOf(m)
}
func (m *NoopMessage) TxAdd(s *Socket) {
	defer m.Close()
	m.Header.Type = NLMSG_NOOP
	s.TxAddReq(&m.Header, 0)
}
func (m *NoopMessage) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, MessageType(m.Header.Type), ":\n")
	indent.Increase(acc)
	m.Header.WriteTo(acc)
	indent.Decrease(acc)
	return acc.N, acc.Err
}

type DoneMessage struct {
	Header
}

const SizeofDoneMessage = SizeofHeader

func NewDoneMessage() *DoneMessage {
	m := pool.DoneMessage.Get().(*DoneMessage)
	runtime.SetFinalizer(m, (*DoneMessage).Close)
	return m
}

func NewDoneMessageBytes(b []byte) *DoneMessage {
	m := NewDoneMessage()
	m.Parse(b)
	return m
}

func (m *DoneMessage) netlinkMessage() {}
func (m *DoneMessage) Close() error {
	runtime.SetFinalizer(m, nil)
	repool(m)
	return nil
}
func (m *DoneMessage) String() string {
	return StringOf(m)
}
func (m *DoneMessage) Parse(b []byte) {
	*m = *(*DoneMessage)(unsafe.Pointer(&b[0]))
}
func (m *DoneMessage) TxAdd(s *Socket) {
	defer m.Close()
	m.Header.Type = NLMSG_NOOP
	s.TxAddReq(&m.Header, 0)
}
func (m *DoneMessage) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, MessageType(m.Header.Type), ":\n")
	indent.Increase(acc)
	m.Header.WriteTo(acc)
	indent.Decrease(acc)
	return acc.N, acc.Err
}

type ErrorMessage struct {
	Header
	// Unix errno for error.
	Errno int32
	// Header for message with error.
	Req Header
}

const SizeofErrorMessage = SizeofHeader + 4 + SizeofHeader

func NewErrorMessage() *ErrorMessage {
	m := pool.ErrorMessage.Get().(*ErrorMessage)
	runtime.SetFinalizer(m, (*ErrorMessage).Close)
	return m
}

func NewErrorMessageBytes(b []byte) *ErrorMessage {
	m := NewErrorMessage()
	m.Parse(b)
	return m
}

func (m *ErrorMessage) netlinkMessage() {}
func (m *ErrorMessage) Close() error {
	runtime.SetFinalizer(m, nil)
	repool(m)
	return nil
}
func (m *ErrorMessage) Parse(b []byte) {
	*m = *(*ErrorMessage)(unsafe.Pointer(&b[0]))
}
func (m *ErrorMessage) String() string {
	return StringOf(m)
}
func (m *ErrorMessage) TxAdd(s *Socket) {
	defer m.Close()
	m.Header.Type = NLMSG_ERROR
	b := s.TxAddReq(&m.Header, 4+SizeofHeader)
	e := (*ErrorMessage)(unsafe.Pointer(&b[0]))
	e.Errno = m.Errno
	e.Req = m.Req
}
func (m *ErrorMessage) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, m.Header.Type, ":\n")
	indent.Increase(acc)
	m.Header.WriteTo(acc)
	fmt.Fprintln(acc, "error:", syscall.Errno(-m.Errno))
	fmt.Fprintln(acc, "req:", m.Req.Type)
	indent.Increase(acc)
	m.Req.WriteTo(acc)
	indent.Decrease(acc)
	indent.Decrease(acc)
	return acc.N, acc.Err
}

type IfInfoMessage struct {
	Header
	IfInfomsg
	Attrs [IFLA_MAX]Attr
}

const SizeofIfInfoMessage = SizeofHeader + SizeofIfInfomsg

type IfInfomsg struct {
	Family uint8
	_      uint8
	Type   uint16
	Index  uint32
	Flags  IfInfoFlags
	Change IfInfoFlags
}

const SizeofIfInfomsg = 16

func NewIfInfoMessage() *IfInfoMessage {
	m := pool.IfInfoMessage.Get().(*IfInfoMessage)
	runtime.SetFinalizer(m, (*IfInfoMessage).Close)
	return m
}

func NewIfInfoMessageBytes(b []byte) *IfInfoMessage {
	m := NewIfInfoMessage()
	m.Parse(b)
	return m
}

func (m *IfInfoMessage) netlinkMessage() {}

func (m *IfInfoMessage) Close() error {
	runtime.SetFinalizer(m, nil)
	closeAttrs(m.Attrs[:])
	repool(m)
	return nil
}

func (m *IfInfoMessage) Parse(b []byte) {
	p := (*IfInfoMessage)(unsafe.Pointer(&b[0]))
	m.Header = p.Header
	m.IfInfomsg = p.IfInfomsg
	b = b[SizeofIfInfoMessage:]
	for i := 0; i < len(b); {
		n, v, next_i := nextAttr(b, i)
		i = next_i
		switch t := IfInfoAttrKind(n.Kind); t {
		case IFLA_IFNAME, IFLA_QDISC:
			m.Attrs[n.Kind] = StringAttrBytes(v[:len(v)-1])
		case IFLA_MTU, IFLA_LINK, IFLA_MASTER,
			IFLA_WEIGHT,
			IFLA_NET_NS_PID, IFLA_NET_NS_FD, IFLA_LINK_NETNSID,
			IFLA_EXT_MASK, IFLA_PROMISCUITY,
			IFLA_NUM_TX_QUEUES, IFLA_NUM_RX_QUEUES, IFLA_TXQLEN,
			IFLA_GSO_MAX_SEGS, IFLA_GSO_MAX_SIZE,
			IFLA_CARRIER_CHANGES,
			IFLA_GROUP:
			m.Attrs[n.Kind] = Uint32AttrBytes(v)
		case IFLA_CARRIER, IFLA_LINKMODE, IFLA_PROTO_DOWN:
			m.Attrs[n.Kind] = Uint8Attr(v[0])
		case IFLA_OPERSTATE:
			m.Attrs[n.Kind] = IfOperState(v[0])
		case IFLA_STATS:
			m.Attrs[n.Kind] = NewLinkStatsBytes(v)
		case IFLA_STATS64:
			m.Attrs[n.Kind] = NewLinkStats64Bytes(v)
		case IFLA_AF_SPEC:
			m.Attrs[n.Kind] = parse_af_spec(v)
		case IFLA_ADDRESS, IFLA_BROADCAST:
			m.Attrs[n.Kind] = afAddr(AF_UNSPEC, v)
		case IFLA_MAP:
		default:
			if t < IFLA_MAX {
				m.Attrs[n.Kind] = NewHexStringAttrBytes(v)
			} else {
				panic(fmt.Errorf("%#v: unknown attr", n.Kind))
			}
		}
	}
}

func (m *IfInfoMessage) String() string {
	return StringOf(m)
}

func (m *IfInfoMessage) TxAdd(s *Socket) {
	defer m.Close()
	as := AttrVec(m.Attrs[:])
	b := s.TxAddReq(&m.Header, SizeofIfInfomsg+as.Size())
	i := (*IfInfoMessage)(unsafe.Pointer(&b[0]))
	i.IfInfomsg = m.IfInfomsg
	as.Set(b[SizeofIfInfoMessage:])
}

func (m *IfInfoMessage) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, MessageType(m.Header.Type), ":\n")
	indent.Increase(acc)
	m.Header.WriteTo(acc)
	fmt.Fprintln(acc, "index:", m.Index)
	fmt.Fprintln(acc, "family:", AddressFamily(m.Family))
	fmt.Fprintln(acc, "type:", IfInfoAttrKind(m.Header.Type))
	fmt.Fprintln(acc, "ifinfo flags:", m.IfInfomsg.Flags)
	if m.Change != 0 {
		fmt.Fprintln(acc, "changed flags:", IfInfoFlags(m.Change))
	}
	fprintAttrs(acc, ifInfoAttrKindNames, m.Attrs[:])
	indent.Decrease(acc)
	return acc.N, acc.Err
}

type IfAddrMessage struct {
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
	return m
}

func NewIfAddrMessageBytes(b []byte) *IfAddrMessage {
	m := NewIfAddrMessage()
	m.Parse(b)
	return m
}

func (m *IfAddrMessage) netlinkMessage() {}

func (m *IfAddrMessage) Close() error {
	runtime.SetFinalizer(m, nil)
	closeAttrs(m.Attrs[:])
	repool(m)
	return nil
}

func (m *IfAddrMessage) Parse(b []byte) {
	p := (*IfAddrMessage)(unsafe.Pointer(&b[0]))
	m.Header = p.Header
	m.IfAddrmsg = p.IfAddrmsg
	b = b[SizeofIfAddrMessage:]
	for i := 0; i < len(b); {
		n, v, next_i := nextAttr(b, i)
		i = next_i
		k := IfAddrAttrKind(n.Kind)
		switch k {
		case IFA_LABEL:
			m.Attrs[n.Kind] = StringAttrBytes(v[:len(v)-1])
		case IFA_FLAGS:
			m.Attrs[n.Kind] = IfAddrFlagAttrBytes(v)
		case IFA_CACHEINFO:
			m.Attrs[n.Kind] = NewIfAddrCacheInfoBytes(v)
		case IFA_ADDRESS, IFA_LOCAL, IFA_BROADCAST, IFA_ANYCAST,
			IFA_MULTICAST:
			m.Attrs[n.Kind] = afAddr(AddressFamily(m.Family), v)
		default:
			if k < IFA_MAX {
				m.Attrs[n.Kind] = NewHexStringAttrBytes(v)
			} else {
				panic(fmt.Errorf("%#v: unknown attr", k))
			}
		}
	}
	return
}

func (m *IfAddrMessage) String() string {
	return StringOf(m)
}

func (m *IfAddrMessage) TxAdd(s *Socket) {
	defer m.Close()
	as := AttrVec(m.Attrs[:])
	b := s.TxAddReq(&m.Header, SizeofIfAddrmsg+as.Size())
	i := (*IfAddrMessage)(unsafe.Pointer(&b[0]))
	i.IfAddrmsg = m.IfAddrmsg
	as.Set(b[SizeofIfAddrMessage:])
}

func (m *IfAddrMessage) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, MessageType(m.Header.Type), ":\n")
	indent.Increase(acc)
	m.Header.WriteTo(acc)
	fmt.Fprintln(acc, "index:", m.Index)
	fmt.Fprintln(acc, "family:", AddressFamily(m.Family))
	fmt.Fprintln(acc, "prefix:", m.Prefixlen)
	fmt.Fprintln(acc, "ifaddr flags:", m.IfAddrmsg.Flags)
	fmt.Fprintln(acc, "scope:", RtScope(m.Scope))
	fprintAttrs(acc, ifAddrAttrKindNames, m.Attrs[:])
	indent.Decrease(acc)
	return acc.N, acc.Err
}

type RouteMessage struct {
	Header
	Rtmsg
	Attrs [RTA_MAX]Attr
}

const SizeofRouteMessage = SizeofHeader + SizeofRtmsg

type Rtmsg struct {
	Family   AddressFamily
	DstLen   uint8
	SrcLen   uint8
	Tos      uint8
	Table    uint8
	Protocol RouteProtocol
	Scope    RtScope
	Type     RouteType
	Flags    RouteFlags
}

const SizeofRtmsg = 12

func NewRouteMessage() *RouteMessage {
	m := pool.RouteMessage.Get().(*RouteMessage)
	runtime.SetFinalizer(m, (*RouteMessage).Close)
	return m
}

func NewRouteMessageBytes(b []byte) *RouteMessage {
	m := NewRouteMessage()
	m.Parse(b)
	return m
}

func (m *RouteMessage) netlinkMessage() {}

func (m *RouteMessage) Close() error {
	runtime.SetFinalizer(m, nil)
	closeAttrs(m.Attrs[:])
	repool(m)
	return nil
}

func (m *RouteMessage) Parse(b []byte) {
	p := (*RouteMessage)(unsafe.Pointer(&b[0]))
	m.Header = p.Header
	m.Rtmsg = p.Rtmsg
	b = b[SizeofRouteMessage:]
	for i := 0; i < len(b); {
		n, v, next_i := nextAttr(b, i)
		i = next_i
		k := RouteAttrKind(n.Kind)
		switch k {
		case RTA_DST, RTA_SRC, RTA_PREFSRC, RTA_GATEWAY:
			m.Attrs[n.Kind] = afAddr(AddressFamily(m.Family), v)
		case RTA_TABLE, RTA_IIF, RTA_OIF, RTA_PRIORITY, RTA_FLOW:
			m.Attrs[n.Kind] = Uint32AttrBytes(v)
		case RTA_ENCAP_TYPE:
			m.Attrs[n.Kind] = Uint16AttrBytes(v)
		case RTA_PREF:
			m.Attrs[n.Kind] = Uint8Attr(v[0])
		case RTA_CACHEINFO:
			m.Attrs[n.Kind] = NewRtaCacheInfoBytes(v)
		default:
			if k < RTA_MAX {
				m.Attrs[n.Kind] = NewHexStringAttrBytes(v)
			} else {
				panic(fmt.Errorf("%#v: unknown attr", k))
			}
		}
	}
	return
}

func (m *RouteMessage) String() string {
	return StringOf(m)
}

func (m *RouteMessage) TxAdd(s *Socket) {
	defer m.Close()
	as := AttrVec(m.Attrs[:])
	b := s.TxAddReq(&m.Header, SizeofRtmsg+as.Size())
	i := (*RouteMessage)(unsafe.Pointer(&b[0]))
	i.Rtmsg = m.Rtmsg
	as.Set(b[SizeofRouteMessage:])
}

func (m *RouteMessage) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, MessageType(m.Header.Type), ":\n")
	indent.Increase(acc)
	m.Header.WriteTo(acc)
	fmt.Fprintln(acc, "family:", AddressFamily(m.Family))
	fmt.Fprintln(acc, "srclen:", m.SrcLen)
	fmt.Fprintln(acc, "dstlen:", m.DstLen)
	fmt.Fprintln(acc, "tos:", m.Tos)
	fmt.Fprintln(acc, "table:", m.Table)
	fmt.Fprintln(acc, "protocol:", m.Protocol)
	fmt.Fprintln(acc, "scope:", m.Scope)
	fmt.Fprintln(acc, "type:", m.Rtmsg.Type)
	if m.Rtmsg.Flags != 0 {
		fmt.Fprintln(acc, "route flags:", m.Rtmsg.Flags)
	}
	fprintAttrs(acc, routeAttrKindNames, m.Attrs[:])
	indent.Decrease(acc)
	return acc.N, acc.Err
}

type NeighborMessage struct {
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
	return m
}

func NewNeighborMessageBytes(b []byte) *NeighborMessage {
	m := NewNeighborMessage()
	m.Parse(b)
	return m
}

func (m *NeighborMessage) netlinkMessage() {}

func (m *NeighborMessage) AttrBytes(kind NeighborAttrKind) []byte {
	return m.Attrs[kind].(Byter).Bytes()
}

func (m *NeighborMessage) Close() error {
	runtime.SetFinalizer(m, nil)
	closeAttrs(m.Attrs[:])
	repool(m)
	return nil
}

func (m *NeighborMessage) Parse(b []byte) {
	p := (*NeighborMessage)(unsafe.Pointer(&b[0]))
	m.Header = p.Header
	m.Ndmsg = p.Ndmsg
	b = b[SizeofNeighborMessage:]
	for i := 0; i < len(b); {
		n, v, next_i := nextAttr(b, i)
		i = next_i
		k := NeighborAttrKind(n.Kind)
		switch k {
		case NDA_DST:
			m.Attrs[n.Kind] = afAddr(AddressFamily(m.Family), v)
		case NDA_LLADDR:
			m.Attrs[n.Kind] = afAddr(AF_UNSPEC, v)
		case NDA_CACHEINFO:
			m.Attrs[n.Kind] = NewNdaCacheInfoBytes(v)
		case NDA_PROBES, NDA_VNI, NDA_IFINDEX, NDA_MASTER,
			NDA_LINK_NETNSID:
			m.Attrs[n.Kind] = Uint32AttrBytes(v)
		case NDA_VLAN:
			m.Attrs[n.Kind] = Uint16AttrBytes(v)
		default:
			if k < NDA_MAX {
				m.Attrs[n.Kind] = NewHexStringAttrBytes(v)
			} else {
				panic(fmt.Errorf("%#v: unknown attr", k))
			}
		}
	}
	return
}

func (m *NeighborMessage) String() string {
	return StringOf(m)
}

func (m *NeighborMessage) TxAdd(s *Socket) {
	defer m.Close()
	as := AttrVec(m.Attrs[:])
	b := s.TxAddReq(&m.Header, SizeofNdmsg+as.Size())
	i := (*NeighborMessage)(unsafe.Pointer(&b[0]))
	i.Ndmsg = m.Ndmsg
	as.Set(b[SizeofNeighborMessage:])
}

func (m *NeighborMessage) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, MessageType(m.Header.Type), ":\n")
	indent.Increase(acc)
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
	return acc.N, acc.Err
}

type NetnsMessage struct {
	GenMessage
	Attrs [NETNSA_MAX]Attr
}

const NetnsPad = 3
const SizeofNetnsmsg = SizeofGenmsg + NetnsPad
const SizeofNetnsMessage = SizeofGenMessage + NetnsPad

func NewNetnsMessage() *NetnsMessage {
	m := pool.NetnsMessage.Get().(*NetnsMessage)
	runtime.SetFinalizer(m, (*NetnsMessage).Close)
	return m
}

func NewNetnsMessageBytes(b []byte) *NetnsMessage {
	m := NewNetnsMessage()
	m.Parse(b)
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

func (m *NetnsMessage) netlinkMessage() {}

func (m *NetnsMessage) AttrBytes(kind NetnsAttrKind) []byte {
	return m.Attrs[kind].(Byter).Bytes()
}

func (m *NetnsMessage) Close() error {
	runtime.SetFinalizer(m, nil)
	closeAttrs(m.Attrs[:])
	repool(m)
	return nil
}

func (m *NetnsMessage) Parse(b []byte) {
	p := (*NetnsMessage)(unsafe.Pointer(&b[0]))
	m.GenMessage = p.GenMessage
	b = b[SizeofNetnsMessage:]
	m.Attrs[NETNSA_NSID] = Int32Attr(-2)
	m.Attrs[NETNSA_PID] = Uint32Attr(0)
	m.Attrs[NETNSA_FD] = Uint32Attr(^uint32(0))
	for i := 0; i < len(b); {
		n, v, next_i := nextAttr(b, i)
		i = next_i
		k := NetnsAttrKind(n.Kind)
		switch k {
		case NETNSA_NONE:
		case NETNSA_NSID:
			m.Attrs[n.Kind] = Int32AttrBytes(v)
		case NETNSA_PID, NETNSA_FD:
			m.Attrs[n.Kind] = Uint32AttrBytes(v)
		default:
			panic(fmt.Errorf("%#v: unknown attr", k))
		}
	}
	return
}

func (m *NetnsMessage) String() string {
	return StringOf(m)
}

func (m *NetnsMessage) TxAdd(s *Socket) {
	defer m.Close()
	as := AttrVec(m.Attrs[:])
	b := s.TxAddReq(&m.Header, SizeofNetnsmsg+as.Size())
	b[SizeofHeader] = byte(m.AddressFamily)
	as.Set(b[SizeofNetnsMessage:])
}

func (m *NetnsMessage) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	m.GenMessage.WriteTo(acc)
	indent.Increase(acc)
	fprintAttrs(acc, netnsAttrKindNames, m.Attrs[:])
	indent.Decrease(acc)
	return acc.N, acc.Err
}
