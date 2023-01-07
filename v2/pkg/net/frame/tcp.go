// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"unsafe"
)

type TCPData struct {
	*TCP
	Data []byte
}

func NewTCPData(data []byte) TCPData {
	tcp := (*TCP)(unsafe.Pointer(&data[0]))
	return TCPData{tcp, data[unsafe.Sizeof(tcp):]}
}

func (tcp TCPData) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "tcp ...")
}

type TCP struct {
	// FIXME
}
