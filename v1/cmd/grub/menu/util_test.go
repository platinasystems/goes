// Copyright © 2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package menu

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/cmd/cli"
	"github.com/platinasystems/goes/lang"
)

var ErrTestpointFailed = errors.New("testpoint failed")

type ts struct {
	line   int
	script []string
}

func (t *ts) Write(p []byte) (n int, err error) {
	//fmt.Printf("Test write %s\n", string(p))
	return n, nil
}

func (t *ts) Read(p []byte) (n int, err error) {
	if t.line >= len(t.script) {
		//fmt.Printf("Test read returned EOF\n")
		return 0, io.EOF
	}
	//fmt.Printf("Test read returning: %s\n", t.script[t.line])
	s := t.script[t.line]
	t.line += 1
	n = copy(p, []byte(s))
	if len(s) > len(p) {
		err = errors.New("input too long")
	}
	return
}

func (t *ts) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "script testpoint",
	}
}

func (t *ts) Main(args ...string) (err error) {
	fmt.Println("testpoint", args)
	if len(args) > 0 {
		if args[0] == "fail" {
			return ErrTestpointFailed
		}
	}
	return nil
}

func (t *ts) String() string { return "testpoint" }

func (t *ts) Usage() string { return "testpoint" }

func newgoes(t *ts) (m, s *Command, g *goes.Goes) {
	m, s = New()
	return m, s, &goes.Goes{
		NAME: "test",
		ByName: map[string]cmd.Cmd{
			"cli":       &cli.Command{},
			"menuentry": m,
			"submenu":   s,
			"testpoint": t,
		},
		Catline: t,
	}
}

func testCmdWithArgs(c cmd.Cmd, t *testing.T,
	expected error, args ...string) {
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

func testScriptNoArgs(s1 *ts, t *testing.T) {
	_, _, g := newgoes(s1)
	err := g.Main("test", "cli", "-")
	if !errors.Is(err, ErrMenuUnexpectedEOL) {
		if err == nil {
			t.Errorf("Got no error, expected %s",
				ErrMenuUnexpectedEOL)
		} else {
			t.Errorf("Got %s, expected %s", err,
				ErrMenuUnexpectedEOL)
		}
	}
}

func testScriptNoOpenBrace(s1 *ts, t *testing.T) {
	_, _, g := newgoes(s1)
	err := g.Main("test", "cli", "-")
	if !errors.Is(err, ErrMenuMissingOpenBrace) {
		if err == nil {
			t.Errorf("Got no error, expected %s",
				ErrMenuMissingOpenBrace)
		} else {
			t.Errorf("Got %s, expected %s", err,
				ErrMenuMissingOpenBrace)
		}
	}
}

func testScriptMissingMenuName(s1 *ts, t *testing.T) {
	_, _, g := newgoes(s1)
	err := g.Main("test", "cli", "-")
	if !errors.Is(err, ErrMenuMissingName) {
		if err == nil {
			t.Errorf("Got no error, expected %s",
				ErrMenuMissingName)
		} else {
			t.Errorf("Got %s, expected %s", err,
				ErrMenuMissingName)
		}
	}
}

func testScriptUnexpectedText(s1 *ts, t *testing.T) {
	_, _, g := newgoes(s1)
	err := g.Main("test", "cli", "-")
	if !errors.Is(err, ErrMenuUnexpectedText) {
		if err == nil {
			t.Errorf("Got no error, expected %s",
				ErrMenuUnexpectedText)
		} else {
			t.Errorf("Got %s, expected %s", err,
				ErrMenuUnexpectedText)
		}
	}
}

func testScriptNestedMenus(s1 *ts, t *testing.T, nesting bool) {
	_, _, g := newgoes(s1)
	err := g.Main("test", "cli", "-")
	if nesting {
		if err != nil {
			t.Errorf("Unexpected error %s", err)
		}
	} else {
		if !errors.Is(err, ErrMenuNoNesting) {
			if err == nil {
				t.Errorf("Got no error, expected %s",
					ErrMenuNoNesting)
			} else {
				t.Errorf("Got %s, expected %s", err,
					ErrMenuNoNesting)
			}
		}
	}
}
