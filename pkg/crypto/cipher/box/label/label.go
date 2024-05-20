// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package label

import (
	"encoding/binary"
	"fmt"
)

const (
	Size = 4

	VersionBit = 28
	IndexMask  = (1 << VersionBit) - 1

	Mislabel Label = 1<<32 - 1
)

type Label uint32
type Index uint32
type Version uint8

var ByteOrder = binary.BigEndian

func Parse(s string) (lbl Label, err error) {
	_, err = fmt.Sscan(s, &lbl)
	return
}

func With(data []byte) Label {
	return Label(ByteOrder.Uint32(data))
}

func (lbl *Label) Bump() {
	*lbl += 1 << VersionBit
}

func (lbl Label) Index() Index {
	return Index(lbl & IndexMask)
}

func (lbl Label) Put(data []byte) {
	ByteOrder.PutUint32(data, uint32(lbl))
}

func (lbl Label) String() string {
	return fmt.Sprintf("%#x", uint32(lbl))
}

func (lbl Label) Version() Version {
	return Version(lbl >> VersionBit)
}

func (i Index) String() string {
	return fmt.Sprintf("%d", uint32(i))
}

func (v Version) String() string {
	return fmt.Sprintf("%d", uint8(v))
}
