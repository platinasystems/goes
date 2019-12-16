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

func testSlice(script []string) (*List, error) {
	calls := 0

	ls, err := Parse(">", func(prompt string) (string, error) {
		if calls < len(script) {
			s := script[calls]
			calls += 1
			return s, nil
		}
		return "", errors.New("parser asked for too much input")
	})

	if err != nil {
		return nil, err
	}

	if calls != len(script) {
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

func TestFail(t *testing.T) {
	t.Error(errors.New("Faceplant into rock!"))
}
