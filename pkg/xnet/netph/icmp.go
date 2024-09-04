// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import "unsafe"

// https://en.wikipedia.org/wiki/Internet_Control_Message_Protocol
type ICMP struct {
	Type uint8
	Code uint8
	Sum  uint16
}

const SizeofICMP = int(unsafe.Sizeof(ICMP{}))
