// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

type MX struct {
	Preference uint16
	Exchange   UniqueString
}

func ParseMX(tokens []string) (MX, []string, error) {
	var mx MX
	if len(tokens) < 2 {
		return mx, tokens, xerrors.Incomplete("MX")
	}
	_, err := fmt.Sscan(tokens[0], &mx.Preference)
	if err == nil {
		mx.Exchange = MakeUniqueString(tokens[1])
		tokens = tokens[2:]
	}
	return mx, tokens, err
}

func (v MX) String() string {
	return fmt.Sprint(v.Preference, " ", v.Exchange)
}
