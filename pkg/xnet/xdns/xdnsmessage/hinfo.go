// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import "strings"

type HINFO struct {
	CPU, OS string
}

func ParseHINFO(tokens []string) (HINFO, []string, error) {
	if len(tokens) < 2 {
		return HINFO{}, tokens, ErrIncomplete
	}
	return HINFO{
		CPU: strings.Clone(tokens[0]),
		OS:  strings.Clone(tokens[1]),
	}, tokens[2:], nil
}

func (v HINFO) String() string {
	return v.CPU + " " + v.OS
}
