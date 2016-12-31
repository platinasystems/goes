// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package prog provides methods that return the current program base and full
// name along with it's minimal PATH. Each of these have cached results.
package prog

import (
	"os"
	"path/filepath"
)

var base, name, path string

func Base() string {
	if len(base) == 0 {
		base = filepath.Base(Name())
	}
	return base
}

func Name() string {
	if len(name) == 0 {
		var err error
		name, err = os.Readlink("/proc/self/exe")
		if err != nil {
			name = os.Args[0]
		}
	}
	return name
}

func Path() string {
	if len(path) == 0 {
		path = "/bin:/usr/bin"
		dir := filepath.Dir(Name())
		if dir != "/bin" && dir != "/usr/bin" {
			path += ":" + dir
		}
	}
	return path
}
