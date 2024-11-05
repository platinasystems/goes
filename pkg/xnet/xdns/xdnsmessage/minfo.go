// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

type MINFO struct {
	RMAILBX, EMAILBX string
}

func ParseMINFO(tokens []string) (MINFO, []string, error) {
	if len(tokens) < 2 {
		return MINFO{}, tokens, xerrors.Incomplete("MINFO")
	}
	return MINFO{
		RMAILBX: strings.Clone(tokens[0]),
		EMAILBX: strings.Clone(tokens[1]),
	}, tokens[2:], nil
}

func (v MINFO) String() string {
	return v.RMAILBX + " " + v.EMAILBX
}
