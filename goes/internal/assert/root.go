// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package assert

import (
	"errors"
	"os"
)

func Root() (err error) {
	if os.Geteuid() != 0 {
		err = errors.New("you aren't root")
	}
	return
}
