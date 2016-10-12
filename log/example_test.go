// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package log

import (
	"os"
	"time"
)

func setup() {
	pid = 6789
	writer = os.Stdout
}

func ExamplePrint() {
	setup()
	Print("a message with default facility/priority [fac/pri]")
	Print("daemon", "a message with daemon fac and default pri")
	Print("daemon", "info", "a message with daemon fac and info pri")
	Print(`
a multi-
line message with default fac/pri`[1:])
	// Output:
	// <15>goes[6789]: a message with default facility/priority [fac/pri]
	// <31>goes[6789]: a message with daemon fac and default pri
	// <30>goes[6789]: a message with daemon fac and info pri
	// <15>goes[6789]: a multi-
	// <15>goes[6789]: line message with default fac/pri
}

func ExamplePrintf() {
	setup()
	Printf("a formatted message %s", "with default fac/pri")
	Printf("daemon", "a formatted message %s",
		"with daemon fac and default pri")
	Printf("daemon", "info", "a formatted message %s",
		"with daemon fac and info pri")
	Printf("a formatted %s", `
multi-
line message with default fac/pri`[1:])
	// Output:
	// <15>goes[6789]: a formatted message with default fac/pri
	// <31>goes[6789]: a formatted message with daemon fac and default pri
	// <30>goes[6789]: a formatted message with daemon fac and info pri
	// <15>goes[6789]: a formatted multi-
	// <15>goes[6789]: line message with default fac/pri
}

func ExampleLimitedPrint() {
	setup()
	l := NewLimited(3)
	l.Print("first message")
	l.Print("second message")
	l.Print("third message")
	l.Print("fourth message should be dropped")
	// Output:
	// <15>goes[6789]: first message
	// <15>goes[6789]: second message
	// <15>goes[6789]: third message
}

func ExampleRateLimitedPrint() {
	setup()
	rl := NewRateLimited(3, 500*time.Millisecond)
	defer rl.Close()
	rl.Print("first message")
	rl.Print("second message")
	rl.Print("third message")
	rl.Print("fourth message should be dropped")
	time.Sleep(750 * time.Millisecond)
	rl.Print("fifth message after fourth is rate limited")
	// Output:
	// <15>goes[6789]: first message
	// <15>goes[6789]: second message
	// <15>goes[6789]: third message
	// <15>goes[6789]: fifth message after fourth is rate limited
}
