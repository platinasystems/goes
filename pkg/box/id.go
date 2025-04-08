// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package box

import (
	"fmt"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/xnet"
)

type Id uint32

var (
	DecodeId = xnet.Decode32[Id]
	EncodeId = xnet.Encode32[Id]
)

const (
	SizeofId = int(unsafe.Sizeof(Id(0)))

	IdBits = SizeofId * 8

	IdVersionBit = IdBits - 4
	IdIndexMask  = (1 << IdVersionBit) - 1

	InvalidId Id = 1<<IdBits - 1

	NextIdVersion = 1 << IdVersionBit
)

func ParseId(s string) (id Id, err error) {
	_, err = fmt.Sscan(s, &id)
	return
}

func (id *Id) BumpVersion() {
	*id += NextIdVersion
}

func (id Id) Encode(buf []byte) {
	EncodeId(buf, id)
}

func (id Id) Index() int {
	return int(id & IdIndexMask)
}

func (id Id) Version() uint8 {
	return uint8(id >> IdVersionBit)
}
