// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package netlink

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"unsafe"

	"github.com/platinasystems/go/accumulate"
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/indent"
)

type AttrType interface {
	attrType()
	IthString(int) string
}

type Attr interface {
	attr()
	Set([]byte)
	Size() int
	fmt.Stringer
	io.WriterTo
}

type Int8er interface {
	Int() int8
}
type Int32er interface {
	Uint() int32
}
type Int64er interface {
	Int() int64
}
type Uint8er interface {
	Uint() uint8
}
type Uint32er interface {
	Uint() uint32
}
type Uint64er interface {
	Uint() uint64
}

func closeAttrs(attrs []Attr) {
	for i, a := range attrs {
		if a != nil {
			if method, found := a.(io.Closer); found {
				method.Close()
			}
			attrs[i] = nil
		}
	}
}

func fprintAttrs(w io.Writer, names []string, attrs []Attr) (int64,
	error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	for i, v := range attrs {
		if v == nil {
			continue
		}
		fmt.Fprint(acc, elib.Stringer(names, i), ":")
		if _, found := v.(multiliner); found {
			fmt.Fprintln(acc)
			indent.Increase(acc)
			v.WriteTo(acc)
			indent.Decrease(acc)
		} else {
			fmt.Fprint(acc, " ")
			v.WriteTo(acc)
			fmt.Fprintln(acc)
		}
	}
	return acc.N, acc.Err
}

func nextAttr(b []byte, i int) (n *NlAttr, v []byte, j int) {
	n = (*NlAttr)(unsafe.Pointer(&b[i]))
	v = b[i+SizeofNlAttr : i+int(n.Len)]
	j = i + attrAlignLen(int(n.Len))
	return
}

func parse_af_spec(b []byte) *AttrArray {
	as := pool.AttrArray.Get().(*AttrArray)
	as.Type = NewAddressFamilyAttrType()
	for i := 0; i < len(b); {
		n, v, next_i := nextAttr(b, i)
		i = next_i
		af := AddressFamily(n.Kind)
		as.X.Validate(uint(af))
		switch af {
		case AF_INET:
			as.X[af] = parse_ip4_af_spec(v)
		case AF_INET6:
			as.X[af] = parse_ip6_af_spec(v)
		default:
			panic("unknown address family " + af.String())
		}
	}
	return as
}

func parse_ip4_af_spec(b []byte) *AttrArray {
	as := pool.AttrArray.Get().(*AttrArray)
	as.Type = NewIp4IfAttrType()
	for i := 0; i < len(b); {
		n, v, next_i := nextAttr(b, i)
		i = next_i
		t := Ip4IfAttrKind(n.Kind)
		as.X.Validate(uint(t))
		switch t {
		case IFLA_INET_UNSPEC:
		case IFLA_INET_CONF:
			as.X[t] = NewIp4DevConfBytes(v)
		default:
			as.X[t] = NewHexStringAttrBytes(v)
		}
	}
	return as
}

func parse_ip6_af_spec(b []byte) *AttrArray {
	as := pool.AttrArray.Get().(*AttrArray)
	as.Type = NewIp6IfAttrType()
	for i := 0; i < len(b); {
		n, v, next_i := nextAttr(b, i)
		i = next_i
		t := Ip6IfAttrKind(n.Kind)
		as.X.Validate(uint(t))
		switch t {
		case IFLA_INET6_UNSPEC:
		case IFLA_INET6_FLAGS:
			flags := Ip6IfFlagsAttrBytes(v)
			if flags != 0 {
				as.X[t] = flags
			}
		case IFLA_INET6_CONF:
			as.X[t] = NewIp6DevConfBytes(v)
		case IFLA_INET6_STATS:
		case IFLA_INET6_MCAST:
		case IFLA_INET6_CACHEINFO:
		case IFLA_INET6_ICMP6STATS:
		case IFLA_INET6_TOKEN:
		case IFLA_INET6_ADDR_GEN_MODE:
		default:
			as.X[t] = NewHexStringAttrBytes(v)
		}
	}
	return as
}

//go:generate gentemplate -d Package=netlink -id Attr -d VecType=AttrVec -d Type=Attr github.com/platinasystems/go/elib/vec.tmpl

func (a AttrVec) Size() (l int) {
	for i := range a {
		if a[i] != nil {
			l += SizeofNlAttr + attrAlignLen(a[i].Size())
		}
	}
	return
}

func (a AttrVec) Set(v []byte) {
	vi := 0
	for i := range a {
		if a[i] == nil {
			continue
		}

		s := a[i].Size()

		// Fill in attribute header.
		nla := (*NlAttr)(unsafe.Pointer(&v[vi]))
		nla.Kind = uint16(i)
		nla.Len = uint16(SizeofNlAttr + s)

		// Fill in attribute value.
		a[i].Set(v[vi+SizeofNlAttr : vi+SizeofNlAttr+s])
		vi += SizeofNlAttr + attrAlignLen(s)
	}
}

type AttrArray struct {
	X    AttrVec
	Type AttrType
}

func (a *AttrArray) attr() {}

func (a *AttrArray) multiline() {}

func (a *AttrArray) Close() error {
	for i, x := range a.X {
		if x != nil {
			if method, found := x.(io.Closer); found {
				method.Close()
			}
			a.X[i] = nil
		}
	}
	if method, found := a.Type.(io.Closer); found {
		method.Close()
	}
	repool(a)
	return nil
}
func (a *AttrArray) Set(v []byte) {
	a.X.Set(v)
}
func (a *AttrArray) Size() int {
	return a.X.Size()
}
func (a *AttrArray) String() string {
	return StringOf(a)
}
func (a *AttrArray) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	for i, v := range a.X {
		if v == nil {
			continue
		}
		fmt.Fprint(acc, a.Type.IthString(i), ":")
		if _, found := v.(multiliner); found {
			fmt.Fprintln(acc)
			indent.Increase(acc)
			v.WriteTo(acc)
			indent.Decrease(acc)
		} else {
			fmt.Fprint(acc, " ")
			v.WriteTo(acc)
			fmt.Fprintln(acc)
		}
	}
	return acc.N, acc.Err
}

type EthernetAddress [6]byte

func NewEthernetAddressBytes(b []byte) *EthernetAddress {
	a := pool.EthernetAddress.Get().(*EthernetAddress)
	a.Parse(b)
	return a
}

func (a *EthernetAddress) attr() {}
func (a *EthernetAddress) Bytes() []byte {
	return a[:]
}
func (a *EthernetAddress) Close() error {
	repool(a)
	return nil
}
func (a *EthernetAddress) Parse(b []byte) {
	copy(a[:], b[:6])
}
func (a *EthernetAddress) Set(v []byte) {
	copy(v, a[:])
}
func (a *EthernetAddress) Size() int {
	return len(a)
}
func (a *EthernetAddress) String() string {
	return StringOf(a)
}
func (a *EthernetAddress) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprintf(acc, "%02x:%02x:%02x:%02x:%02x:%02x",
		a[0], a[1], a[2], a[3], a[4], a[5])
	return acc.N, acc.Err
}

type HexStringAttr bytes.Buffer

func NewHexStringAttrBytes(b []byte) *HexStringAttr {
	h := (*HexStringAttr)(pool.Bytes.Get().(*bytes.Buffer))
	h.Parse(b)
	return h
}

func (a *HexStringAttr) attr() {}
func (a *HexStringAttr) Buffer() *bytes.Buffer {
	return (*bytes.Buffer)(a)
}
func (a *HexStringAttr) Close() error {
	repool(a.Buffer())
	return nil
}
func (a *HexStringAttr) Parse(b []byte) {
	a.Buffer().Write(b)
}
func (a *HexStringAttr) Set(v []byte) {
	copy(v, a.Buffer().Bytes())
}
func (a *HexStringAttr) Size() int {
	return a.Buffer().Len()
}
func (a *HexStringAttr) String() string {
	return StringOf(a)
}
func (a *HexStringAttr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, hex.EncodeToString(a.Buffer().Bytes()))
	return acc.N, acc.Err
}

type IfAddrCacheInfo struct {
	Prefered         uint32
	Valid            uint32
	CreatedTimestamp uint32 /* created timestamp, hundredths of seconds */
	UpdatedTimestamp uint32 /* updated timestamp, hundredths of seconds */
}

func NewIfAddrCacheInfoBytes(b []byte) *IfAddrCacheInfo {
	a := pool.IfAddrCacheInfo.Get().(*IfAddrCacheInfo)
	a.Parse(b)
	return a
}

func (a *IfAddrCacheInfo) attr() {}

func (a *IfAddrCacheInfo) multiline() {}

func (a *IfAddrCacheInfo) Close() error {
	repool(a)
	return nil
}
func (a *IfAddrCacheInfo) Set(v []byte) {
	panic("should never be called")
}
func (a *IfAddrCacheInfo) Size() int {
	panic("should never be called")
	return 0
}
func (a *IfAddrCacheInfo) String() string {
	return StringOf(a)
}
func (a *IfAddrCacheInfo) Parse(b []byte) {
	*a = *(*IfAddrCacheInfo)(unsafe.Pointer(&b[0]))
}
func (a *IfAddrCacheInfo) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprintln(acc, "prefered:", a.Prefered)
	fmt.Fprintln(acc, "valid:", a.Valid)
	fmt.Fprintln(acc, "created:", a.CreatedTimestamp)
	fmt.Fprintln(acc, "updated:", a.UpdatedTimestamp)
	return acc.N, acc.Err
}

type IfAddrFlagAttr uint32

func IfAddrFlagAttrBytes(b []byte) IfAddrFlagAttr {
	return *(*IfAddrFlagAttr)(unsafe.Pointer(&b[0]))
}

func (a IfAddrFlagAttr) attr() {}
func (a IfAddrFlagAttr) Size() int {
	return 4
}
func (a IfAddrFlagAttr) Set(v []byte) {
	*(*IfAddrFlagAttr)(unsafe.Pointer(&v[0])) = a
}
func (a IfAddrFlagAttr) String() string {
	return IfAddrFlags(a).String()
}
func (a IfAddrFlagAttr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, IfAddrFlags(a))
	return acc.N, acc.Err
}

type IfOperState uint8

func (a IfOperState) attr() {}
func (a IfOperState) Set(v []byte) {
	panic("should never be called")
}
func (a IfOperState) Size() int {
	return 1
}
func (a IfOperState) String() string {
	if int(a) >= len(ifOperStates) {
		a = 0
	}
	return ifOperStates[a]
}
func (a IfOperState) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a)
	return acc.N, acc.Err
}

type Int8Attr int8

func (a Int8Attr) attr() {}
func (a Int8Attr) Rune() rune {
	return rune(a)
}
func (a Int8Attr) Set(v []byte) {
	v[0] = byte(a)
}
func (a Int8Attr) Size() int {
	return 1
}
func (a Int8Attr) String() string {
	return StringOf(a)
}
func (a Int8Attr) Int() int8 {
	return int8(a)
}
func (a Int8Attr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a.Int())
	return acc.N, acc.Err
}

type Int16Attr int16

func Int16AttrBytes(b []byte) Int16Attr {
	return Int16Attr(*(*int16)(unsafe.Pointer(&b[0])))
}

func (a Int16Attr) attr() {}
func (a Int16Attr) Set(v []byte) {
	*(*Int16Attr)(unsafe.Pointer(&v[0])) = a
}
func (a Int16Attr) Size() int {
	return 2
}
func (a Int16Attr) String() string {
	return StringOf(a)
}
func (a Int16Attr) Int() int16 {
	return int16(a)
}
func (a Int16Attr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a.Int())
	return acc.N, acc.Err
}

type Int32Attr int32

func Int32AttrBytes(b []byte) Int32Attr {
	return Int32Attr(*(*int32)(unsafe.Pointer(&b[0])))
}

func (a Int32Attr) attr() {}
func (a Int32Attr) Set(v []byte) {
	*(*Int32Attr)(unsafe.Pointer(&v[0])) = a
}
func (a Int32Attr) Size() int {
	return 4
}
func (a Int32Attr) String() string {
	return StringOf(a)
}
func (a Int32Attr) Int() int32 {
	return int32(a)
}
func (a Int32Attr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a.Int())
	return acc.N, acc.Err
}

type Int64Attr int64

func Int64AttrBytes(b []byte) Int64Attr {
	return Int64Attr(*(*int64)(unsafe.Pointer(&b[0])))
}

func (a Int64Attr) attr() {}
func (a Int64Attr) Set(v []byte) {
	*(*Int64Attr)(unsafe.Pointer(&v[0])) = a
}
func (a Int64Attr) Size() int {
	return 8
}
func (a Int64Attr) String() string {
	return StringOf(a)
}
func (a Int64Attr) Int() int64 {
	return int64(a)
}
func (a Int64Attr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a.Int())
	return acc.N, acc.Err
}

type Ip4Address [4]byte

func NewIp4AddressBytes(b []byte) *Ip4Address {
	a := pool.Ip4Address.Get().(*Ip4Address)
	a.Parse(b)
	return a
}

func (a *Ip4Address) attr() {}
func (a *Ip4Address) Bytes() []byte {
	return a[:]
}
func (a *Ip4Address) Close() error {
	repool(a)
	return nil
}
func (a *Ip4Address) Parse(b []byte) {
	copy(a[:], b[:4])
}
func (a *Ip4Address) Set(v []byte) {
	copy(v, a[:])
}
func (a *Ip4Address) Size() int {
	return len(a)
}
func (a *Ip4Address) String() string {
	return StringOf(a)
}
func (a *Ip4Address) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprintf(acc, "%d.%d.%d.%d", a[0], a[1], a[2], a[3])
	return acc.N, acc.Err
}

type Ip4DevConf [IPV4_DEVCONF_MAX]uint32

func NewIp4DevConfBytes(b []byte) *Ip4DevConf {
	a := pool.Ip4DevConf.Get().(*Ip4DevConf)
	a.Parse(b)
	return a
}

func (a *Ip4DevConf) attr() {}

func (a *Ip4DevConf) multiline() {}

func (a *Ip4DevConf) Close() error {
	repool(a)
	return nil
}
func (a *Ip4DevConf) Parse(b []byte) {
	*a = *(*Ip4DevConf)(unsafe.Pointer(&b[0]))
}
func (a *Ip4DevConf) Set(v []byte) {
	panic("not implemented")
}
func (a *Ip4DevConf) Size() int {
	panic("not implemented")
	return 0
}
func (a *Ip4DevConf) String() string {
	return StringOf(a)
}
func (a *Ip4DevConf) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	indent.Increase(acc)
	for i, v := range a {
		if v != 0 {
			fmt.Fprint(acc, Ip4DevConfKind(i), ": ", v, "\n")
		}
	}
	indent.Decrease(acc)
	return acc.N, acc.Err
}

type Ip6Address [16]byte

func NewIp6AddressBytes(b []byte) *Ip6Address {
	a := pool.Ip6Address.Get().(*Ip6Address)
	a.Parse(b)
	return a
}

func (a *Ip6Address) attr() {}
func (a *Ip6Address) Bytes() []byte {
	return a[:]
}
func (a *Ip6Address) Close() error {
	repool(a)
	return nil
}
func (a *Ip6Address) Parse(b []byte) {
	copy(a[:], b[:16])
}
func (a *Ip6Address) Set(v []byte) {
	copy(v, a[:])
}
func (a *Ip6Address) Size() int {
	return len(a)
}
func (a *Ip6Address) String() string {
	return StringOf(a)
}
func (a *Ip6Address) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, net.IP(a[:]))
	return acc.N, acc.Err
}

type Ip6DevConf [IPV6_DEVCONF_MAX]uint32

func NewIp6DevConfBytes(b []byte) *Ip6DevConf {
	a := pool.Ip6DevConf.Get().(*Ip6DevConf)
	a.Parse(b)
	return a
}

func (a *Ip6DevConf) attr() {}

func (a *Ip6DevConf) multiline() {}

func (a *Ip6DevConf) Close() error {
	repool(a)
	return nil
}
func (a *Ip6DevConf) Parse(b []byte) {
	*a = *(*Ip6DevConf)(unsafe.Pointer(&b[0]))
}
func (a *Ip6DevConf) Set(v []byte) {
	panic("not implemented")
}
func (a *Ip6DevConf) Size() int {
	panic("not implemented")
	return 0
}
func (a *Ip6DevConf) String() string {
	return StringOf(a)
}
func (a *Ip6DevConf) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	indent.Increase(acc)
	for i, v := range a {
		if v != 0 {
			fmt.Fprint(acc, Ip6DevConfKind(i), ": ", v, "\n")
		}
	}
	indent.Decrease(acc)
	return acc.N, acc.Err
}

type Ip6IfFlagsAttr uint32

func Ip6IfFlagsAttrBytes(b []byte) Ip6IfFlagsAttr {
	return Ip6IfFlagsAttr(*(*uint32)(unsafe.Pointer(&b[0])))
}

func (a Ip6IfFlagsAttr) attr() {}
func (a Ip6IfFlagsAttr) Set(v []byte) {
	*(*Ip6IfFlagsAttr)(unsafe.Pointer(&v[0])) = a
}
func (a Ip6IfFlagsAttr) Size() int {
	return 4
}
func (a Ip6IfFlagsAttr) String() string {
	return StringOf(a)
}
func (a Ip6IfFlagsAttr) Uint() uint32 {
	return uint32(a)
}
func (a Ip6IfFlagsAttr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	for _, match := range []struct {
		bit  Ip6IfFlagsAttr
		name string
	}{
		{INET6_IF_PREFIX_ONLINK, "ONLINK"},
		{INET6_IF_PREFIX_AUTOCONF, "AUTOCONF"},
		{INET6_IF_RA_OTHERCONF, "OTHERCONF"},
		{INET6_IF_RA_MANAGED, "MANAGED"},
		{INET6_IF_RA_RCVD, "RCVD"},
		{INET6_IF_RS_SENT, "SENT"},
		{INET6_IF_READY, "READY"},
	} {
		if a&match.bit == match.bit {
			if acc.N > 0 {
				fmt.Fprint(acc, " | ")
			}
			fmt.Fprint(acc, match.name)
		}
	}
	return acc.N, acc.Err
}

type LinkStats [N_link_stat]uint32

func NewLinkStatsBytes(b []byte) *LinkStats {
	a := pool.LinkStats.Get().(*LinkStats)
	a.Parse(b)
	return a
}

func (a *LinkStats) attr() {}

func (a *LinkStats) multiline() {}

func (a *LinkStats) Close() error {
	repool(a)
	return nil
}
func (a *LinkStats) Parse(b []byte) {
	*a = *(*LinkStats)(unsafe.Pointer(&b[0]))
}
func (a *LinkStats) Set(v []byte) {
	*(*LinkStats)(unsafe.Pointer(&v[0])) = *a
}
func (a *LinkStats) Size() int {
	return int(N_link_stat) * 4
}
func (a *LinkStats) String() string {
	return StringOf(a)
}
func (a *LinkStats) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	for i, v := range a {
		t := LinkStatType(i)
		if v != 0 || t == Rx_packets || t == Tx_packets {
			fmt.Fprint(acc, t, ": ", v, "\n")
		}
	}
	return acc.N, acc.Err
}

type LinkStats64 [N_link_stat]uint64

func NewLinkStats64Bytes(b []byte) *LinkStats64 {
	a := pool.LinkStats64.Get().(*LinkStats64)
	a.Parse(b)
	return a
}

func (a *LinkStats64) attr() {}

func (a *LinkStats64) multiline() {}

func (a *LinkStats64) Close() error {
	repool(a)
	return nil
}
func (a *LinkStats64) Parse(b []byte) {
	*a = *(*LinkStats64)(unsafe.Pointer(&b[0]))
}
func (a *LinkStats64) Set(v []byte) {
	*(*LinkStats64)(unsafe.Pointer(&v[0])) = *a
}
func (a *LinkStats64) Size() int {
	return int(N_link_stat) * 8
}
func (a *LinkStats64) String() string {
	return StringOf(a)
}
func (a *LinkStats64) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	for i, v := range a {
		t := LinkStatType(i)
		if v != 0 || t == Rx_packets || t == Tx_packets {
			fmt.Fprint(acc, t, ": ", v, "\n")
		}
	}
	return acc.N, acc.Err
}

type NdaCacheInfo struct {
	Confirmed uint32
	Used      uint32
	Updated   uint32
	RefCnt    uint32
}

func NewNdaCacheInfoBytes(b []byte) *NdaCacheInfo {
	a := pool.NdaCacheInfo.Get().(*NdaCacheInfo)
	a.Parse(b)
	return a
}

func (a *NdaCacheInfo) attr() {}

func (a *NdaCacheInfo) multiline() {}

func (a *NdaCacheInfo) Close() error {
	repool(a)
	return nil
}
func (a *NdaCacheInfo) Parse(b []byte) {
	*a = *(*NdaCacheInfo)(unsafe.Pointer(&b[0]))
}
func (a *NdaCacheInfo) Set(v []byte) {
	panic("should never be called")
}
func (a *NdaCacheInfo) Size() int {
	panic("should never be called")
	return 0
}
func (a *NdaCacheInfo) String() string {
	return StringOf(a)
}
func (a *NdaCacheInfo) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprintln(acc, "confirmed:", a.Confirmed)
	fmt.Fprintln(acc, "used:", a.Used)
	fmt.Fprintln(acc, "updated:", a.Updated)
	fmt.Fprintln(acc, "refcnt:", a.RefCnt)
	return acc.N, acc.Err
}

type RtaCacheInfo struct {
	ClntRef uint32
	LastUse uint32
	Expires uint32
	Error   uint32
	Used    uint32
}

func NewRtaCacheInfoBytes(b []byte) *RtaCacheInfo {
	a := pool.RtaCacheInfo.Get().(*RtaCacheInfo)
	a.Parse(b)
	return a
}

func (a *RtaCacheInfo) attr() {}

func (a *RtaCacheInfo) multiline() {}

func (a *RtaCacheInfo) Close() error {
	repool(a)
	return nil
}
func (a *RtaCacheInfo) Set(v []byte) {
	panic("should never be called")
}
func (a *RtaCacheInfo) Size() int {
	panic("should never be called")
	return 0
}
func (a *RtaCacheInfo) String() string {
	return StringOf(a)
}
func (a *RtaCacheInfo) Parse(b []byte) {
	*a = *(*RtaCacheInfo)(unsafe.Pointer(&b[0]))
}
func (a *RtaCacheInfo) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprintln(acc, "clntref:", a.ClntRef)
	fmt.Fprintln(acc, "lastuse:", a.LastUse)
	fmt.Fprintln(acc, "expires:", a.Expires)
	fmt.Fprintln(acc, "error:", a.Error)
	fmt.Fprintln(acc, "used:", a.Used)
	return acc.N, acc.Err
}

type StringAttr string

func StringAttrBytes(b []byte) StringAttr {
	return StringAttr(string(b))
}
func (a StringAttr) attr() {}
func (a StringAttr) Size() int {
	return len(a) + 1
}
func (a StringAttr) Set(v []byte) {
	copy(v, a)
	v = append(v, 0)
}
func (a StringAttr) String() string {
	return string(a)
}
func (a StringAttr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a)
	return acc.N, acc.Err
}

type Uint8Attr uint8

func (a Uint8Attr) attr() {}
func (a Uint8Attr) Rune() rune {
	return rune(a)
}
func (a Uint8Attr) Set(v []byte) {
	v[0] = byte(a)
}
func (a Uint8Attr) Size() int {
	return 1
}
func (a Uint8Attr) String() string {
	return StringOf(a)
}
func (a Uint8Attr) Uint() uint8 {
	return uint8(a)
}
func (a Uint8Attr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a.Uint())
	return acc.N, acc.Err
}

type Uint16Attr uint16

func Uint16AttrBytes(b []byte) Uint16Attr {
	return Uint16Attr(*(*uint16)(unsafe.Pointer(&b[0])))
}

func (a Uint16Attr) attr() {}
func (a Uint16Attr) Set(v []byte) {
	*(*Uint16Attr)(unsafe.Pointer(&v[0])) = a
}
func (a Uint16Attr) Size() int {
	return 2
}
func (a Uint16Attr) String() string {
	return StringOf(a)
}
func (a Uint16Attr) Uint() uint16 {
	return uint16(a)
}
func (a Uint16Attr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a.Uint())
	return acc.N, acc.Err
}

type Uint32Attr uint32

func Uint32AttrBytes(b []byte) Uint32Attr {
	return Uint32Attr(*(*uint32)(unsafe.Pointer(&b[0])))
}

func (a Uint32Attr) attr() {}
func (a Uint32Attr) Set(v []byte) {
	*(*Uint32Attr)(unsafe.Pointer(&v[0])) = a
}
func (a Uint32Attr) Size() int {
	return 4
}
func (a Uint32Attr) String() string {
	return StringOf(a)
}
func (a Uint32Attr) Uint() uint32 {
	return uint32(a)
}
func (a Uint32Attr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a.Uint())
	return acc.N, acc.Err
}

type Uint64Attr uint64

func Uint64AttrBytes(b []byte) Uint64Attr {
	return Uint64Attr(*(*uint64)(unsafe.Pointer(&b[0])))
}

func (a Uint64Attr) attr() {}
func (a Uint64Attr) Set(v []byte) {
	*(*Uint64Attr)(unsafe.Pointer(&v[0])) = a
}
func (a Uint64Attr) Size() int {
	return 8
}
func (a Uint64Attr) String() string {
	return StringOf(a)
}
func (a Uint64Attr) Uint() uint64 {
	return uint64(a)
}
func (a Uint64Attr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a.Uint())
	return acc.N, acc.Err
}
