// Copyright © 2017-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package magic

import (
	"github.com/platinasystems/go/internal/magic/ext2"
	"github.com/platinasystems/go/internal/magic/ext3"
	"github.com/platinasystems/go/internal/magic/ext4"
	"github.com/platinasystems/go/internal/magic/iso9660"
	"github.com/platinasystems/go/internal/magic/vfat"
)

const (
	mbrMediaTypeOff = 0x15
	mbrMagicOffL    = 0x1fe
	mbrMagicOffM    = 0x1ff
	mbrMagicValL    = 0x55
	mbrMagicValM    = 0xaa
)

func isFatValidMedia(sniff []byte) bool {
	return 0xf8 <= sniff[mbrMediaTypeOff] ||
		sniff[mbrMediaTypeOff] == 0xf0
}

func IdentifyPartitionMap(sniff []byte) string {
	if sniff[mbrMagicOffL] == mbrMagicValL &&
		sniff[mbrMagicOffM] == mbrMagicValM &&
		!isFatValidMedia(sniff) &&
		IdentifyPartition(sniff) == "" {
		return "mbr"
	}
	return ""
}

func IdentifyPartition(sniff []byte) string {
	if ext2.Probe(sniff) {
		return "ext2"
	}
	if ext3.Probe(sniff) {
		return "ext3"
	}
	if ext4.Probe(sniff) {
		return "ext4"
	}
	if vfat.Probe(sniff) {
		return "vfat"
	}
	if iso9660.Probe(sniff) {
		return "iso9660"
	}
	return ""
}

func IdentifyFile(sniff []byte) string {
	return ""
}
