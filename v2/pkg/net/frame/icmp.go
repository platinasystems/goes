// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"unsafe"
)

type ICMPData struct {
	*ICMP
	Data []byte
}

func NewICMPData(data []byte) ICMPData {
	icmp := (*ICMP)(unsafe.Pointer(&data[0]))
	return ICMPData{icmp, data[unsafe.Sizeof(icmp):]}
}

func (icmp *ICMP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "icmp ...")
}

type ICMP struct {
	// FIXME
}
