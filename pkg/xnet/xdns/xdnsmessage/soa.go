// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import "fmt"

type SOA struct {
	MName, RName UniqueString

	Serial, Refresh, Retry, Expire, Minimum uint32
}

func ParseSOA(tokens []string) (SOA, []string, error) {
	var soa SOA
	if len(tokens) < 7 {
		return soa, tokens, ErrIncomplete
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
	return soa, tokens[7:], err
}

func (v SOA) String() string {
	return fmt.Sprint(v.MName, "\n",
		v.RName.String(), "\n",
		v.Serial, "\n",
		v.Refresh, "\n",
		v.Retry, "\n",
		v.Expire, "\n",
		v.Minimum)
}
