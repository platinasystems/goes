// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"unsafe"
)

type MMPLSData struct {
	*MMPLS
	Data []byte
}

func NewMMPLSData(data []byte) MMPLSData {
	mmpls := (*MMPLS)(unsafe.Pointer(&data[0]))
	return MMPLSData{mmpls, data[unsafe.Sizeof(mmpls):]}
}

func (mmpls MMPLSData) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "m-mpls ...")
}

type MMPLS struct {
	// FIXME
}
