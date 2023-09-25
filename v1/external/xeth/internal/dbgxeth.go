// Copyright © 2018-2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build dbgxeth

package internal

import "fmt"

func (kind MsgKind) String() string {
	s, found := map[MsgKind]string{
		MsgKindBreak:                         "break",
		MsgKindLinkStat:                      "link-stat",
		MsgKindEthtoolStat:                   "ethtool-stat",
		MsgKindEthtoolFlags:                  "ethtool-flags",
		MsgKindEthtoolSettings:               "ethtool-settings",
		MsgKindEthtoolLinkModesSupported:     "supported-link-modes",
		MsgKindEthtoolLinkModesAdvertising:   "advertising-link-modes",
		MsgKindEthtoolLinkModesLPAdvertising: "link-partner-advertising-link-modes",
		MsgKindDumpIfInfo:                    "dump-IfInfo",
		MsgKindCarrier:                       "carrier",
		MsgKindSpeed:                         "speed",
		MsgKindIfInfo:                        "ifInfo",
		MsgKindIfa:                           "ifa",
		MsgKindIfa6:                          "ifa6",
		MsgKindDumpFibInfo:                   "dump-fib-info",
		MsgKindFibEntry:                      "fib-entry",
		MsgKindNeighUpdate:                   "neighbor-update",
		MsgKindChangeUpperXid:                "change-upper",
	}[kind]
	if !found {
		s = fmt.Sprint(uint8(kind))
	}
	return s
}
