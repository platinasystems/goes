// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"fmt"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"golang.org/x/net/dns/dnsmessage"
)

type OPT []dnsmessage.Option

func ParseOPT(tokens []string) (OPT, []string, error) {
	return OPT{}, tokens, xerrors.FIXME("opt")
}

func (v OPT) String() string {
	switch len(v) {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf("code=%d, data=%#x", v[0].Code, v[0].Data)
	}
	w := new(strings.Builder)
	for _, opt := range v {
		fmt.Fprintf(w, "code=%d, data=%#x\n", opt.Code, opt.Data)
	}
	return w.String()
}
