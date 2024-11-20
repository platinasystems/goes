// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"golang.org/x/net/dns/dnsmessage"
)

type TypeSRVResource struct {
	Priority, Weight, Port uint16

	Target UniqueString
}

func ParseSRV(tokens []string) (TypeSRVResource, error) {
	var srv TypeSRVResource
	if len(tokens) < 4 {
		return srv, xerrors.Incomplete("SRV")
	}
	_, err := fmt.Sscan(tokens[0], &srv.Priority)
	if err == nil {
		_, err = fmt.Sscan(tokens[1], &srv.Weight)
	}
	if err == nil {
		_, err = fmt.Sscan(tokens[2], &srv.Port)
	}
	if err == nil {
		srv.Target = MakeUniqueString(tokens[3])
	}
	return srv, err
}

func (v TypeSRVResource) construct(
	mb *dnsmessage.Builder, h dnsmessage.ResourceHeader,
) error {
	srv := dnsmessage.SRVResource{
		Priority: v.Priority,
		Weight:   v.Weight,
		Port:     v.Port,
	}
	v.Target.rename(&srv.Target)
	return mb.SRVResource(h, srv)
}

func (v TypeSRVResource) String() string {
	return fmt.Sprint(v.Priority, " ", v.Weight, " ", v.Port, " ", v.Target)
}
