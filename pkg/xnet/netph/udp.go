// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

// https://en.wikipedia.org/wiki/User_Datagram_Protocol
type UDP struct {
	SP  uint16
	DP  uint16
	Len uint16
	Sum uint16
}

const UDPLenIndex = 2 + 2
const UDPSumIndex = 2 + 2 + 2
const UDPSize = 4 * 2

const (
	UDPEcho   = 7
	UDPDomain = 53
)
