// Copyright © 2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package menuentry

import (
	"errors"
	"fmt"
	"testing"
)

func (c Command) testCmdWithArgs(t *testing.T, expected error, args ...string) {
	err := c.Main(args...)
	if !errors.Is(err, expected) {
		r := "success"
		if err != nil {
			r = err.Error()
		}
		x := "success"
		if expected != nil {
			x = expected.Error()
		}
		t.Errorf("%s Main(%v) failed: returned %s [expected %s]",
			c.String(), args, r, x)
	}
}

func TestMenuentryMain(t *testing.T) {
	c := Command{}
	c.testCmdWithArgs(t, InternalError)
}

func ExampleDefaultCommand() {
	c := Command{}
	fmt.Println(c)
	fmt.Println(c.Usage())
	fmt.Println(c.Apropos())
	fmt.Println(c.Man())
	// Output:
	// menuentry
	// menuentry [options] [name] { script ... }
	// define a menu item
	//
	// DESCRIPTION
	//	Define a menu item.
	//
	//	Options and names are currently ignored. They do not return
	//	errors for compatibility with existing grub scripts.
	//
	//	The menu itself is a script which will be run when the menu
	//	item is selected.
}
