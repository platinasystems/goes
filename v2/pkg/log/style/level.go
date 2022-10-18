// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package style

type Level uint8

const (
	Emergency Level = iota
	Alert
	Critical
	Errata
	Warning
	Notice
	Info
	Debug
)

func (lvl Level) String() (s string) {
	if lvl <= Debug {
		s = []string{
			"emergency",
			"alert",
			"critical",
			"errata",
			"warning",
			"notice",
			"info",
			"debug",
		}[lvl]
	}
	return
}
