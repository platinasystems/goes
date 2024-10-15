// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdns

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"
	_ "unsafe"

	"golang.org/x/net/dns/dnsmessage"
)

//go:embed types.txt
var TypesTxt string

//go:linkname runtime_rand runtime.rand
func runtime_rand() uint64

//go:linkname unpackCNAMEResource golang.org/x/net/dns/dnsmessage.unpackCNAMEResource
func unpackCNAMEResource([]byte, int) (dnsmessage.CNAMEResource, error)

//go:linkname unpackMXResource golang.org/x/net/dns/dnsmessage.unpackMXResource
func unpackMXResource([]byte, int) (dnsmessage.MXResource, error)

//go:linkname unpackUint16 golang.org/x/net/dns/dnsmessage.unpackUint16
func unpackUint16([]byte, int) (uint16, int, error)

var (
	ErrIncomplete = errors.New("incomplete")
	ErrInvalid    = errors.New("invalid")
	ErrOverrun    = errors.New("overrun")
	ErrUnderrun   = errors.New("underrun")
)

const MaxPacketSize = 1232

type Class dnsmessage.Class

const DefaultClass = Class(dnsmessage.ClassINET)
const ZeroClass = Class(0)

func (v Class) Class() dnsmessage.Class {
	return dnsmessage.Class(v)
}

func (v Class) MarshalText() ([]byte, error) {
	if v == 0 {
		return []byte{}, nil
	}
	s, ok := map[dnsmessage.Class]string{
		dnsmessage.ClassINET:   "IN",
		dnsmessage.ClassCSNET:  "CS",
		dnsmessage.ClassCHAOS:  "CH",
		dnsmessage.ClassHESIOD: "HS",
		dnsmessage.ClassANY:    "ANY",
	}[dnsmessage.Class(v)]
	if !ok {
		s = fmt.Sprintf("CLASS%d", v)
	}
	return []byte(s), nil
}

// If current `Class` is zero,
// `Class.UnmarshalText` the first argument.
// If mismatch or there were no arguments,
// set to DefaultClass and return args unchanged;
// otherwise, set with match and
// pull the first argument from the returned list.
func (p *Class) Pull(args []string) ([]string, error) {
	var err error
	if *p != 0 {
	} else if len(args) == 0 {
		*p = DefaultClass
	} else if err = p.UnmarshalText([]byte(args[0])); err == nil {
		args = args[1:]
	} else if errors.Is(err, ErrInvalid) {
		*p = DefaultClass
		err = nil
	}
	return args, err
}

func (v Class) String() string {
	b, err := v.MarshalText()
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (p *Class) UnmarshalText(text []byte) error {
	s := string(text)
	if len(s) == 0 {
		*p = 0
	} else if strings.HasPrefix(s, "CLASS") {
		s = strings.TrimPrefix(s, "CLASS")
		u, err := strconv.ParseUint(s, 0, 8)
		if err != nil {
			return err
		}
		*p = Class(u)
	} else if c, ok := map[string]dnsmessage.Class{
		"IN":  dnsmessage.ClassINET,
		"CS":  dnsmessage.ClassCSNET,
		"CH":  dnsmessage.ClassCHAOS,
		"HS":  dnsmessage.ClassHESIOD,
		"ANY": dnsmessage.ClassANY,
	}[s]; !ok {
		return ErrInvalid
	} else {
		*p = Class(c)
	}
	return nil
}

// EDNS(0) wire constants.
const (
	EDNS0Version = 0

	EDNS0DNSSECOK     = 0x00008000
	EDNSVersionMask   = 0x00ff0000
	EDNS0DNSSECOKMask = 0x00ff8000
)

func EDNSVersion(ttl uint32) uint8 {
	return uint8(ttl >> 16)
}

func HasEDNS0DNSSECOK(ttl uint32) bool {
	return (ttl & EDNS0DNSSECOK) != 0
}

func EDNS0MBZ(ttl uint32) uint16 {
	return uint16(ttl & 0x7fff)
}

type Name struct {
	dnsmessage.Name
}

func (v Name) MarshalText() ([]byte, error) {
	return v.Data[:v.Length], nil
}

// If `Name` is empty, pull first argument;
// otherwise return unchanged arguments.
// If there are no arguments, return `ErrIncomplete`.
func (p *Name) Pull(args []string) ([]string, error) {
	var err error
	if p.Length == 0 {
		if len(args) == 0 {
			err = ErrIncomplete
		} else if err = p.UnmarshalText([]byte(args[0])); err == nil {
			args = args[1:]
		}
	}
	return args, err

}

func (p *Name) Reverse(addr netip.Addr) {
	var n int
	p.Length = 0
	b := bytes.NewBuffer(p.Data[:])
	if addr.Is4() {
		a4 := addr.As4()
		n, _ = fmt.Fprintf(b, "%d.%d.%d.%d.in-addr.arpa.",
			a4[3], a4[2], a4[1], a4[0])
		p.Length += uint8(n)
	} else {
		sl := addr.AsSlice()
		for i := len(sl) - 1; i >= 0; i-- {
			n, _ = fmt.Fprintf(b, "%d.", sl[i]>>4)
			p.Length += uint8(n)
			n, _ = fmt.Fprintf(b, "%d.", sl[i]&0xf)
			p.Length += uint8(n)
		}
		n, _ = fmt.Fprint(b, "ip6.arpa.")
		p.Length += uint8(n)
	}
}

func (p *Name) UnmarshalText(text []byte) error {
	data := p.Data[:]
	if len(text) > len(data) {
		return ErrOverrun
	}
	p.Length = uint8(copy(data, text))
	return nil
}

const (
	OpCodeQuery dnsmessage.OpCode = iota
	OpCodeIQuery
	OpCodeStatus
)

var OpCodeName = map[dnsmessage.OpCode]string{
	OpCodeQuery:  "QUERY",
	OpCodeIQuery: "IQUERY",
	OpCodeStatus: "STATUS",
}

var RCodeName = map[dnsmessage.RCode]string{
	dnsmessage.RCodeSuccess:        "NoError",
	dnsmessage.RCodeFormatError:    "FormatError",
	dnsmessage.RCodeServerFailure:  "ServerFailure",
	dnsmessage.RCodeNameError:      "NameError",
	dnsmessage.RCodeNotImplemented: "NotImplemented",
	dnsmessage.RCodeRefused:        "Refused",
}

var ResetDeadline time.Time

var Resolver = &net.Resolver{
	PreferGo:     true,
	StrictErrors: true,
}

const (
	dnsmessageTypeSVCB  dnsmessage.Type = 64
	dnsmessageTypeHTTPS dnsmessage.Type = 65
	dnsmessageTypeCAA   dnsmessage.Type = 257
)

type Type dnsmessage.Type

const (
	DefaultType = Type(dnsmessage.TypeA)
	ZeroType    = Type(0)
	TypeA       = Type(dnsmessage.TypeA)
	TypeNS      = Type(dnsmessage.TypeNS)
	TypeCNAME   = Type(dnsmessage.TypeCNAME)
	TypeSOA     = Type(dnsmessage.TypeSOA)
	TypePTR     = Type(dnsmessage.TypePTR)
	TypeMX      = Type(dnsmessage.TypeMX)
	TypeTXT     = Type(dnsmessage.TypeTXT)
	TypeAAAA    = Type(dnsmessage.TypeAAAA)
	TypeSRV     = Type(dnsmessage.TypeSRV)
	TypeOPT     = Type(dnsmessage.TypeOPT)
	TypeWKS     = Type(dnsmessage.TypeWKS)
	TypeHINFO   = Type(dnsmessage.TypeHINFO)
	TypeMINFO   = Type(dnsmessage.TypeMINFO)
	TypeSVCB    = Type(dnsmessageTypeSVCB)
	TypeHTTPS   = Type(dnsmessageTypeHTTPS)
	TypeAXFR    = Type(dnsmessage.TypeAXFR)
	TypeALL     = Type(dnsmessage.TypeALL)
	TypeANY     = TypeALL
	TypeCAA     = Type(dnsmessageTypeCAA)
)

func (v Type) MarshalText() ([]byte, error) {
	if v == 0 {
		return []byte{}, nil
	}
	s, ok := map[dnsmessage.Type]string{
		dnsmessage.TypeA:     "A",
		dnsmessage.TypeNS:    "NS",
		dnsmessage.TypeCNAME: "CNAME",
		dnsmessage.TypeSOA:   "SOA",
		dnsmessage.TypePTR:   "PTR",
		dnsmessage.TypeMX:    "MX",
		dnsmessage.TypeTXT:   "TXT",
		dnsmessage.TypeAAAA:  "AAAA",
		dnsmessage.TypeSRV:   "SRV",
		dnsmessage.TypeOPT:   "OPT",
		dnsmessage.TypeWKS:   "WKS",
		dnsmessage.TypeHINFO: "HINFO",
		dnsmessage.TypeMINFO: "MINFO",
		dnsmessageTypeSVCB:   "SVCB",
		dnsmessageTypeHTTPS:  "HTTPS",
		dnsmessage.TypeAXFR:  "AXFR",
		dnsmessage.TypeALL:   "ANY",
		dnsmessageTypeCAA:    "CAA",
	}[dnsmessage.Type(v)]
	if !ok {
		s = fmt.Sprintf("TYPE%d", v)
	}
	return []byte(s), nil
}

// If current `Type` is zero,
// `Type.UnmarshalText` the first argument.
// If mismatch or there were no arguments,
// set to DefaultType and return args unchanged;
// otherwise, set with match and
// pull the first argument from the returned list.
func (p *Type) Pull(args []string) ([]string, error) {
	var err error
	if *p != 0 {
	} else if len(args) == 0 {
		*p = DefaultType
	} else if err = p.UnmarshalText([]byte(args[0])); err == nil {
		args = args[1:]
	} else if errors.Is(err, ErrInvalid) {
		*p = DefaultType
		err = nil
	}
	return args, err
}

func (v Type) String() string {
	b, err := v.MarshalText()
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (p *Type) UnmarshalText(text []byte) error {
	s := string(text)
	if len(s) == 0 {
		*p = 0
	} else if strings.HasPrefix(s, "TYPE") {
		s = strings.TrimPrefix(s, "TYPE")
		u, err := strconv.ParseUint(s, 0, 8)
		if err != nil {
			return err
		}
		*p = Type(u)
	} else if t, ok := map[string]dnsmessage.Type{
		"A":     dnsmessage.TypeA,
		"NS":    dnsmessage.TypeNS,
		"CNAME": dnsmessage.TypeCNAME,
		"SOA":   dnsmessage.TypeSOA,
		"PTR":   dnsmessage.TypePTR,
		"MX":    dnsmessage.TypeMX,
		"TXT":   dnsmessage.TypeTXT,
		"AAAA":  dnsmessage.TypeAAAA,
		"SRV":   dnsmessage.TypeSRV,
		"OPT":   dnsmessage.TypeOPT,
		"WKS":   dnsmessage.TypeWKS,
		"HINFO": dnsmessage.TypeHINFO,
		"MINFO": dnsmessage.TypeMINFO,
		"SVCB":  dnsmessageTypeSVCB,
		"HTTPS": dnsmessageTypeHTTPS,
		"AXFR":  dnsmessage.TypeAXFR,
		"ALL":   dnsmessage.TypeALL,
		"ANY":   dnsmessage.TypeALL,
		"CAA":   dnsmessageTypeCAA,
	}[s]; !ok {
		return ErrInvalid
	} else {
		*p = Type(t)
	}
	return nil
}

type AResource = dnsmessage.AResource
type NSResource = dnsmessage.NSResource
type CNAMEResource = dnsmessage.CNAMEResource
type SOAResource = dnsmessage.SOAResource
type PTRResource = dnsmessage.PTRResource
type MXResource = dnsmessage.MXResource
type TXTResource = dnsmessage.TXTResource
type AAAAResource = dnsmessage.AAAAResource
type SRVResource = dnsmessage.SRVResource
type OPTResource = dnsmessage.OPTResource
type UnknownResource = dnsmessage.UnknownResource

type SVCBResource struct {
	Pri uint16
	dnsmessage.Name
	Params []SVCBParam
}

type SVCBParam struct {
	Key   uint16
	Value any
}

type HTTPSResource struct{ SVCBResource }

func (r *HTTPSResource) UnmarshalBinary(msg []byte) error {
	return r.SVCBResource.UnmarshalBinary(msg)
}

type CAAResource struct{ SVCBResource }

func (r *CAAResource) UnmarshalBinary(msg []byte) error {
	return r.SVCBResource.UnmarshalBinary(msg)
}

func AnswerString(r dnsmessage.Resource) (s string) {
	switch r.Header.Type {
	case dnsmessage.TypeA:
		rb := r.Body.(*AResource)
		s = net.IP(rb.A[:]).String()
	case dnsmessage.TypeNS:
		s = r.Body.(*NSResource).NS.String()
	case dnsmessage.TypeCNAME:
		s = r.Body.(*CNAMEResource).CNAME.String()
	case dnsmessage.TypeSOA:
		rb := r.Body.(*SOAResource)
		s = fmt.Sprintf("ns %v, mbox %v, s/n %d",
			rb.NS, rb.MBox, rb.Serial)
	case dnsmessage.TypePTR:
		s = r.Body.(*PTRResource).PTR.String()
	case dnsmessage.TypeMX:
		rb := r.Body.(*MXResource)
		s = fmt.Sprintf("%v, pref %d", rb.MX, rb.Pref)
	case dnsmessage.TypeTXT:
		s = strings.Join(r.Body.(*TXTResource).TXT, " ")
	case dnsmessage.TypeAAAA:
		rb := r.Body.(*AAAAResource)
		s = net.IP(rb.AAAA[:]).String()
	case dnsmessage.TypeSRV:
		rb := r.Body.(*SRVResource)
		s = fmt.Sprintf("%v, port %d, pri %d, weight %d",
			rb.Target, rb.Port, rb.Priority, rb.Weight)
	case dnsmessageTypeSVCB:
		var rb SVCBResource
		data := r.Body.(*UnknownResource).Data
		if err := rb.UnmarshalBinary(data); err != nil {
			s = err.Error()
		} else {
			s = rb.String()
		}
	case dnsmessageTypeHTTPS:
		var rb HTTPSResource
		data := r.Body.(*UnknownResource).Data
		if err := rb.UnmarshalBinary(data); err != nil {
			s = err.Error()
		} else {
			s = rb.String()
		}
	case dnsmessageTypeCAA:
		var rb CAAResource
		data := r.Body.(*UnknownResource).Data
		if err := rb.UnmarshalBinary(data); err != nil {
			s = err.Error()
		} else {
			s = rb.String()
		}
	default:
		s = fmt.Sprintf("%#x", r.Body.(*UnknownResource).Data)
	}
	return
}

// TenaciousAsk resends the buffered request every 1 sec until it receives a
// response or context is cancelled.
// If successful, it returns the response within the same buffer.
func TenaciousAsk(ctx context.Context, udp *net.UDPConn, buf []byte) (
	[]byte, error,
) {
	for {
		err := ctx.Err()
		if err != nil {
			return buf[:0], err
		}
		n, err := udp.Write(buf)
		if err != nil {
			return buf[:0], err
		} else if n != len(buf) {
			return buf[:0], ErrUnderrun
		}
		err = udp.SetReadDeadline(time.Now().Add(time.Second))
		if err != nil {
			return buf[:0], err
		}
		n, err = udp.Read(buf[:cap(buf)])
		udp.SetReadDeadline(ResetDeadline)
		if err == nil {
			return buf[:n], nil
		} else if !errors.Is(err, os.ErrDeadlineExceeded) {
			return buf[:0], err
		}
	}
}

// TimeLimitedAsk is a `TenaciousAsk` with a deadlined context.
func TimeLimitedAsk(
	ctx context.Context, udp *net.UDPConn, buf []byte, dur time.Duration,
) ([]byte, error) {
	dl := time.Now().Add(dur)
	cctx, cancel := context.WithDeadline(ctx, dl)
	defer cancel()
	return TenaciousAsk(cctx, udp, buf)
}

type HeaderFlags struct{ dnsmessage.Header }

func (h HeaderFlags) Format(w fmt.State, verb rune) {
	const space = " "
	var sep string
	if h.Response {
		fmt.Fprint(w, "qr")
		sep = space
	}
	if h.Authoritative {
		fmt.Fprint(w, sep, "aa")
		sep = space
	}
	if h.Truncated {
		fmt.Fprint(w, sep, "tr")
		sep = space
	}
	if h.RecursionDesired {
		fmt.Fprint(w, sep, "rd")
		sep = space
	}
	if h.RecursionAvailable {
		fmt.Fprint(w, sep, "ra")
		sep = space
	}
	if h.AuthenticData {
		fmt.Fprint(w, sep, "ad")
		sep = space
	}
	if h.CheckingDisabled {
		fmt.Fprint(w, sep, "cd")
	}
}

func NewID() uint16 {
	return uint16(runtime_rand())
}

func NewQuestion(
	buf []byte,
	hdr dnsmessage.Header,
	name Name,
	t Type,
	c Class,
) ([]byte, error) {
	var rh dnsmessage.ResourceHeader
	if buf == nil {
		buf = make([]byte, 0, 512)
	}
	mb := dnsmessage.NewBuilder(buf, hdr)
	err := mb.StartQuestions()
	if err != nil {
		return nil, err
	}
	err = mb.Question(dnsmessage.Question{
		Name:  name.Name,
		Type:  dnsmessage.Type(t),
		Class: dnsmessage.Class(c),
	})
	if err != nil {
		return nil, err
	}
	err = mb.StartAdditionals()
	if err != nil {
		return nil, err
	}
	err = rh.SetEDNS0(MaxPacketSize, dnsmessage.RCodeSuccess, false)
	if err != nil {
		return nil, err
	}
	err = mb.OPTResource(rh, dnsmessage.OPTResource{})
	if err != nil {
		return nil, err
	}
	return mb.Finish()
}

func NewUDP(ctx context.Context, svr string, port uint) (*net.UDPConn, error) {
	a := &net.UDPAddr{
		IP:   net.ParseIP(svr),
		Port: int(port),
	}
	if a.IP == nil {
		sl, err := Resolver.LookupIPAddr(ctx, svr)
		if err != nil {
			return nil, err
		}
		a.IP = sl[0].IP
		a.Zone = sl[0].Zone
	}
	nw := "udp"
	if a.IP.To4() != nil {
		nw = "udp4"
	} else if a.IP.To16() != nil {
		nw = "udp6"
	}
	return net.DialUDP(nw, nil, a)
}

const (
	SVCBParamKeyMandatory = iota
	SVCBParamKeyALPN
	SVCBParamKeyNoDefaultALPN
	SVCBParamKeyPort
	SVCBParamKeyIPV4Hint
	SVCBParamKeyECH
	SVCBParamKeyIPV6Hint
	SVCBParamKeyDOHPath
)

func (r *SVCBResource) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d", r.Pri)
	fmt.Fprint(&sb, " ", r.Name)
	for _, param := range r.Params {
		fmt.Fprint(&sb, " ", param)
	}
	return sb.String()
}

func (r *SVCBResource) UnmarshalBinary(msg []byte) error {
	mx, err := unpackMXResource(msg, 0)
	if err != nil {
		return err
	}
	r.Pri = mx.Pref
	r.Name = mx.MX
	if r.Pri == 0 {
		// Alias mode [rfc9460 section 2.4.2]
		return nil
	}
	msg = msg[2+r.Name.Length:]
	for len(msg) >= 4 {
		u16, n, err := unpackUint16(msg, 0)
		if err != nil {
			return err
		}
		param := SVCBParam{Key: u16}
		msg = msg[n:]
		u16, n, err = unpackUint16(msg, 0)
		if err != nil {
			return err
		}
		l := int(u16)
		msg = msg[n:]
		switch param.Key {
		case SVCBParamKeyMandatory:
			// FIXME txt
		case SVCBParamKeyALPN:
			var alpns []string
			for i := 0; i < l; i += 1 + n {
				n = int(msg[i])
				if 1+n > l {
					return ErrUnderrun
				}
				b := make([]byte, n)
				copy(b, msg[i+1:i+1+n])
				alpns = append(alpns, string(b))
			}
			param.Value = alpns
		case SVCBParamKeyNoDefaultALPN:
			// empty
		case SVCBParamKeyPort:
			param.Value, n, err = unpackUint16(msg, 0)
		case SVCBParamKeyIPV4Hint:
			if len(msg) < 4 {
				return ErrUnderrun
			}
			param.Value, _ = netip.AddrFromSlice(msg[:4])
		case SVCBParamKeyECH:
			// FIXME decode base64
		case SVCBParamKeyIPV6Hint:
			if len(msg) < 16 {
				return ErrUnderrun
			}
			param.Value, _ = netip.AddrFromSlice(msg[:16])
		case SVCBParamKeyDOHPath:
			// FIXME decode DOH path
		default:
			param.Value = msg[:l]
		}
		r.Params = append(r.Params, param)
		msg = msg[l:]
	}
	return nil
}

func (param SVCBParam) Format(w fmt.State, verb rune) {
	if s, ok := map[uint16]string{
		SVCBParamKeyMandatory:     "mandatory",
		SVCBParamKeyALPN:          "alpn",
		SVCBParamKeyNoDefaultALPN: "no-default-alpn",
		SVCBParamKeyPort:          "port",
		SVCBParamKeyIPV4Hint:      "ipv4hint",
		SVCBParamKeyECH:           "ech",
		SVCBParamKeyIPV6Hint:      "ipv6hint",
		SVCBParamKeyDOHPath:       "dohpath",
	}[param.Key]; ok {
		fmt.Fprint(w, s, "=")
	} else {
		fmt.Fprintf(w, "key%d=", param.Key)
	}
	switch param.Key {
	case SVCBParamKeyMandatory:
		fmt.Fprint(w, "FIXME")
	case SVCBParamKeyALPN:
		if ss, ok := param.Value.([]string); !ok {
			fmt.Fprint(w, "invalid")
		} else {
			fmt.Fprintf(w, "%q", strings.Join(ss, ","))
		}
	case SVCBParamKeyNoDefaultALPN:
		// empty
	case SVCBParamKeyPort:
		fmt.Fprintf(w, "%d", param.Value)
	case SVCBParamKeyIPV4Hint:
		fmt.Fprintf(w, "%q", param.Value)
	case SVCBParamKeyECH:
		fmt.Fprint(w, "FIXME")
	case SVCBParamKeyIPV6Hint:
		fmt.Fprintf(w, "%q", param.Value)
	case SVCBParamKeyDOHPath:
		fmt.Fprint(w, "FIXME")
	}
}
