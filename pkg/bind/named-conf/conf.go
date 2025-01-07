// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package named-conf provides [NewConf] to unmarshal [named.conf] into [Conf],
// a list of [Statement]s.
//
// Each [Statement] is a semicolon terminated parameter.
//
// The parameter begins with it's name, a potentially hyphenated keyword,
// followed by space separated argument(s).
// The arguments may be a single, formatted [CoreValue];
// a sequence of sub-parameter (name, value) pairs;
// or a brace encapsulated [Block];
//
// A Block may be contain a list of [Statements];
// or w/o semicolons, a sequence of (name, value) parameters.
// or a list of unnamed values.
//
// [named.conf]: https://bind9.readthedocs.io/en/latest/reference.html#configuration-reference
package named_conf

import (
	"errors"
	"fmt"
	"io"
	"net/netip"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

var (
	mute  = func(args ...any) {}
	mutef = func(format string, args ...any) {}
	tlog  = mute
	tlogf = mutef
)

var (
	ErrNoEOL  = errors.New(`no end of line`)
	ErrNoEOC  = errors.New(`no end of comment (*/)`)
	ErrNoEOQ  = errors.New(`no end quote (")`)
	ErrNoEOS  = errors.New(`no end of statement (;)`)
	ErrSyntax = errors.New(`invalid syntax`)
)

// Include tests may be override Open.
var Open = func(filename string) (io.ReadCloser, error) {
	f, err := os.Open(filename)
	return f, err
}

func SyntaxErr(format string, args ...any) error {
	return fmt.Errorf("%w - "+format, append([]any{ErrSyntax},
		args...)...)
}

// A Block may be a brace encapsulated zero or more semicolon terminated
// [Statement]s followed by zero or more space separated [CoreValue]s,
// e.g.
//
//	{ one 1; two 2; three 3; }
//	{ one 1; two 2; three 3 }
//	{ one 1; two 2 three 3 }
//	{ one 1 two 2 three 3 }
type Block []any

func (block Block) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "{")
	iw := NewIndent(w)
	for _, v := range block {
		fmt.Fprint(iw, "\n", v)
	}
	fmt.Fprint(w, "\n}")
}

func (block Block) ranger(do Do, acc []string, path ...any) bool {
	if len(path) == 0 || len(block) == 0 {
		return true
	}
	for _, v := range block {
		if stmt, ok := v.(Statement); ok {
			if !stmt.ranger(do, acc, path...) {
				return false
			}
		}
	}
	return true
}

func (block Block) verify(expect Block) error {
	for i, v := range block {
		if i >= len(expect) {
			return fmt.Errorf("UNEXPECTED: %v", v)
		}
		if err := verify(v, expect[i]); err != nil {
			return err
		}
	}
	if len(block) < len(expect) {
		return fmt.Errorf("MISSING: %v", expect[len(block)])
	}
	return nil
}

type Conf []Statement

func NewConf(filename string) (Conf, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return NewScanner(f).conf()
}

func (conf Conf) Format(w fmt.State, verb rune) {
	for _, stmt := range conf {
		fmt.Fprintln(w, stmt)
	}
}

// Run callback with each [Block] or remaining [Statement] matched by the
// descendent path.
// The path elements may be an exact match string or pattern matched regexp.
// (see [TestRange])
func (conf Conf) Range(do Do, path ...any) {
	var acc []string
	for _, stmt := range conf {
		if !stmt.ranger(do, acc, path...) {
			break
		}
	}
}

func (conf Conf) verify(expect Conf) error {
	for i, stmt := range conf {
		if i >= len(expect) {
			return fmt.Errorf("UNEXPECTED: %v", stmt)
		}
		if err := stmt.verify(expect[i]); err != nil {
			return err
		}
	}
	if len(conf) < len(expect) {
		return fmt.Errorf("MISSING: %v", expect[len(conf)])
	}
	return nil
}

type CoreValue interface {
	bool | float32 | string | uint16 | uint32 | uint64 |
		netip.Addr | netip.Prefix |
		time.Duration
}

// Do is called by [Conf.Range] passing accumulated path and the matching
// [Block] or remaining [Statement].
// Range stops iteration when Do returns false.
type Do func([]string, any) bool

// A Statement is a semicolon terminated list of [CoreValue]s and [Block]s, e.g.
//
//	controls {
//		inet 127.0.0.1 port 9953 allow { any; } keys { rndc_key; };
//	};
type Statement []any

func (stmt Statement) Format(w fmt.State, verb rune) {
	for i, v := range stmt {
		if i > 0 {
			fmt.Fprint(w, " ")
		}
		if s, ok := v.(string); ok && strings.ContainsAny(s, " \t\n") {
			s = strings.ReplaceAll(s, `"`, `\"`)
			fmt.Fprintf(w, `"%s"`, s)
		} else {
			fmt.Fprint(w, v)
		}
	}
	fmt.Fprint(w, ";")
}

func (stmt Statement) ranger(do Do, acc []string, path ...any) bool {
	if len(path) == 0 || len(stmt) < 2 {
		return true
	}
	match := pathmatch(path)
	s, ok := stmt[0].(string)
	if !ok || !match(s) {
		return true
	}
	acc = append(acc, s)
	if len(path) == 1 {
		return do(acc, stmt[1:])
	}
	switch t := stmt[1].(type) {
	case Block:
		return t.ranger(do, acc, path[1:]...)
	case string:
		return stmt[1:].ranger(do, acc, path[1:]...)
	}
	return true
}

func (stmt Statement) verify(expect Statement) error {
	if len(stmt) != len(expect) {
		return fmt.Errorf("MISMATCH\n%#v\n---\n%#v", stmt, expect)
	}
	var k string
	for i, v := range stmt {
		if err := verify(v, expect[i]); err != nil {
			if len(k) > 0 {
				err = xerrors.Label(err, k)
			}
			return err
		}
		if s, ok := v.(string); ok {
			k = s
		}
	}
	return nil
}

func pathmatch(path []any) func(string) bool {
	match := func(string) bool {
		return false
	}
	if len(path) > 0 {
		if s, ok := path[0].(string); ok {
			match = func(kw string) bool {
				return s == kw
			}
		} else if re, ok := path[0].(*regexp.Regexp); ok {
			match = re.MatchString
		}
	}
	return match
}

func verify(have, want any) error {
	switch t := have.(type) {
	case Block:
		if x, ok := want.(Block); !ok {
			return fmt.Errorf("MISTYPED %T != %T", t, want)
		} else {
			return t.verify(x)
		}
	case Statement:
		if x, ok := want.(Statement); !ok {
			return fmt.Errorf("MISTYPED %T != %#T", t, want)
		} else {
			return t.verify(x)
		}
	case string:
		if s, ok := want.(string); !ok {
			return fmt.Errorf("MISTYPED %T != %T", t, want)
		} else if t != s {
			return fmt.Errorf("MISMATCH %q != %q", t, s)
		}
	default:
		return fmt.Errorf("%w - %T(%v)", ErrSyntax, t, t)
	}
	return nil
}
