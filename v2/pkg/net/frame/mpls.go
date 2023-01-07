// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"unsafe"
)

type MPLSData struct {
	*MPLS
	Data []byte
}

func NewMPLSData(data []byte) MPLSData {
	mpls := (*MPLS)(unsafe.Pointer(&data[0]))
	return MPLSData{mpls, data[unsafe.Sizeof(mpls):]}
}

func (mpls MPLSData) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "mpls ...")
}

type MPLS struct {
	// FIXME
}
