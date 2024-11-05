// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

type TXT []string

func ParseTXT(tokens []string) (TXT, []string, error) {
	if len(tokens) == 0 {
		return TXT{}, tokens, xerrors.Incomplete("TXT")
	}
	txt := make(TXT, len(tokens))
	for i, s := range tokens {
		txt[i] = strings.Clone(s)
	}
	return txt, tokens[len(tokens):], nil
}

func (v TXT) String() string {
	return strings.Join(v, " ")
}
