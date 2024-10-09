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

var (
	ErrIncomplete = errors.New("incomplete")
	ErrInvalid    = errors.New("invalid")
	ErrOverrun    = errors.New("overrun")
	ErrUnderrun   = errors.New("underrun")
)

// provided by runtime
//
//go:linkname runtime_rand runtime.rand
func runtime_rand() uint64

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

type Type dnsmessage.Type

const DefaultType = Type(dnsmessage.TypeA)
const ZeroType = Type(0)
const TypePTR = Type(dnsmessage.TypePTR)

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
		dnsmessage.TypeAXFR:  "AXFR",
		dnsmessage.TypeALL:   "ALL",
		dnsmessage.Type(65):  "HTTPS",
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
		"AXFR":  dnsmessage.TypeAXFR,
		"ALL":   dnsmessage.TypeALL,
		"HTTPS": dnsmessage.Type(65),
	}[s]; !ok {
		return ErrInvalid
	} else {
		*p = Type(t)
	}
	return nil
}

// AskUDP sends req then reads response with deadline upto 3 times.
// If successful, this returns the unpacked message and its total length.
func AskUDP(ctx context.Context, udp *net.UDPConn, req []byte) (
	*dnsmessage.Message, int, error,
) {
	buf := make([]byte, 512)
	for tries := 0; tries < 3; tries++ {
		n, err := udp.Write(req)
		if err != nil {
			return nil, 0, err
		} else if n != len(req) {
			return nil, 0, ErrUnderrun
		}
		err = udp.SetReadDeadline(time.Now().Add(time.Second * 3))
		if err != nil {
			return nil, 0, err
		}
		n, err = udp.Read(buf)
		udp.SetReadDeadline(ResetDeadline)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			}
			return nil, 0, err
		}
		msg := new(dnsmessage.Message)
		return msg, n, msg.Unpack(buf[:n])
	}
	return nil, 0, os.ErrDeadlineExceeded
}

func HeaderFlagNames(h *dnsmessage.Header) string {
	const space = " "
	var sb strings.Builder
	var sep string
	if h.Response {
		fmt.Fprint(&sb, "qr")
		sep = space
	}
	if h.Authoritative {
		fmt.Fprint(&sb, sep, "aa")
		sep = space
	}
	if h.Truncated {
		fmt.Fprint(&sb, sep, "tr")
		sep = space
	}
	if h.RecursionDesired {
		fmt.Fprint(&sb, sep, "rd")
		sep = space
	}
	if h.RecursionAvailable {
		fmt.Fprint(&sb, sep, "ra")
		sep = space
	}
	if h.AuthenticData {
		fmt.Fprint(&sb, sep, "ad")
		sep = space
	}
	if h.CheckingDisabled {
		fmt.Fprint(&sb, sep, "cd")
		sep = space
	}
	return sb.String()
}

func NewID() uint16 {
	return uint16(runtime_rand())
}

func NewQuestion(
	buf []byte,
	hdr dnsmessage.Header,
	name dnsmessage.Name,
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
		Name:  name,
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
