// Copyright © 2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nop

import (
	"fmt"
	"testing"
)

func (c Command) testCmdWithArgs(t *testing.T, args ...string) {
	err := c.Main(args...)
	if err != nil {
		t.Errorf("%s main(%v) failed: %w", c.String(), err, args)
	}
}

func (c Command) testLazyCmd(t *testing.T) {
	args := []string{}
	c.testCmdWithArgs(t, args...)
	args = append(args, "the")
	c.testCmdWithArgs(t, args...)
	args = append(args, " quick")
	c.testCmdWithArgs(t, args...)
	args = append(args, " brown")
	c.testCmdWithArgs(t, args...)
	args = append(args, " fox")
	c.testCmdWithArgs(t, args...)
	args = append(args, " jumps")
	c.testCmdWithArgs(t, args...)
	args = append(args, " over")
	c.testCmdWithArgs(t, args...)
	args = append(args, " the")
	c.testCmdWithArgs(t, args...)
	args = append(args, " lazy")
	c.testCmdWithArgs(t, args...)
	args = append(args, " dog")
	c.testCmdWithArgs(t, args...)
}

func TestDefaultCommand(t *testing.T) {
	c := Command{}
	c.testLazyCmd(t)
}

func ExampleDefaultCommand() {
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

func ExampleXyzzyCommand() {
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
