// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"unsafe"
)

type UDPData struct {
	*UDP
	Data []byte
}

func NewUDPData(data []byte) UDPData {
	udp := (*UDP)(unsafe.Pointer(&data[0]))
	return UDPData{udp, data[unsafe.Sizeof(udp):]}
}

func (udp UDPData) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "udp ...")
}

type UDP struct {
	// FIXME
}
