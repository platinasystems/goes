// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"golang.org/x/net/dns/dnsmessage"
)

type TypeTXTResource []string

func ParseTXT(tokens []string) (TypeTXTResource, error) {
	if len(tokens) == 0 {
		return TypeTXTResource{}, xerrors.Incomplete("TXT")
	}
	txt := make(TypeTXTResource, len(tokens))
	for i, s := range tokens {
		txt[i] = strings.Clone(s)
	}
	return txt, nil
}

func (v TypeTXTResource) construct(
	mb *dnsmessage.Builder, h dnsmessage.ResourceHeader,
) error {
	return mb.TXTResource(h, dnsmessage.TXTResource{
		TXT: []string(v),
	})
}

func (v TypeTXTResource) String() string {
	return strings.Join(v, " ")
}
