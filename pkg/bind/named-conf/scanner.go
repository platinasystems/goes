// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named_conf

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

// [Scanner].Text() returns the next:
//   - key continuation “\s+(-)\s\+”
//   - quoted-string “"([^"]"”
//   - space separated string “\s?(\w\+)\s?”
//   - beginning of block “{”
//   - end of block “}”
//   - end of statement “;”
type Scanner struct {
	*bufio.Scanner
	name  string
	line  int
	total int64
}

func NewScanner(r io.Reader) *Scanner {
	sc := new(Scanner)
	sc.Scanner = bufio.NewScanner(r)
	sc.Split(sc.split)
	if namer, ok := r.(interface {
		Name() string
	}); ok {
		sc.name = filepath.Clean(namer.Name())
		for i, seps := len(sc.name), 0; ; {
			i = strings.LastIndexFunc(sc.name[:i],
				func(ŕ rune) bool {
					return ŕ == filepath.Separator
				})
			if i < 0 {
				break
			}
			if seps += 1; seps == 3 {
				sc.name = "..." + sc.name[i+1:]
				break
			}
		}
	}
	sc.line = 1
	return sc
}

// Scan input up to semicolon, ';' end of statement,
// or close brace, '}' end of block.
// Returns [Statement] if terminated with semicolon;
// [Block] if terminated with close brace;
// []any non empty args before ![Scanner].Scan(),
// or [Scanner].Err() if there weren't any accumulated args.
func (sc *Scanner) args() (any, error) {
	var args []any
	for sc.Scan() {
		switch token := sc.Text(); token {
		case "/*":
			if err := sc.noComment(); err != nil {
				return args, err
			}
		case "{":
			if blk, err := sc.block(); err != nil {
				return args, err
			} else {
				args = append(args, blk)
			}
		case "}":
			return Block(args), nil
		case ";":
			return Statement(args), nil
		default:
			args = attach(args, token)
		}
	}
	return args, sc.Err()
}

// Attach token to previous hyphenated (“-”) key-word,
// slash (“/”) separated net-prefix,
// or colon (“:”) separated ip6 address;
// otherwise, clone and append to args.
func attach(args []any, token string) []any {
	if n := len(args); n >= 2 {
		s_1, s_1_ok := args[n-1].(string)
		s_2, s_2_ok := args[n-2].(string)
		if s_1_ok && s_2_ok &&
			(s_1 == "-" || s_1 == "/" || s_1 == ":") {
			args[n-2] = fmt.Sprint(s_2, s_1, token)
			return args[:n-1]
		}
	}
	return append(args, strings.Clone(token))
}

func (sc *Scanner) block() (Block, error) {
	var block Block
	for {
		args, err := sc.args()
		if err != nil {
			return block, err
		}
		switch t := args.(type) {
		case []any:
			return append(block, t...), err
		case Block:
			return append(block, t...), err
		case Statement:
			if err != nil {
				return block, err
			}
			block = append(block, t)
		default:
			err := SyntaxErr("%T(%v) isn't a Block or Statement",
				t, t)
			return block, sc.label(err)
		}
	}
}

func (sc *Scanner) conf() (Conf, error) {
	var conf Conf
	for {
		args, err := sc.args()
		if err != nil {
			return conf, err
		}
		switch t := args.(type) {
		case []any:
			if len(t) == 0 {
				return conf, err
			} else {
				return conf, sc.label(ErrNoEOS)
			}
		case Statement:
			if len(t) == 0 {
				// ignore empty statement
			} else if s, ok := t[0].(string); ok &&
				s == "include" {
				if len(t) != 2 {
					err := SyntaxErr("missing file name")
					return conf, sc.label(err)
				}
				if fn, ok := t[0].(string); !ok {
					err := SyntaxErr("improper file name")
					return conf, sc.label(err)
				} else if f, err := Open(fn); err != nil {
					return conf, err
				} else {
					stmts, err := NewScanner(f).conf()
					f.Close()
					if err != nil {
						return conf, err
					}
					conf = append(conf, stmts...)
				}
			} else {
				conf = append(conf, t)
			}
		default:
			err := SyntaxErr("%T(%v) isn't a Statement", t, t)
			return conf, sc.label(err)
		}
	}
}

// Label an error with [Scanner]'s file name and line number.
func (sc *Scanner) label(err error, args ...any) error {
	if len(sc.name) == 0 {
		return xerrors.Label(err, sc.line)
	}
	return xerrors.Label(err, sc.name, sc.line)
}

func (sc *Scanner) noComment() error {
	for sc.Scan() {
		switch sc.Text() {
		case "/*":
			if err := sc.noComment(); err != nil {
				return err
			}
		case "*/":
			return nil
		}
	}
	return ErrNoEOC
}

func (sc *Scanner) split(data []byte, atEOF bool) (int, []byte, error) {
	for i := 0; i < len(data); {
		r, sz := utf8.DecodeRune(data[i:])
		switch r {
		case '\n':
			sc.line += 1
			fallthrough
		case '\x01', '\x00', ' ', '\f', '\r', '\v', '\t':
			i += sz
		case '{', '}', ';':
			return i + sz, data[i : i+sz], nil
		case '"':
			// quoted string,
			// gather to end-quote while replacing escaped quotes.
			i += sz
			for last, lastsz, j, nls := r, sz, i, 0; true; j += sz {
				if j == len(data) {
					if atEOF {
						return j, nil, sc.label(ErrNoEOQ)
					}
					return 0, nil, nil
				}
				r, sz = utf8.DecodeRune(data[j:])
				switch r {
				case '\n':
					nls += 1
				case '"':
					if last == '\\' {
						j -= lastsz
						copy(data[j:], data[j+sz:])
						data = data[:len(data)-sz]
					} else {
						sc.line += nls
						return j + sz, data[i:j], nil
					}
				}
				last = r
				lastsz = sz
			}
		case '#':
			// shell-comment, ignore to EOL
			var j int
			for j = i + sz; r != '\n'; j += sz {
				if j == len(data) {
					if atEOF {
						return j, nil, nil
					}
					return 0, nil, nil
				}
				r, sz = utf8.DecodeRune(data[j:])
			}
			sc.line += 1
			i = j
		case '/':
			j := i + sz
			if j == len(data) {
				if atEOF {
					return j, data[i:j], nil
				}
				return 0, nil, nil
			}
			r, sz = utf8.DecodeRune(data[j:])
			switch r {
			case ' ', '\t':
				return j + sz, data[i:j], nil
			case '/':
				// slash-comment, ignore to EOL
				// tlogf("begin slash comment: %q", data[i:])
				for j += sz; r != '\n'; j += sz {
					if j == len(data) {
						if atEOF {
							return j, nil, nil
						}
						return 0, nil, nil
					}
					r, sz = utf8.DecodeRune(data[j:])
				}
				sc.line += 1
				i = j
			case '*':
				// begin c-lang comment
				// tlogf("begin c-lang comment: %q", data[i:])
				j += sz
				return j, data[i:j], nil
			default:
				return sc.token(i, j+sz, data, atEOF)
			}
		case '*':
			j := i + sz
			if j == len(data) {
				if atEOF {
					return j, data[i:j], nil
				}
				return 0, nil, nil
			}
			r, sz = utf8.DecodeRune(data[j:])
			switch r {
			case '/':
				// end c-lang comment
				return j + sz, data[i : j+sz], nil
			case '\n':
				// `\*\n`
				sc.line += 1
				fallthrough
			case '.':
				// *.host.domain.A
				return sc.token(i, i+j+sz, data, atEOF)
			default:
				return j, data[i:j], nil
			}
		default:
			return sc.token(i, i+sz, data, atEOF)
		}
	}
	return len(data), nil, nil
}

// return token upto an unescaped terminator.
func (sc *Scanner) token(i, j int, data []byte, atEOF bool) (
	int, []byte, error,
) {
	last, lastsz := utf8.DecodeRune(data[i:])
	for j < len(data) {
		r, sz := utf8.DecodeRune(data[j:])
		if last == '\\' {
			j -= lastsz
			copy(data[j:], data[j+sz:])
			data = data[:len(data)-sz]
		} else {
			switch r {
			case '\n':
				sc.line += 1
				fallthrough
			case ' ', '\t':
				return j + sz, data[i:j], nil
			case '{', '}', ';':
				return j, data[i:j], nil
			}
		}
		last = r
		lastsz = sz
		j += sz
	}
	if atEOF {
		return j, data[i:j], nil
	}
	return 0, nil, nil
}
