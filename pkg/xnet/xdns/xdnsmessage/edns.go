// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

// EDNS(0) wire constants.
const (
	EDNS0Version = 0

	EDNS0DNSSECOK     = 0x00008000
	EDNSVersionMask   = 0x00ff0000
	EDNS0DNSSECOKMask = 0x00ff8000
)

func EDNSVersion(ttl uint32) uint8 {
	return uint8(ttl >> 16)
}

func HasEDNS0DNSSECOK(ttl uint32) bool {
	return (ttl & EDNS0DNSSECOK) != 0
}

func EDNS0MBZ(ttl uint32) uint16 {
	return uint16(ttl & 0x7fff)
}
