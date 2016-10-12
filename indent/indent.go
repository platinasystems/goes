// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package indent provides a wrapper to a Writer interface that inserts a
// leveled preface to the beginning of each line.
package indent

import (
	"bytes"
	"io"
	"reflect"
)

// Decrease the preface repetition level for the given or embedded indentor.
// This will not go below zero.
func Decrease(w io.Writer) {
	if indent := find(w); indent != nil && indent.lvl > 0 {
		indent.lvl -= 1
	}
}

// Increase the preface repetition level for the given or embedded indentor.
func Increase(w io.Writer) {
	if indent := find(w); indent != nil {
		indent.lvl += 1
	}
}

type indentor struct {
	io.Writer
	preface string
	lvl     int
	bol     bool
}

// recursive search for the embedded indentor
func find(w io.Writer) *indentor {
	indent, found := w.(*indentor)
	if !found {
		el := reflect.ValueOf(w).Elem()
		if el.Kind() == reflect.Struct {
			t := el.Type()
			for i := 0; i < el.NumField(); i++ {
				tf := t.Field(i)
				if len(tf.PkgPath) == 0 || tf.Anonymous {
					// test field only if exported
					elf := el.Field(i)
					w, found := elf.Interface().(io.Writer)
					if found {
						return find(w)
					}
				}
			}
		}
	}
	return indent
}

// New returns an indentor initialized to level zero, meaning, no preface is
// inserted on each line until this is Increase[d].
func New(w io.Writer, preface string) io.Writer {
	return &indentor{w, preface, 0, true}
}

func (indent *indentor) insert() (n int, err error) {
	for lvl := 0; lvl < indent.lvl; lvl++ {
		var i int
		i, err = indent.Writer.Write([]byte(indent.preface))
		if err != nil {
			return
		}
		n += i
	}
	return
}

func (indent *indentor) Write(b []byte) (n int, err error) {
	var i int
	if len(b) == 0 {
		return
	}
	if indent.lvl == 0 {
		n, err = indent.Writer.Write(b)
		indent.bol = b[len(b)-1] == '\n'
		return
	}
	if indent.bol {
		indent.bol = false
		if i, err = indent.insert(); err != nil {
			return
		}
		n += i
	}
	for len(b) > 0 {
		nl := bytes.Index(b, []byte("\n"))
		if nl < 0 {
			if i, err = indent.Writer.Write(b); err == nil {
				n += i
			}
			break
		}
		if i, err = indent.Writer.Write(b[:nl+1]); err != nil {
			break
		}
		n += i
		if nl == len(b)-1 {
			indent.bol = true
			break
		}
		if i, err = indent.insert(); err != nil {
			break
		}
		b = b[nl+1:]
	}
	return
}
