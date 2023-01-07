// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"unsafe"
)

type HopByHopData struct {
	*HopByHop
	Data []byte
}

func NewHopByHopData(data []byte) HopByHopData {
	hop := (*HopByHop)(unsafe.Pointer(&data[0]))
	return HopByHopData{hop, data[unsafe.Sizeof(hop):]}
}

func (hop HopByHopData) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "hop-by-hop ...")
}

type HopByHop struct {
	// FIXME
}
