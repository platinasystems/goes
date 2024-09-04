// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import "unsafe"

// https://en.wikipedia.org/wiki/ICMPv6
type ICMP6 struct {
	Type uint8
	Code uint8
	Sum  uint16
}

const SizeofICMP6 = int(unsafe.Sizeof(ICMP6{}))
