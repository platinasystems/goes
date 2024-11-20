// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"golang.org/x/net/dns/dnsmessage"
)

type TypeMINFOResource struct {
	RMAILBX, EMAILBX UniqueString
}

func ParseMINFO(tokens []string) (TypeMINFOResource, error) {
	if len(tokens) < 2 {
		return TypeMINFOResource{}, xerrors.Incomplete("MINFO")
	}
	return TypeMINFOResource{
		RMAILBX: MakeUniqueString(tokens[0]),
		EMAILBX: MakeUniqueString(tokens[1]),
	}, nil
}

func (v TypeMINFOResource) construct(
	mb *dnsmessage.Builder, h dnsmessage.ResourceHeader,
) error {
	r := dnsmessage.UnknownResource{
		Type: h.Type,
	}
	r.Data = v.RMAILBX.AppendTo(r.Data)
	r.Data = v.EMAILBX.AppendTo(r.Data)
	return mb.UnknownResource(h, r)
}

func (v TypeMINFOResource) String() string {
	return fmt.Sprint(v.RMAILBX, " ", v.EMAILBX)
}
