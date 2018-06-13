// Copyright © 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ext

import (
	"encoding/binary"
)

const (
	MagicOffL = 0x438
	MagicOffM = 0x439

	MagicValL = 0x53
	MagicValM = 0xef

	FeatureCompatOff            = 0x45c
	FeatureCompatExt3HasJournal = 0x4

	FeatureIncompatOff            = 0x460
	FeatureIncompatExt2Filetype   = 0x2
	FeatureIncompatExt3Recover    = 0x4
	FeatureIncompatExt3JournalDev = 0x8
	FeatureIncompatExt2MetaBg     = 0x10
	FeatureIncompatExt4Extents    = 0x40
	FeatureIncompatExt464Bit      = 0x80
	FeatureIncompatExt4MMP        = 0x100
	FeatureIncompatExt4FlexBg     = 0x200

	FeatureIncompatExt2Unsupp = (FeatureIncompatExt2Filetype | FeatureIncompatExt2MetaBg) ^ 0xffff
	FeatureIncompatExt3Unsupp = (FeatureIncompatExt2Filetype | FeatureIncompatExt3Recover | FeatureIncompatExt2MetaBg) ^ 0xffff

	FeatureRoCompatOff             = 0x464
	FeatureRoCompatExt2SparseSuper = 0x1
	FeatureRoCompatExt2LargeFile   = 0x2
	FeatureRoCompatExt2BtreeDir    = 0x4
	FeatureRoCompatExt4HugeFile    = 0x8
	FeatureRoCompatExt4GdtCsum     = 0x10
	FeatureRoCompatExt4DirNlink    = 0x20
	FeatureRoCompatExt4ExtraIsize  = 0x40

	FeatureRoCompatExt2Unsupp = (FeatureRoCompatExt2SparseSuper | FeatureRoCompatExt2LargeFile | FeatureRoCompatExt2BtreeDir) ^ 0xffff
	FeatureRoCompatExt3Unsupp = (FeatureRoCompatExt2SparseSuper | FeatureRoCompatExt2LargeFile | FeatureRoCompatExt2BtreeDir) ^ 0xffff
)

func Probe(s []byte) bool {
	if s[MagicOffL] == MagicValL && s[MagicOffM] == MagicValM {
		return true
	}
	return false
}

func Compat(s []byte) (compat uint32, incompat uint32, roCompat uint32) {
	compat = binary.LittleEndian.Uint32(s[FeatureCompatOff:])
	incompat = binary.LittleEndian.Uint32(s[FeatureIncompatOff:])
	roCompat = binary.LittleEndian.Uint32(s[FeatureRoCompatOff:])

	return
}
