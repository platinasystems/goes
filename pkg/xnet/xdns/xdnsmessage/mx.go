// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"golang.org/x/net/dns/dnsmessage"
)

type TypeMXResource struct {
	Preference uint16
	Exchange   UniqueString
}

func ParseMX(tokens []string) (TypeMXResource, error) {
	var mx TypeMXResource
	if len(tokens) < 2 {
		return mx, xerrors.Incomplete("MX")
	}
	_, err := fmt.Sscan(tokens[0], &mx.Preference)
	if err == nil {
		mx.Exchange = MakeUniqueString(tokens[1])
	}
	return mx, err
}

func (v TypeMXResource) construct(
	mb *dnsmessage.Builder, h dnsmessage.ResourceHeader,
) error {
	var mx dnsmessage.MXResource
	v.Exchange.rename(&mx.MX)
	mx.Pref = v.Preference
	return mb.MXResource(h, mx)
}

func (v TypeMXResource) String() string {
	return fmt.Sprint(v.Preference, " ", v.Exchange)
}
