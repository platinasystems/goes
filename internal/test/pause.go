// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"bufio"
	"fmt"
	"os"
)

func Pause(args ...interface{}) {
	if len(args) == 0 {
		args = []interface{}{"Press enter to continue..."}
	}
	fmt.Print(args...)
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
