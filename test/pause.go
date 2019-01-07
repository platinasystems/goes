// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
)

func Pause(args ...interface{}) {
	if !*MustPause {
		return
	}
	if len(args) > 0 {
		fmt.Print(args...)
		fmt.Print("; ")
	}
	fmt.Print("press enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func PauseQuit(t *testing.T) {
	assert := Assert{t}
	fmt.Println("Press enter to continue, Q+enter to quit")
	buf, _ := bufio.NewReader(os.Stdin).ReadBytes('\n')
	if strings.ContainsAny(string(buf[:]), "Q") {
		assert.Fatal("quitting")
	}
}
