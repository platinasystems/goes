// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/box"
)

const (
	IdBits = box.SizeofId * 8

	IdVersionBit = IdBits - 4
	IdIndexMask  = (1 << IdVersionBit) - 1

	InvalidId box.Id = 1<<IdBits - 1

	NextIdVersion = 1 << IdVersionBit
)

func ParseId(s string) (id box.Id, err error) {
	_, err = fmt.Sscan(s, &id)
	return
}

func IdIndex(id box.Id) int {
	return int(id & IdIndexMask)
}

func IdVersion(id box.Id) uint8 {
	return uint8(id >> IdVersionBit)
}

func BumpIdVersion(id box.Id) box.Id {
	return id + NextIdVersion
}
