// Copyright © 2021-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xlog

import (
	"bytes"
	"fmt"
)

func ExampleMutable() {
	buf := new(bytes.Buffer)
	m := NewUnmutedLogger(buf)
	m.Write([]byte("write\n"))
	m.Print("print\n")
	m.Printf("%s\n", "printf")
	m.Println("println")
	m.Mute()
	m.Print("muted")
	m.Toggle()
	m.Print("toggled")
	fmt.Print(buf)
	// Output:
	// write
	// example_mute_test.go:16: print
	// example_mute_test.go:17: printf
	// example_mute_test.go:18: println
	// example_mute_test.go:22: toggled
}
