// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import "unsafe"

// https://en.wikipedia.org/wiki/Address_Resolution_Protocol
type ARP struct {
	HTYPE uint16
	PTYPE uint16
	HLEN  uint8
	PLEN  uint8
	OPER  uint16
	SHA   [6]byte
	SPA   [4]byte
	THA   [6]byte
	TPA   [4]byte
}

const SizeofARP = int(unsafe.Sizeof(ARP{}))
