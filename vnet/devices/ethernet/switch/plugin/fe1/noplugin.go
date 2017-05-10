// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build noplugin

package fe1

import (
	"github.com/platinasystems/fe1"
	firmware "github.com/platinasystems/firmware-fe1a"
)

func Packages() []map[string]string {
	return []map[string]string{fe1.Package, firmware.Package}
}
