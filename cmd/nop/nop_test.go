// Copyright © 2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nop

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

func (c Command) testLazyCmdForcingError(t *testing.T,
	pre string, preExpected error,
	mid string, midExpected error,
	post string, postExpected error) {
	args := []string{}
	if pre != "" {
		args = append(args, pre)
		c.testCmdWithArgs(t, preExpected, args...)
	}
	c.testCmdWithArgs(t, preExpected, args...)
	args = append(args, "the")
	c.testCmdWithArgs(t, preExpected, args...)
	args = append(args, " quick")
	c.testCmdWithArgs(t, preExpected, args...)
	args = append(args, " brown")
	c.testCmdWithArgs(t, preExpected, args...)
	args = append(args, " fox")
	c.testCmdWithArgs(t, preExpected, args...)
	if mid != "" {
		args = append(args, mid)
		c.testCmdWithArgs(t, midExpected, args...)
	}
	args = append(args, " jumps")
	c.testCmdWithArgs(t, midExpected, args...)
	args = append(args, " over")
	c.testCmdWithArgs(t, midExpected, args...)
	args = append(args, " the")
	c.testCmdWithArgs(t, midExpected, args...)
	args = append(args, " lazy")
	c.testCmdWithArgs(t, midExpected, args...)
	args = append(args, " dog")
	c.testCmdWithArgs(t, midExpected, args...)
	if post != "" {
		args = append(args, post)
		c.testCmdWithArgs(t, postExpected, args...)
	}

}

func (c Command) testLazyCmd(t *testing.T) {
	c.testLazyCmdForcingError(t, "", nil, "", nil, "", nil)
	c.testLazyCmdForcingError(t,
		"--force-error", ErrorForced,
		"", ErrorForced,
		"", ErrorForced)
	c.testLazyCmdForcingError(t,
		"", nil,
		"--force-error", ErrorForced,
		"", ErrorForced)
	c.testLazyCmdForcingError(t,
		"", nil,
		"", nil,
		"--force-error", ErrorForced)
}

func TestDefaultCommand(t *testing.T) {
	c := Command{}
	c.testLazyCmd(t)
}

func ExampleMain() {
	c := Command{}
	fmt.Println(c)
	fmt.Println(c.Usage())
	fmt.Println(c.Apropos())
	fmt.Println(c.Man())
	fmt.Println(c.Kind())
	// Output:
	// nop
	// nop ...
	// do nothing
	//
	// DESCRIPTION
	//	The nop command does nothing. It is intended for use in
	//	scripts.
	// don't fork

}

func TestXyzzyCommand(t *testing.T) {
	c := Command{C: "xyzzy"}
	c.testLazyCmd(t)
}

func ExampleMain_xyzzy() {
	c := Command{C: "xyzzy"}
	fmt.Println(c)
	fmt.Println(c.Usage())
	fmt.Println(c.Apropos())
	fmt.Println(c.Man())
	fmt.Println(c.Kind())
	// Output:
	// xyzzy
	// xyzzy ...
	// do nothing
	//
	// DESCRIPTION
	//	The xyzzy command does nothing. It is intended for use in
	//	scripts.
	// don't fork
}
