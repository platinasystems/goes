// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package httpx

import (
	"fmt"
	"strings"
)

type Authority []string

func NewAuthority(network, addr string) Authority {
	s := "https://unix"
	if network != "unix" {
		s = fmt.Sprint("https://", addr)
	}
	return Authority{s}
}

func (a Authority) Join(elements ...string) string {
	if len(elements) == 0 {
		elements = []string{""}
	}
	return strings.Join([]string(append(a, elements...)), "/")
}
