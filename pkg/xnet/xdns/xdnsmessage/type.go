// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/integer"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"golang.org/x/net/dns/dnsmessage"
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

func (v Type) MarshalText() (b []byte, _ error) {
	if v != Type0 {
		b = []byte(v.String())
	}
	return
}

func (t Type) Parse(tokens []string) (TypedResource, error) {
	switch t {
	case TypeA:
		return ParseAddr[TypeAResource](tokens)
	case TypeNS:
		return ParseString[TypeNSResource](tokens)
	case TypeCNAME:
		return ParseString[TypeCNAMEResource](tokens)
	case TypeSOA:
		return ParseSOA(tokens)
	case TypePTR:
		return ParseString[TypePTRResource](tokens)
	case TypeHINFO:
		return ParseHINFO(tokens)
	case TypeMINFO:
		return ParseMINFO(tokens)
	case TypeMX:
		return ParseMX(tokens)
	case TypeTXT:
		return ParseTXT(tokens)
	case TypeAAAA:
		return ParseAddr[TypeAAAAResource](tokens)
	case TypeLOC:
		return ParseLOC(tokens)
	case TypeSRV:
		return ParseSRV(tokens)
	case TypeOPT:
		return ParseOPT(tokens)
	}
	buf := new(bytes.Buffer)
	for i, s := range tokens {
		if i > 0 {
			buf.WriteRune(' ')
			buf.WriteString(s)
		}
	}
	return TypeTBDResource{t, buf.Bytes()}, nil
}

func (p *Type) UnmarshalText(text []byte) (err error) {
	*p, err = TypeNamed(string(text))
	return
}

func (TypeAResource) Type() Type     { return TypeA }
func (TypeNSResource) Type() Type    { return TypeNS }
func (TypeCNAMEResource) Type() Type { return TypeCNAME }
func (TypeSOAResource) Type() Type   { return TypeSOA }
func (TypePTRResource) Type() Type   { return TypePTR }
func (TypeHINFOResource) Type() Type { return TypeHINFO }
func (TypeMINFOResource) Type() Type { return TypeMINFO }
func (TypeMXResource) Type() Type    { return TypeMX }
func (TypeTXTResource) Type() Type   { return TypeTXT }
func (TypeAAAAResource) Type() Type  { return TypeAAAA }
func (TypeLOCResource) Type() Type   { return TypeLOC }
func (TypeSRVResource) Type() Type   { return TypeSRV }
func (TypeOPTResource) Type() Type   { return TypeOPT }
func (TypeSVCBResource) Type() Type  { return TypeSVCB }
func (TypeHTTPSResource) Type() Type { return TypeHTTPS }
func (TypeCAAResource) Type() Type   { return TypeCAA }

type TypeTBDResource struct {
	t    Type
	Data []byte
}

func (v TypeTBDResource) construct(
	mb *dnsmessage.Builder, h dnsmessage.ResourceHeader,
) error {
	return mb.UnknownResource(h, dnsmessage.UnknownResource{
		Type: dnsmessage.Type(v.t),
		Data: v.Data,
	})
}

func (v TypeTBDResource) String() string {
	return fmt.Sprintf("%#x", v.Data)
}

func (v TypeTBDResource) Type() Type { return v.t }

func (p *TypeTBDResource) UnmarshalBinary(data []byte) error {
	p.Data = bytes.Clone(data)
	return nil
}

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
	return Type0, xerrors.Invalid(name)
}
