// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package ndp

import "golang.org/x/sys/unix"

func ndpFlags(f uint) (s string) {
	if (f & unix.NTF_ROUTER) != 0 {
		s = "R"
	}
	if (f & unix.NTF_PROXY) != 0 {
		s += "p"
	}
	return
}

func ndpState(st int) string {
	s, ok := map[int]string{
		unix.NUD_INCOMPLETE: "I",
		unix.NUD_REACHABLE:  "R",
		unix.NUD_STALE:      "S",
		unix.NUD_DELAY:      "D",
		unix.NUD_PROBE:      "P",
		unix.NUD_FAILED:     "F",
	}[st]
	if !ok {
		s = "N"
	}
	return s
}
