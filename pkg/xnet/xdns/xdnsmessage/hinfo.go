// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"golang.org/x/net/dns/dnsmessage"
)

type TypeHINFOResource struct {
	CPU, OS UniqueString
}

func ParseHINFO(tokens []string) (TypeHINFOResource, error) {
	if len(tokens) < 2 {
		return TypeHINFOResource{}, xerrors.Incomplete("HINFO")
	}
	return TypeHINFOResource{
		CPU: MakeUniqueString(tokens[0]),
		OS:  MakeUniqueString(tokens[1]),
	}, nil
}

func (v TypeHINFOResource) construct(
	mb *dnsmessage.Builder, h dnsmessage.ResourceHeader,
) error {
	r := dnsmessage.UnknownResource{
		Type: h.Type,
	}
	r.Data = v.CPU.AppendTo(r.Data)
	r.Data = v.OS.AppendTo(r.Data)
	return mb.UnknownResource(h, r)
}

func (v TypeHINFOResource) String() string {
	return fmt.Sprint(v.CPU, " ", v.OS)
}
