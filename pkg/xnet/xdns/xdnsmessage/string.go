// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
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

// If successful, return unique cloned string and remaining data.
func (p *UniqueString) Decode(data []byte) ([]byte, error) {
	r, err := unpackCNAMEResource(data, 0)
	if err != nil {
		return data, err
	}
	p.Handle = unique.Make(r.CNAME.String())
	return data[r.CNAME.Length:], nil
}

func (v UniqueString) String() string {
	return v.Value()
}

type CNAME struct{ UniqueString }
type NS struct{ UniqueString }
type PTR struct{ UniqueString }

func ParseString[T CNAME | NS | PTR](tokens []string) (
	T, []string, error,
) {
	if len(tokens) == 0 {
		return T{UniqueEmptyString}, tokens, xerrors.
			Incomplete("CNAME|NS|PTR")
	}
	return T{MakeUniqueString(tokens[0])}, tokens[1:], nil
}
