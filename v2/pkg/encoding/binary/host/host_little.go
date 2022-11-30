// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build 386 || amd64 || amd64p32 || alpha || arm || arm64 || loong64 || mipsle || mips64le || mips64p32le || nios2 || ppc64le || riscv || riscv64 || sh

package host

import "github.com/platinasystems/goes/v2/pkg/encoding/binary/little"

type Uint8 = little.Uint8
type Uint16 = little.Uint16
type Uint32 = little.Uint32
type Uint64 = little.Uint64
type Float32 = little.Float32
type Float64 = little.Float64
