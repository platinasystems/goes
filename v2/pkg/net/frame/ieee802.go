// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"unsafe"
)

type IEEE8021QData struct {
	*IEEE8021Q
	Data []byte
}

func NewIEEE8021QData(data []byte) IEEE8021QData {
	q := (*IEEE8021Q)(unsafe.Pointer(&data[0]))
	return IEEE8021QData{q, data[unsafe.Sizeof(q):]}
}

func (q IEEE8021QData) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ieee802_1Q ...")
}

type IEEE8021ADData struct {
	*IEEE8021AD
	Data []byte
}

func NewIEEE8021ADData(data []byte) IEEE8021ADData {
	ad := (*IEEE8021AD)(unsafe.Pointer(&data[0]))
	return IEEE8021ADData{ad, data[unsafe.Sizeof(ad):]}
}

func (ad IEEE8021ADData) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ieee802_1ad ...")
}

type IEEE8021Q struct {
	// FIXME
}

type IEEE8021AD struct {
	// FIXME
}
