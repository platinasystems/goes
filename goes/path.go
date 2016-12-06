// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import "path/filepath"

var path string

func Path() string {
	if len(path) == 0 {
		path = "/bin:/usr/bin"
		dir := filepath.Dir(Prog())
		if dir != "/bin" && dir != "/usr/bin" {
			path += ":" + dir
		}
	}
	return path
}
