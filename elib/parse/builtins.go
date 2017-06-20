// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"encoding/hex"
	"regexp"
	"unicode"
)

// Boolean parser accepting yes/no 0/1
type Bool bool

func (b *Bool) Parse(in *Input) {
	switch text := in.Token(); text {
	case "true", "yes", "1":
		*b = true
	case "false", "no", "0":
		*b = false
	default:
		panic(ErrInput)
	}
	return
}

// Boolean parser accepting enable/disable yes/no
type Enable bool

func (b *Enable) Parse(in *Input) {
	switch text := in.Token(); text {
	case "enable", "yes", "1":
		*b = true
	case "disable", "no", "0":
		*b = false
	default:
		panic(ErrInput)
	}
	return
}

// Boolean parser accepting up/down yes/no
type UpDown bool

func (b *UpDown) Parse(in *Input) {
	switch text := in.Token(); text {
	case "up", "yes", "1":
		*b = true
	case "down", "no", "0":
		*b = false
	default:
		panic(ErrInput)
	}
	return
}

type HexString []byte

func (x *HexString) Parse(in *Input) {
	src := in.Token()
	dst := make([]byte, len(src)/2)
	if _, err := hex.Decode(dst, []byte(src)); err != nil {
		panic(err)
	}
	*x = HexString(dst)
}

type StringMap map[string]uint

func (sm *StringMap) Set(v string, i uint) {
	m := *sm
	if m == nil {
		m = make(StringMap)
	}
	m[v] = i
	*sm = m
}

func NewStringMap(a []string) (m StringMap) {
	m = make(map[string]uint)
	for i := range a {
		if len(a[i]) > 0 {
			m[a[i]] = uint(i)
		}
	}
	return m
}

func (m StringMap) ParseWithArgs(in *Input, args *Args) {
	text := in.TokenF(func(r rune) bool {
		// Want to accept foo-bar as a token, otherwise terminate by punctuation.
		return unicode.IsSpace(r) || (unicode.IsPunct(r) && r != '-')
	})
	if v, ok := m[text]; ok {
		args.SetNextInt(uint64(v))
	} else {
		panic(ErrInput)
	}
	return
}

type Regexp struct{ *regexp.Regexp }

func (r *Regexp) Valid() bool { return r.Regexp != nil }
func (r *Regexp) Parse(in *Input) {
	text := in.Token()
	var err error
	if r.Regexp, err = regexp.Compile(text); err != nil {
		panic(err)
	}
}

type Comment struct{}

func (x *Comment) Parse(in *Input) {
	if in.Parse("//") {
		for !in.EndNoSkip() {
			if x, _ := in.ReadRune(); x == '\n' {
				return
			}
		}
	}
	panic(ErrInput)
}
