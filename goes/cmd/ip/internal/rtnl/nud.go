// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import "strings"

const (
	NUD_INCOMPLETE uint16 = 1 << iota
	NUD_REACHABLE
	NUD_STALE
	NUD_DELAY
	NUD_PROBE
	NUD_FAILED
	NUD_NOARP
	NUD_PERMANENT
	N_NUD
)

const NUD_NONE uint16 = 0
const NUD_ALL = N_NUD - 1

var NudByName = map[string]uint16{
	"incomplete": NUD_INCOMPLETE,
	"reachable":  NUD_REACHABLE,
	"stale":      NUD_STALE,
	"delay":      NUD_DELAY,
	"probe":      NUD_PROBE,
	"failed":     NUD_FAILED,
	"noarp":      NUD_NOARP,
	"permanent":  NUD_PERMANENT,
}

func CompleteNud(s string) (list []string) {
	for k := range NudByName {
		if len(s) == 0 || strings.HasPrefix(k, s) {
			list = append(list, k)
		}
	}
	return
}
