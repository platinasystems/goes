// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"strings"
	"unique"
	_ "unsafe"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"golang.org/x/net/dns/dnsmessage"
)

// FIXME replace with local name parser
//
//go:linkname unpackCNAMEResource golang.org/x/net/dns/dnsmessage.unpackCNAMEResource
func unpackCNAMEResource([]byte, int) (dnsmessage.CNAMEResource, error)

type UniqueString struct {
	unique.Handle[string]
}

var UniqueEmptyString = MakeUniqueString("")

func MakeUniqueString(s string) UniqueString {
	return UniqueString{unique.Make(s)}
}

func (v UniqueString) AppendTo(buf []byte) []byte {
	s := v.String()
	if !strings.HasSuffix(s, ".") {
		s += "."
	}
	for _, ss := range strings.Split(s, ".") {
		n := len(ss)
		if n > 255 {
			n = 0
		}
		buf = append(buf, uint8(n))
		if n == 0 {
			break
		}
		buf = append(buf, ss...)
	}
	return buf
}

// If successful, return unique cloned string and remaining data.
func (p *UniqueString) Decode(data []byte) ([]byte, error) {
	r, err := unpackCNAMEResource(data, 0)
	if err != nil {
		return data, err
	}
	p.Handle = unique.Make(r.CNAME.String())
	return data[r.CNAME.Length:], nil
}

func (v UniqueString) MarshalBinary() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v UniqueString) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v UniqueString) rename(name *dnsmessage.Name) {
	name.Length = uint8(copy(name.Data[:], v.String()))
}

func (v UniqueString) String() string {
	return v.Value()
}

type TypeCNAMEResource struct{ UniqueString }
type TypeNSResource struct{ UniqueString }
type TypePTRResource struct{ UniqueString }

type StringResources interface {
	TypeCNAMEResource | TypeNSResource | TypePTRResource
}

func ParseString[T StringResources](tokens []string) (T, error) {
	if len(tokens) == 0 {
		return T{UniqueEmptyString}, xerrors.Incomplete("CNAME|NS|PTR")
	}
	return T{MakeUniqueString(tokens[0])}, nil
}

func NewPTR(s string) TypePTRResource {
	return TypePTRResource{MakeUniqueString(s)}
}

func (v TypeCNAMEResource) construct(
	mb *dnsmessage.Builder, h dnsmessage.ResourceHeader,
) error {
	var cname dnsmessage.CNAMEResource
	v.rename(&cname.CNAME)
	return mb.CNAMEResource(h, cname)
}

func (v TypeNSResource) construct(
	mb *dnsmessage.Builder, h dnsmessage.ResourceHeader,
) error {
	var ns dnsmessage.NSResource
	v.rename(&ns.NS)
	return mb.NSResource(h, ns)
}

func (v TypePTRResource) construct(
	mb *dnsmessage.Builder, h dnsmessage.ResourceHeader,
) error {
	var ptr dnsmessage.PTRResource
	v.rename(&ptr.PTR)
	return mb.PTRResource(h, ptr)
}
