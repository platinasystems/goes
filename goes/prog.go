// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import "os"

var prog string

func Prog() string {
	if len(prog) == 0 {
		var err error
		prog, err = os.Readlink("/proc/self/exe")
		if err != nil {
			prog = InstallName
		}
	}
	return prog
}
