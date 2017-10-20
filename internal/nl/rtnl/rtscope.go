// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"sort"
	"strings"
	"syscall"
)

const (
	RT_SCOPE_UNIVERSE uint8 = syscall.RT_SCOPE_UNIVERSE
	RT_SCOPE_SITE     uint8 = syscall.RT_SCOPE_SITE
	RT_SCOPE_LINK     uint8 = syscall.RT_SCOPE_LINK
	RT_SCOPE_HOST     uint8 = syscall.RT_SCOPE_HOST
	RT_SCOPE_NOWHERE  uint8 = syscall.RT_SCOPE_NOWHERE
)

var RtScopeByName = map[string]uint8{
	"global":   RT_SCOPE_UNIVERSE,
	"universe": RT_SCOPE_UNIVERSE,
	"site":     RT_SCOPE_SITE,
	"link":     RT_SCOPE_LINK,
	"host":     RT_SCOPE_HOST,
	"nowhere":  RT_SCOPE_NOWHERE,
}

var RtScopeName = map[uint8]string{
	RT_SCOPE_UNIVERSE: "global",
	RT_SCOPE_SITE:     "site",
	RT_SCOPE_LINK:     "link",
	RT_SCOPE_HOST:     "host",
	RT_SCOPE_NOWHERE:  "nowhere",
}

func CompleteRtScope(s string) (list []string) {
	for k := range RtScopeByName {
		if len(s) == 0 || strings.HasPrefix(k, s) {
			list = append(list, k)
		}
	}
	if len(list) > 0 {
		sort.Strings(list)
	}
	return
}
