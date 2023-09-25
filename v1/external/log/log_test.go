// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package log

import (
	"bytes"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	cache.pid = 6789
	tee.exclusive = true
	os.Exit(m.Run())
}

func TestPrint(t *testing.T) {
	defer expect(`
<15>log.test[6789]: a message with default facility/priority [fac/pri]
<31>log.test[6789]: a message with daemon fac and default pri
<30>log.test[6789]: a message with daemon fac and info pri
<15>log.test[6789]: a multi-
<15>log.test[6789]: line message with default fac/pri
`[1:]).results(t)

	Print("a message with default facility/priority [fac/pri]")
	Print("daemon", "a message with daemon fac and default pri")
	Print("daemon", "info", "a message with daemon fac and info pri")
	Print(`
a multi-
line message with default fac/pri`[1:])
}

func TestPrintf(t *testing.T) {
	defer expect(`
<15>log.test[6789]: a formatted message with default fac/pri
<31>log.test[6789]: a formatted message with daemon fac and default pri
<30>log.test[6789]: a formatted message with daemon fac and info pri
<15>log.test[6789]: a formatted multi-
<15>log.test[6789]: line message with default fac/pri
`[1:]).results(t)

	Printf("a formatted message %s", "with default fac/pri")
	Printf("daemon", "a formatted message %s",
		"with daemon fac and default pri")
	Printf("daemon", "info", "a formatted message %s",
		"with daemon fac and info pri")
	Printf("a formatted %s", `
multi-
line message with default fac/pri`[1:])
}

func TestLimitedPrint(t *testing.T) {
	defer expect(`
<15>log.test[6789]: first message
<15>log.test[6789]: second message
<15>log.test[6789]: third message
`[1:]).results(t)

	l := NewLimited(3)
	l.Print("first message")
	l.Print("second message")
	l.Print("third message")
	l.Print("fourth message should be dropped")
}

func TestRateLimitedPrint(t *testing.T) {
	defer expect(`
<15>log.test[6789]: first message
<15>log.test[6789]: second message
<15>log.test[6789]: third message
<15>log.test[6789]: fifth message after fourth is rate limited
`[1:]).results(t)

	rl := NewRateLimited(3, 500*time.Millisecond)
	defer rl.Close()
	rl.Print("first message")
	rl.Print("second message")
	rl.Print("third message")
	rl.Print("fourth message should be dropped")
	time.Sleep(750 * time.Millisecond)
	rl.Print("fifth message after fourth is rate limited")
}

func expect(s string) want {
	Tee(new(bytes.Buffer))
	return want(s)
}

type want string

func (s want) results(t *testing.T) {
	t.Helper()
	buf := tee.w.(*bytes.Buffer)
	got := buf.String()
	defer buf.Reset()
	if got != string(s) {
		t.Error("got:\n", got, "want:\n", string(s))
	} else if testing.Verbose() {
		os.Stdout.WriteString(got)
	}
}
