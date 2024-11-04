// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	_ "embed"

	"github.com/platinasystems/goes/v2/pkg/integer"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

//go:embed type_help.txt
var TypeHelpTxt string

//go:generate stringer -type Type -trimprefix Type
type Type uint16

// https://en.wikipedia.org/wiki/List_of_DNS_record_types
const (
	Type0 Type = iota
	TypeA
	TypeNS
	TypeMD
	TypeMF
	TypeCNAME
	TypeSOA
	TypeMB
	TypeMG
	TypeMR
	TypeNULL
	TypeWKS
	TypePTR
	TypeHINFO
	TypeMINFO
	TypeMX
	TypeTXT
	TypeRP
	TypeAFSDB
	TypeX25
	TypeISDN
	TypeRT
	TypeNSAP
	TypeNSAPPTR
	TypeSIG
	TypeKEY
	TypePX
	TypeGPOS
	TypeAAAA
	TypeLOC
	TypeNXT
	TypeEID
	TypeNB
	TypeSRV
	TypeATMA
	TypeNAPTR
	TypeKX
	TypeCERT
	TypeA6
	TypeDNAME
	TypeSINK
	TypeOPT
	TypeAPL
	TypeDS
	TypeSSHFP
	TypeIPSECKEY
	TypeRRSIG
	TypeNSEC
	TypeDNSKEY
	TypeDHCID
	TypeNSEC3
	TypeNSEC3PARAM
	TypeTLSA
	TypeSMIMEA
	Type54
	TypeHIP
	TypeNINFO
	TypeRKEY
	TypeTALINK
	TypeCDS
	TypeCDNSKEY
	TypeOPENPGPKEY
	TypeCSYNC
	TypeZONEMD
	TypeSVCB
	TypeHTTPS
)

const (
	TypeEUI48 Type = 108 + iota
	TypeEUI64
)

const (
	TypeTKEY Type = 249 + iota
	TypeTSIG
	TypeIXFR
	TypeAXFR
	TypeMAILB
	TypeMAILA
	TypeANY
	TypeURI
	TypeCAA
)

func (t Type) ParseResource(tokens []string) (Resource, []string, error) {
	switch t {
	case TypeA:
		return ParseAddr[A](tokens)
	case TypeNS:
		return ParseString[NS](tokens)
	case TypeCNAME:
		return ParseString[CNAME](tokens)
	case TypeSOA:
		return ParseSOA(tokens)
	case TypePTR:
		return ParseString[PTR](tokens)
	case TypeHINFO:
		return ParseHINFO(tokens)
	case TypeMINFO:
		return ParseMINFO(tokens)
	case TypeMX:
		return ParseMX(tokens)
	case TypeTXT:
		return ParseTXT(tokens)
	case TypeAAAA:
		return ParseAddr[AAAA](tokens)
	case TypeLOC:
		return ParseLOC(tokens)
	case TypeSRV:
		return ParseSRV(tokens)
	}
	return nil, tokens, xerrors.Unsupported(t)
}

func (A) Type() Type     { return TypeA }
func (NS) Type() Type    { return TypeNS }
func (CNAME) Type() Type { return TypeCNAME }
func (SOA) Type() Type   { return TypeSOA }
func (PTR) Type() Type   { return TypePTR }
func (HINFO) Type() Type { return TypeHINFO }
func (MINFO) Type() Type { return TypeMINFO }
func (MX) Type() Type    { return TypeMX }
func (TXT) Type() Type   { return TypeTXT }
func (AAAA) Type() Type  { return TypeAAAA }
func (LOC) Type() Type   { return TypeLOC }
func (SRV) Type() Type   { return TypeSRV }
func (SVCB) Type() Type  { return TypeSVCB }
func (HTTPS) Type() Type { return TypeHTTPS }
func (CAA) Type() Type   { return TypeCAA }

func TypeNamed(name string) (Type, error) {
	v, found := integer.Named(Type0, name, _Type_name_0,
		_Type_index_0[:]...)
	if found {
		return v, nil
	}
	v, found = integer.Named(TypeEUI48, name, _Type_name_1,
		_Type_index_1[:]...)
	if found {
		return v, nil
	}
	v, found = integer.Named(TypeTKEY, name, _Type_name_2,
		_Type_index_2[:]...)
	if found {
		return v, nil
	}
	return Type0, ErrInvalid
}

func (v Type) MarshalText() (b []byte, _ error) {
	if v != Type0 {
		b = []byte(v.String())
	}
	return
}

func (p *Type) UnmarshalText(text []byte) (err error) {
	*p, err = TypeNamed(string(text))
	return
}
