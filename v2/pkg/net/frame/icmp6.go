// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"unsafe"
)

type ICMP6Data struct {
	*ICMP6
	Data []byte
}

func NewICMP6Data(data []byte) ICMP6Data {
	icmp6 := (*ICMP6)(unsafe.Pointer(&data[0]))
	return ICMP6Data{icmp6, data[unsafe.Sizeof(icmp6):]}
}

func (icmp6 ICMP6Data) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "icmp6 ...")
}

type ICMP6 struct {
	// FIXME
}
