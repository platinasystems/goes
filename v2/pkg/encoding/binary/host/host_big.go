// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build armbe || arm64be || m68k || mips || mips64 || mips64p32 || ppc || ppc64 || s390 || s390x || shbe || sparc || sparc64

package host

import "github.com/platinasystems/goes/v2/pkg/encoding/binary/big"

const (
	IsBigEndian    = true
	IsLittleEndian = false
)

type Uint16 = big.Uint16
type Uint32 = big.Uint32
type Uint64 = big.Uint64
type Float32 = big.Float32
type Float64 = big.Float64

func Net16(v uint16) uint16 { return v }
func Net32(v uint32) uint32 { return v }
func Net64(v uint64) uint64 { return v }
