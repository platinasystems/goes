// Copyright © 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package magic

import (
	"io/ioutil"
	"testing"
)

func testFile(t *testing.T, f, pmType, pType, fType string) {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		t.Error("error reading reference file", f, err)
	} else {
		pmt := IdentifyPartitionMap(b)
		if pmt != pmType {
			t.Error("Partition map expected ", pmType, " got ", pmt)
		}
		pt := IdentifyPartition(b)
		if pt != pType {
			t.Error("Partition type expected ", pType, " got", pt)
		}
		ft := IdentifyFile(b)
		if ft != fType {
			t.Error("File type expected ", fType, " got ", ft)
		}
	}
}

func TestExt2(t *testing.T) {
	testFile(t, "ext2-sb.dat", "", "ext2", "")
}

func TestExt3(t *testing.T) {
	testFile(t, "ext3-sb.dat", "", "ext3", "")
}

func TestExt4(t *testing.T) {
	testFile(t, "ext4-sb.dat", "", "ext4", "")
}
