// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"bufio"
	"os"
)

func Pause() {
	os.Stdout.WriteString("Pause, hit return to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
