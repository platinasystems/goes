// Copyright © 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ext4

import (
	"github.com/platinasystems/go/internal/magic/ext"
)

func Probe(s []byte) bool {
	if !ext.Probe(s) {
		return false
	}

	_, incompat, rocompat := ext.Compat(s)

	if ((incompat & ext.FeatureIncompatExt3Unsupp) == 0) &&
		((rocompat & ext.FeatureRoCompatExt3Unsupp) == 0) {
		return false
	}

	return true
}
