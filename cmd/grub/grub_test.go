// Copyright © 2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package grub

import (
	"fmt"
	"os"
	"testing"
)

func TestGrubCommand(t *testing.T) {
	r, _ := os.Getwd()
	c := &Command{g: Goes}
	err := c.Main(r, "testdata/hello.grub")
	if err != ErrNoDefinedKernelOrMenus {
		if err != nil {
			t.Error("main failed:", err)
		} else {
			t.Error("main failed - no error")
		}
	}
}

func ExampleParser() {
	c := &Command{g: Goes}
	err := c.runScript("testdata/hello.grub")
	if err != nil {
		fmt.Printf("Unexpected error from runScript: %s\n", err)
	}
	// Output:
	// Hello world!
}
