// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package shellutils

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

type ts struct {
	line   int
	script []string
}

func (t *ts) Write(p []byte) (n int, err error) {
	return n, nil
}

func (t *ts) Read(p []byte) (n int, err error) {
	if t.line >= len(t.script) {
		return 0, errors.New("parser asked for too much input")
	}
	s := t.script[t.line]
	t.line += 1
	n = copy(p, []byte(s))
	if len(s) > len(p) {
		err = errors.New("input too long")
	}
	return
}

func testSlice(script []string) (*List, error) {
	t := &ts{script: script}
	ls, err := Parse(">", t)

	if err != nil {
		return nil, err
	}

	if t.line != len(script) {
		err := errors.New("parser did not consume all input")
		return nil, err
	}
	return ls, err
}

func (ls *List) print() {
	for _, sl := range ls.Cmds {
		_, cmdline := sl.Slice(os.Getenv)
		term := sl.Term.String()
		if term == "" {
			term = "\n"
		} else {
			term = " " + term + " "
		}
		fmt.Print(strings.Join(cmdline, " "), term)
	}
}

func TestCommandOnly(t *testing.T) {
	script := []string{"ls -l"}

	cmd, err := testSlice(script)

	if err != nil {
		t.Error(err)
		return
	}
	cmd.print()
}

func TestBackquoteContinuation(t *testing.T) {
	script := []string{"echo this \\", "is a \\", "test"}

	cmd, err := testSlice(script)
	if err != nil {
		t.Error(err)
		return
	}

	cmd.print()
}

func TestPipeline(t *testing.T) {
	script := []string{"ls | more"}

	cmd, err := testSlice(script)
	if err != nil {
		t.Error(err)
		return
	}

	cmd.print()
}

func TestBoolean(t *testing.T) {
	script := []string{"true && echo yes || echo no"}

	cmd, err := testSlice(script)
	if err != nil {
		t.Error(err)
		return
	}

	cmd.print()
}

// TestDoublequote: The backslash retains its special meaning only when followed by one of the following characters:
// ‘$’, ‘`’, ‘"’, ‘\’, or newline.
func TestDoublequote(t *testing.T) {
	script := []string{`"hello" "foo\$\"\a\b\c\n\"`, "world", "keep", "eating", `until now"`}

	cmd, err := testSlice(script)
	if err != nil {
		t.Error(err)
		return
	}

	cmd.print()
}
