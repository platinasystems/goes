// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package shellutils

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Word is a slice of Tokens. When converting to a string, all of the Tokens
// are evaluated to produce strings, which are concatenated.

type Word struct {
	Tokens []Token
}

// add adds a Token to the current Word being parsed
func (w *Word) add(s string, ty Tokentype) {
	if w.Tokens == nil {
		w.Tokens = make([]Token, 0)
	}
	t := Token{V: s, T: ty}
	w.Tokens = append(w.Tokens, t)
}

// addLiteral is a helper routine to add literal text. It has the optimization
// of concatenating successful calls to addLiteral. This is helpful because
// addLiteral is mostly called rune by rune
func (w *Word) addLiteral(s string) {
	if len(w.Tokens) > 0 {
		end := len(w.Tokens) - 1
		if w.Tokens[end].T == TokenLiteral {
			w.Tokens[end].V += s
			return
		}
	}
	w.add(s, TokenLiteral)
}

func (w *Word) parseEnv(s string) (string, error) {
	envvar := ""
	if s[0] == '{' {
		s = s[1:]
		for len(s) > 0 {
			r, wid := utf8.DecodeRuneInString(s)
			s = s[wid:]
			if r == '}' {
				w.add(envvar, TokenEnvget)
				return s, nil
			}
			if unicode.IsSpace(r) || strings.ContainsRune("|&;()<>{'\"$/", r) {
				return "", fmt.Errorf("Unexpected `%c'", r)
			}
			envvar += string(r)
		}
		return "", errors.New("Unexpected end-of-line")
	}

	for len(s) > 0 {
		r, wid := utf8.DecodeRuneInString(s)
		if unicode.IsSpace(r) || strings.ContainsRune("|&;()<>{}'\"$/", r) {
			break
		}
		s = s[wid:]
		envvar += string(r)
	}
	w.add(envvar, TokenEnvget)
	return s, nil
}

func (w *Word) String() string {
	s := ""
	for _, t := range w.Tokens {
		s += t.V
	}
	return s
}
