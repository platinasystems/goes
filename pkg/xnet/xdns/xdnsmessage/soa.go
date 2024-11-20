// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"golang.org/x/net/dns/dnsmessage"
)

type TypeSOAResource struct {
	MName, RName UniqueString

	Serial, Refresh, Retry, Expire, Minimum uint32
}

func ParseSOA(tokens []string) (TypeSOAResource, error) {
	var soa TypeSOAResource
	if len(tokens) < 7 {
		return soa, xerrors.Incomplete("SOA")
	}
	soa.MName = MakeUniqueString(tokens[0])
	soa.RName = MakeUniqueString(tokens[1])
	_, err := fmt.Sscan(tokens[2], &soa.Serial)
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
	return soa, err
}

func (v TypeSOAResource) construct(
	mb *dnsmessage.Builder, h dnsmessage.ResourceHeader,
) error {
	var soa dnsmessage.SOAResource

	v.MName.rename(&soa.MBox)
	v.RName.rename(&soa.NS)
	soa.Serial = v.Serial
	soa.Refresh = v.Refresh
	soa.Retry = v.Retry
	soa.Expire = v.Expire
	soa.MinTTL = v.Minimum
	return mb.SOAResource(h, soa)
}

func (v TypeSOAResource) String() string {
	return fmt.Sprint(v.MName, "\n",
		v.RName.String(), "\n",
		v.Serial, "\n",
		v.Refresh, "\n",
		v.Retry, "\n",
		v.Expire, "\n",
		v.Minimum)
}
