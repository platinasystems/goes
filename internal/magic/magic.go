// Copyright © 2017-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package magic

import (
	"github.com/platinasystems/go/internal/magic/ext2"
	"github.com/platinasystems/go/internal/magic/ext3"
	"github.com/platinasystems/go/internal/magic/ext4"
	"github.com/platinasystems/go/internal/magic/iso9660"
	"github.com/platinasystems/go/internal/magic/mbr"
	"github.com/platinasystems/go/internal/magic/vfat"
)

func IdentifyPartitionMap(sniff []byte) string {
	if mbr.Probe(sniff) && IdentifyPartition(sniff) == "" {
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
