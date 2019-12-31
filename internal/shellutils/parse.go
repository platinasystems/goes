// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.
package shellutils

import (
	"errors"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

var ErrMissingEndQuote = errors.New("Unexpected EOF while looking for matching quote")

func srcin(i io.ReadWriter, prompt string) (s string, err error) {
	i.Write([]byte(prompt))
	buf := make([]byte, 1024)
	n, err := i.Read(buf)
	s = string(buf[0:n])
	return
}

// break up string into Lists, Pipelines, and command lines
// a List is a slice of Pipelines [][]Cmdline{}
// a Pipeline is a slice of commandlines []Cmdline{}
// a command line is a set of arguments and a terminator

// Parse calls the srcin function for command input as strings, and
// return a pointer to a parsed command List, or an error
func Parse(prompt string, i io.ReadWriter) (*List, error) {
	s, err := srcin(i, prompt)
	if err != nil {
		return nil, err
	}
	cl := List{}
	c := Cmdline{}
	w := Word{}
	inWS := true
processRune:
	for len(s) > 0 {
		r, wid := utf8.DecodeRuneInString(s)
		s = s[wid:]
		if inWS {
			if unicode.IsSpace(r) {
				continue
			}
			if r == '#' {
				break
			}
			inWS = false
		} else {
			if unicode.IsSpace(r) {
				c.add(&w)
				inWS = true
				continue
			}

			if strings.ContainsRune("|&;()<>", r) {
				c.add(&w)
			}
		}

		if strings.ContainsRune("&;()<", r) {
			w.addLiteral(string(r))
			// hack - we know these are single-byte runes
			if len(s) >= 1 && s[0] == byte(r) {
				s = s[1:]
				w.addLiteral(string(r))
			}
			if w.String() == ";" || w.String() == "&&" ||
				w.String() == "||" {
				c.Term = w
				w = Word{}
				cl.add(&c)
			} else {
				c.add(&w)
			}
			inWS = true
			continue
		}

		// Check for |, |&, or ||
		if r == '|' {
			w.addLiteral("|")
			if len(s) >= 1 {
				if s[0] == '&' {
					s = s[1:]
					w.addLiteral("&")
				} else {
					if s[0] == '|' {
						s = s[1:]
						w.addLiteral("|")
					}
				}
			}
			c.Term = w
			cl.add(&c)
			w = Word{}
			inWS = true
			continue
		}

		if r == '=' {
			w.add("=", TokenEnvset)
			continue
		}

		if r == '$' && len(s) > 0 {
			s, err = w.parseEnv(s)
			if err != nil {
				return nil, err
			}
			continue
		}

		if r == '>' {
			w.addLiteral(">")
			if len(s) >= 1 && s[0] == '>' {
				s = s[1:]
				w.addLiteral(">")
				if len(s) >= 1 && s[0] == '>' {
					s = s[1:]
					w.addLiteral(">")
					if len(s) >= 1 && s[0] == '>' {
						s = s[1:]
						w.addLiteral(">")
					}
				}
			}
			c.add(&w)
			inWS = true
			continue
		}

		if r == '\'' {
			for {
				for len(s) > 0 {
					r, wid := utf8.DecodeRuneInString(s)
					s = s[wid:]
					if r == '\'' {
						continue processRune
					}
					w.addLiteral(string(r))
				}
				w.addLiteral("\n")
				s, err = srcin(i, "> ")
				if err != nil {
					if err == io.EOF {
						return nil, ErrMissingEndQuote
					}
					return nil, err
				}
			}
		}
		if r == '"' {
			for {
				for len(s) > 0 {
					r, wid := utf8.DecodeRuneInString(s)
					s = s[wid:]
					if r == '"' {
						continue processRune
					}

					if r == '$' && len(s) > 0 {
						s, err = w.parseEnv(s)
						if err != nil {
							return nil, err
						}
						continue
					}
					if r == '\\' {
						if len(s) == 0 {
							s, err = srcin(i, "> ")
							if err != nil {
								if err == io.EOF {
									return nil, ErrMissingEndQuote
								}
								return nil, err
							}
							continue
						}
						r1, wid := utf8.DecodeRuneInString(s)
						if r1 == '$' || r1 == '"' || r1 == '\\' {
							r = r1
							s = s[wid:]
						}
					}
					w.addLiteral(string(r))
				}
				w.addLiteral("\n")
				s, err = srcin(i, "> ")
				if err != nil {
					if err == io.EOF {
						return nil, ErrMissingEndQuote
					}
					return nil, err
				}
			}
		}
		if r == '\\' {
			if len(s) > 0 {
				r, wid := utf8.DecodeRuneInString(s)
				s = s[wid:]
				w.addLiteral(string(r))
				continue
			}
			s, err = srcin(i, "... ")
			if err != nil {
				return nil, err
			}
			continue
		}
		w.addLiteral(string(r))
	}
	if len(w.Tokens) != 0 {
		c.add(&w)
	}
	if len(c.Cmds) != 0 {
		cl.add(&c)
	}
	return &cl, nil
}
