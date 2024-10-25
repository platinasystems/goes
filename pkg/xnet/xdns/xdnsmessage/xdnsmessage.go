// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"math"
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
	ErrIncomplete  = errors.New("incomplete")
	ErrInvalid     = errors.New("invalid")
	ErrOverrun     = errors.New("overrun")
	ErrUnderrun    = errors.New("underrun")
	ErrUnsupported = errors.New("unsupported")

	FIXME = errors.New("FIXME")
)

const MaxPacketSize = 1232

type Class uint16

const (
	Class0 Class = iota
	ClassINET
	ClassCSNET
	ClassCHAOS
	ClassHESIOD
)

const ClassANY = Class(255)
const DefaultClass = Class(dnsmessage.ClassINET)

func ClassNamed(s string) (Class, error) {
	if len(s) == 0 {
		return Class0, nil
	}
	if strings.HasPrefix(s, "CLASS") {
		s = strings.TrimPrefix(s, "CLASS")
		u, err := strconv.ParseUint(s, 10, 16)
		return Class(u), err
	}
	c, ok := map[string]Class{
		"IN":  ClassINET,
		"CS":  ClassCSNET,
		"CH":  ClassCHAOS,
		"HS":  ClassHESIOD,
		"ANY": ClassANY,
	}[s]
	if !ok {
		return Class0, ErrInvalid
	}
	return Class(c), nil
}

func (v Class) Class() dnsmessage.Class {
	return dnsmessage.Class(v)
}

func (v Class) IsINET() bool {
	return dnsmessage.Class(v) == dnsmessage.ClassINET
}

func (v Class) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// If current `Class` is zero,
// `ClassNamed` the first argument.
// If mismatch or there were no arguments,
// set to DefaultClass and return args unchanged;
// otherwise, set with match and
// pull the first argument from the returned list.
func (p *Class) Pull(args []string) ([]string, error) {
	var err error
	if *p != 0 {
	} else if len(args) == 0 {
		*p = DefaultClass
	} else if *p, err = ClassNamed(args[0]); err == nil {
		args = args[1:]
	} else if errors.Is(err, ErrInvalid) {
		*p = DefaultClass
		err = nil
	}
	return args, err
}

func (v Class) String() string {
	if v == 0 {
		return ""
	}
	s, ok := map[Class]string{
		ClassINET:   "IN",
		ClassCSNET:  "CS",
		ClassCHAOS:  "CH",
		ClassHESIOD: "HS",
	}[v]
	if !ok {
		if v == ClassANY {
			s = "ANY"
		} else {
			s = fmt.Sprintf("CLASS%d", v)
		}
	}
	return s
}

func (p *Class) UnmarshalText(text []byte) (err error) {
	*p, err = ClassNamed(string(text))
	return
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

func (p *Name) Scan(r fmt.ScanState, verb rune) error {
	token, err := r.Token(true, nil)
	if err == nil {
		err = p.UnmarshalText(token)
	}
	return err
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

type Type uint16

const (
	Type0 Type = iota
	TypeA
	TypeNS
	_ // 3
	_ // 4
	TypeCNAME
	TypeSOA
	_ // 7
	_ // 8
	_ // 9
	_ // 10
	TypeWKS
	TypePTR
	TypeHINFO
	TypeMINFO
	TypeMX
	TypeTXT
	_ // 17
	_ // 18
	_ // 19
	_ // 20
	_ // 21
	_ // 22
	_ // 23
	_ // 24
	_ // 25
	_ // 26
	_ // 27
	TypeAAAA
	TypeLOC
	_ // 30
	_ // 31
	_ // 32
	TypeSRV
	_ // 34
	_ // 35
	_ // 36
	_ // 37
	_ // 38
	_ // 39
	_ // 40
	TypeOPT
)

const (
	TypeSVCB  = Type(64)
	TypeHTTPS = Type(65)
	TypeAXFR  = Type(252)
	TypeALL   = Type(255)
	TypeANY   = TypeALL
	TypeCAA   = Type(257)
)

const DefaultType = TypeA

func TypeNamed(s string) (Type, error) {
	if len(s) == 0 {
		return Type0, nil
	}
	if strings.HasPrefix(s, "TYPE") {
		s = strings.TrimPrefix(s, "TYPE")
		u, err := strconv.ParseUint(s, 10, 16)
		return Type(u), err
	}
	t, ok := map[string]Type{
		"A":     TypeA,
		"AAAA":  TypeAAAA,
		"ALL":   TypeALL,
		"ANY":   TypeANY,
		"AXFR":  TypeAXFR,
		"CAA":   TypeCAA,
		"CNAME": TypeCNAME,
		"HINFO": TypeHINFO,
		"HTTPS": TypeHTTPS,
		"LOC":   TypeLOC,
		"MINFO": TypeMINFO,
		"MX":    TypeMX,
		"NS":    TypeNS,
		"OPT":   TypeOPT,
		"PTR":   TypePTR,
		"SOA":   TypeSOA,
		"SRV":   TypeSRV,
		"SVCB":  TypeSVCB,
		"TXT":   TypeTXT,
		"WKS":   TypeWKS,
	}[s]
	if !ok {
		return Type0, ErrInvalid
	}
	return t, nil
}

func (v Type) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// If current `Type` is zero,
// `TypeNamed` the first argument.
// If mismatch or there were no arguments,
// set to DefaultType and return args unchanged;
// otherwise, set with match and
// pull the first argument from the returned list.
func (p *Type) Pull(args []string) ([]string, error) {
	var err error
	if *p != 0 {
	} else if len(args) == 0 {
		*p = DefaultType
	} else if *p, err = TypeNamed(args[0]); err == nil {
		args = args[1:]
	} else if errors.Is(err, ErrInvalid) {
		*p = DefaultType
		err = nil
	}
	return args, err
}

func (v Type) String() string {
	if v == 0 {
		return ""
	}
	s, ok := map[Type]string{
		TypeA:     "A",
		TypeNS:    "NS",
		TypeCNAME: "CNAME",
		TypeSOA:   "SOA",
		TypePTR:   "PTR",
		TypeMX:    "MX",
		TypeTXT:   "TXT",
		TypeAAAA:  "AAAA",
		TypeLOC:   "LOC",
		TypeSRV:   "SRV",
		TypeOPT:   "OPT",
		TypeWKS:   "WKS",
		TypeHINFO: "HINFO",
		TypeMINFO: "MINFO",
		TypeSVCB:  "SVCB",
		TypeHTTPS: "HTTPS",
		TypeAXFR:  "AXFR",
		TypeALL:   "ANY",
		TypeCAA:   "CAA",
	}[v]
	if !ok {
		s = fmt.Sprintf("TYPE%d", v)
	}
	return s
}

func (p *Type) UnmarshalText(text []byte) (err error) {
	*p, err = TypeNamed(string(text))
	return
}

type Resource interface {
	String() string
	Type() Type
}

func ParseResource(t Type, tokens []string) (Resource, error) {
	switch t {
	case TypeA:
		return parseA(tokens)
	case TypeNS:
		return parseNS(tokens)
	case TypeCNAME:
		return parseCNAME(tokens)
	case TypeSOA:
		return parseSOA(tokens)
	case TypePTR:
		return parsePTR(tokens)
	case TypeHINFO:
		return parseHINFO(tokens)
	case TypeMINFO:
		return parseMINFO(tokens)
	case TypeMX:
		return parseMX(tokens)
	case TypeTXT:
		return parseTXT(tokens)
	case TypeAAAA:
		return parseAAAA(tokens)
	case TypeLOC:
		return parseLOC(tokens)
	case TypeSRV:
		return parseSRV(tokens)
	}
	return nil, fmt.Errorf("%v: %w", t, ErrUnsupported)
}

type A netip.Addr

func parseA(tokens []string) (A, error) {
	addr, err := parseAddr(tokens)
	return A(addr), err
}

func (v A) String() string { return netip.Addr(v).String() }

func (A) Type() Type { return TypeA }

type AAAA netip.Addr

func parseAAAA(tokens []string) (AAAA, error) {
	addr, err := parseAddr(tokens)
	return AAAA(addr), err
}

func (v AAAA) String() string { return netip.Addr(v).String() }

func (AAAA) Type() Type { return TypeAAAA }

type CNAME string

func parseCNAME(tokens []string) (CNAME, error) {
	s, err := parseString(tokens)
	return CNAME(s), err
}

func (CNAME) Type() Type { return TypeCNAME }

func (v CNAME) String() string { return string(v) }

type HINFO struct {
	CPU, OS string
}

func parseHINFO(tokens []string) (HINFO, error) {
	if len(tokens) < 2 {
		return HINFO{}, ErrIncomplete
	}
	return HINFO{
		CPU: tokens[0],
		OS:  tokens[1],
	}, nil
}

func (HINFO) Type() Type { return TypeHINFO }

func (v HINFO) String() string { return v.CPU + " " + v.OS }

// RFC-1876
type LOC struct {
	// Must be 0
	Version,
	// Centimeters in nibbles of N1eN2
	Size,
	// Precision in Size units
	Horizontal, Vertical uint8
	// Thousandths arc seconds
	Latitude, Longitude,
	// Centimeters + 100,000m
	Altitude uint32
}

const locNE = 1 << 31
const locArcSec = 1000
const locArcMin = locArcSec * 60
const locArcDeg = locArcMin * 60

func parseLOC(tokens []string) (LOC, error) {
	var err error
	var loc LOC
	loc.Latitude, tokens, err = parseLOCArc(tokens)
	if err != nil {
		return loc, fmt.Errorf("latitude: %w", err)
	}
	loc.Longitude, tokens, err = parseLOCArc(tokens)
	if err != nil {
		return loc, fmt.Errorf("longitude: %w", err)
	}
	loc.Altitude, tokens, err = parseLOCAlt(tokens)
	if err != nil {
		return loc, fmt.Errorf("altitude: %w", err)
	}
	if len(tokens) == 0 {
		return loc, nil
	}
	loc.Size, tokens, err = parseLOCSizePrecision(tokens)
	if err != nil {
		return loc, fmt.Errorf("size: %w", err)
	}
	if len(tokens) == 0 {
		return loc, nil
	}
	loc.Horizontal, tokens, err = parseLOCSizePrecision(tokens)
	if err != nil {
		return loc, fmt.Errorf("horizontal: %w", err)
	}
	if len(tokens) == 0 {
		return loc, nil
	}
	loc.Vertical, tokens, err = parseLOCSizePrecision(tokens)
	if err != nil {
		return loc, fmt.Errorf("vertical: %w", err)
	}
	return loc, nil
}

func parseLOCAlt(tokens []string) (uint32, []string, error) {
	var alt float32
	if len(tokens) == 0 {
		return 0, tokens, ErrIncomplete
	}
	_, err := fmt.Sscan(strings.TrimSuffix(tokens[0], "m"), &alt)
	if err == nil {
		tokens = tokens[1:]
	}
	return uint32((alt * 100) + 100000), tokens, nil
}

func parseLOCArc(tokens []string) (uint32, []string, error) {
	var arc, deg, min uint32
	var sec float32
	if len(tokens) < 2 {
		return arc, tokens, ErrIncomplete
	}
	_, err := fmt.Sscan(tokens[0], &deg)
	if err != nil {
		return arc, tokens, err
	}
	arc = deg * locArcDeg
	tokens = tokens[1:]
	switch tokens[0] {
	case "N", "n", "E", "e":
		arc |= locNE
		tokens = tokens[1:]
	case "S", "s", "W", "w":
		tokens = tokens[1:]
	default:
		if _, err = fmt.Sscan(tokens[0], &min); err != nil {
			return arc, tokens, err
		}
		arc += min * locArcMin
		tokens = tokens[1:]
		if len(tokens) == 0 {
			return arc, tokens, ErrIncomplete
		}
		switch tokens[0] {
		case "N", "E":
			arc |= locNE
			tokens = tokens[1:]
		case "S", "W":
			tokens = tokens[1:]
		default:
			if _, err = fmt.Sscan(tokens[0], &sec); err != nil {
				return arc, tokens, err
			}
			arc += uint32(sec * locArcSec)
			tokens = tokens[1:]
			if len(tokens) == 0 {
				return arc, tokens, ErrIncomplete
			}
			switch tokens[0] {
			case "N", "n", "E", "e":
				arc |= locNE
				tokens = tokens[1:]
			case "S", "s", "W", "w":
				tokens = tokens[1:]
			default:
				return arc, tokens, ErrInvalid
			}
		}
	}
	return arc, tokens, nil
}

func parseLOCSizePrecision(tokens []string) (uint8, []string, error) {
	var sp float32
	_, err := fmt.Sscan(strings.TrimSuffix(tokens[0], "m"), &sp)
	if err != nil {
		return 0, tokens, err
	}
	tokens = tokens[1:]
	var exp uint8
	n := uint64(sp * 100)
	for ; n >= 10; exp++ {
		n /= 10
	}
	return (uint8(n) << 4) | exp, tokens, nil
}

func (v LOC) String() string {
	var sb strings.Builder
	formatLOCArc(&sb, v.Latitude&(locNE-1))
	hem := "S"
	if (v.Latitude & locNE) != 0 {
		hem = "N"
	}
	fmt.Fprint(&sb, " ", hem, "\n")
	formatLOCArc(&sb, v.Longitude&(locNE-1))
	hem = "W"
	if (v.Longitude & locNE) != 0 {
		hem = "E"
	}
	fmt.Fprint(&sb, " ", hem, "\n")
	fmt.Fprintf(&sb, "%.2fm", (float64(v.Altitude)-100000)/100)
	if v.Size != 0 {
		fmt.Fprintf(&sb, "\n%.2fm", locSizePrecision(v.Size))
	}
	if v.Horizontal != 0 {
		fmt.Fprintf(&sb, "\n%.2fm", locSizePrecision(v.Horizontal))
	}
	if v.Vertical != 0 {
		fmt.Fprintf(&sb, "\n%.2fm", locSizePrecision(v.Vertical))
	}
	return sb.String()
}

func formatLOCArc(w io.Writer, arc uint32) {
	fmt.Fprintf(w, "%d", arc/locArcDeg)
	arc %= locArcDeg
	if arc == 0 {
		return
	}
	fmt.Fprintf(w, " %d", arc/locArcMin)
	arc %= locArcMin
	if arc == 0 {
		return
	}
	fmt.Fprintf(w, " %d.%03d", arc/locArcSec, arc%locArcSec)
}

func locSizePrecision(n uint8) float64 {
	return (float64(n>>4) * math.Pow10(int(n&0xf))) / 100
}

func (LOC) Type() Type { return TypeLOC }

type MINFO struct {
	RMAILBX, EMAILBX string
}

func parseMINFO(tokens []string) (MINFO, error) {
	if len(tokens) < 2 {
		return MINFO{}, ErrIncomplete
	}
	return MINFO{
		RMAILBX: tokens[0],
		EMAILBX: tokens[1],
	}, nil
}

func (MINFO) Type() Type { return TypeMINFO }

func (v MINFO) String() string { return v.RMAILBX + " " + v.EMAILBX }

type MX struct {
	Preference uint16
	Exchange   string
}

func (MX) Type() Type { return TypeMX }

func parseMX(tokens []string) (mx MX, err error) {
	if len(tokens) < 2 {
		err = ErrIncomplete
	}
	if err == nil {
		_, err = fmt.Sscan(tokens[0], &mx.Preference)
	}
	if err == nil {
		mx.Exchange = strings.Clone(tokens[1])
	}
	return
}

func (v MX) String() string {
	return fmt.Sprintf("%d %s", v.Preference, v.Exchange)
}

type NS string

func parseNS(tokens []string) (NS, error) {
	s, err := parseString(tokens)
	return NS(s), err
}

func (v NS) String() string { return string(v) }

func (NS) Type() Type { return TypeNS }

type PTR string

func parsePTR(tokens []string) (PTR, error) {
	s, err := parseString(tokens)
	return PTR(s), err
}

func (v PTR) String() string { return string(v) }

func (PTR) Type() Type { return TypePTR }

type SOA struct {
	MName, RName string

	Serial, Refresh, Retry, Expire, Minimum uint32
}

func parseSOA(tokens []string) (soa SOA, err error) {
	if len(tokens) < 7 {
		err = ErrIncomplete
	} else {
		soa.MName = strings.Clone(tokens[0])
		soa.RName = strings.Clone(tokens[1])
		_, err = fmt.Sscan(tokens[2], &soa.Serial)
		if err == nil {
			_, err = fmt.Sscan(tokens[3], &soa.Refresh)
		}
		if err == nil {
			_, err = fmt.Sscan(tokens[4], &soa.Retry)
		}
		if err == nil {
			_, err = fmt.Sscan(tokens[5], &soa.Expire)
		}
		if err == nil {
			_, err = fmt.Sscan(tokens[6], &soa.Minimum)
		}
	}
	return
}

func (v SOA) String() string {
	return fmt.Sprintf("%s\n%s\n%d\n%d\n%d\n%d\n%d",
		v.MName, v.RName,
		v.Serial, v.Refresh, v.Retry, v.Expire, v.Minimum)
}

func (SOA) Type() Type { return TypeSOA }

type SRV struct {
	Priority, Weight, Port uint16

	Target string
}

func parseSRV(tokens []string) (srv SRV, err error) {
	if len(tokens) < 4 {
		err = ErrIncomplete
	}
	if err == nil {
		_, err = fmt.Sscan(tokens[0], &srv.Priority)
	}
	if err == nil {
		_, err = fmt.Sscan(tokens[1], &srv.Weight)
	}
	if err == nil {
		_, err = fmt.Sscan(tokens[2], &srv.Port)
	}
	if err == nil {
		srv.Target = strings.Clone(tokens[3])
	}
	return
}

func (v SRV) String() string {
	return fmt.Sprintf("%d %d %d %s",
		v.Priority, v.Weight, v.Port, v.Target)
}

func (SRV) Type() Type { return TypeSRV }

type TXT []string

func parseTXT(tokens []string) (TXT, error) {
	ss, err := parseStrings(tokens)
	return TXT(ss), err
}

func (v TXT) String() string { return strings.Join(v, " ") }

func (TXT) Type() Type { return TypeTXT }

func parseAddr(tokens []string) (netip.Addr, error) {
	if len(tokens) == 0 {
		return netip.Addr{}, ErrIncomplete
	}
	return netip.ParseAddr(tokens[0])
}

func parseString(tokens []string) (string, error) {
	if len(tokens) == 0 {
		return "", ErrIncomplete
	}
	return strings.Clone(tokens[0]), nil
}

func parseStrings(tokens []string) ([]string, error) {
	if len(tokens) == 0 {
		return []string{}, ErrIncomplete
	}
	clones := make([]string, len(tokens))
	for i, s := range tokens {
		clones[i] = strings.Clone(s)
	}
	return clones, nil
}

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
	switch Type(r.Header.Type) {
	case TypeA:
		rb := r.Body.(*dnsmessage.AResource)
		s = net.IP(rb.A[:]).String()
	case TypeNS:
		s = r.Body.(*dnsmessage.NSResource).NS.String()
	case TypeCNAME:
		s = r.Body.(*dnsmessage.CNAMEResource).CNAME.String()
	case TypeSOA:
		rb := r.Body.(*dnsmessage.SOAResource)
		s = fmt.Sprintf("ns %v, mbox %v, s/n %d",
			rb.NS, rb.MBox, rb.Serial)
	case TypePTR:
		s = r.Body.(*dnsmessage.PTRResource).PTR.String()
	case TypeMX:
		rb := r.Body.(*dnsmessage.MXResource)
		s = fmt.Sprintf("%v, pref %d", rb.MX, rb.Pref)
	case TypeTXT:
		s = strings.Join(r.Body.(*dnsmessage.TXTResource).TXT, " ")
	case TypeAAAA:
		rb := r.Body.(*dnsmessage.AAAAResource)
		s = net.IP(rb.AAAA[:]).String()
	case TypeSRV:
		rb := r.Body.(*dnsmessage.SRVResource)
		s = fmt.Sprintf("%v, port %d, pri %d, weight %d",
			rb.Target, rb.Port, rb.Priority, rb.Weight)
	case TypeSVCB:
		var rb SVCBResource
		data := r.Body.(*dnsmessage.UnknownResource).Data
		if err := rb.UnmarshalBinary(data); err != nil {
			s = err.Error()
		} else {
			s = rb.String()
		}
	case TypeHTTPS:
		var rb HTTPSResource
		data := r.Body.(*dnsmessage.UnknownResource).Data
		if err := rb.UnmarshalBinary(data); err != nil {
			s = err.Error()
		} else {
			s = rb.String()
		}
	case TypeCAA:
		var rb CAAResource
		data := r.Body.(*dnsmessage.UnknownResource).Data
		if err := rb.UnmarshalBinary(data); err != nil {
			s = err.Error()
		} else {
			s = rb.String()
		}
	default:
		data := r.Body.(*dnsmessage.UnknownResource).Data
		s = fmt.Sprintf("%#x", data)
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
