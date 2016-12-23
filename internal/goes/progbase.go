// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import "path/filepath"

var progbase string

func ProgBase() string {
	if len(progbase) == 0 {
		progbase = filepath.Base(Prog())
	}
	return progbase
}
