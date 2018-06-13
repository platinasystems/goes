// Copyright © 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ext2

import (
	"github.com/platinasystems/go/internal/magic/ext"
)

func Probe(s []byte) bool {
	if !ext.Probe(s) {
		return false
	}

	compat, incompat, rocompat := ext.Compat(s)

	if (compat & ext.FeatureCompatExt3HasJournal) != 0 {
		return false
	}

	if (incompat & ext.FeatureIncompatExt2Unsupp) != 0 {
		return false
	}

	if (rocompat & ext.FeatureRoCompatExt2Unsupp) != 0 {
		return false
	}
	return true
}
