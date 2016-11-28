// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package sch provides a string channel with encapsulated endpoints that have
// common i/o helper methods.
package sch

import (
	"bufio"
	"fmt"
	"io"
)

type In chan<- string
type Out <-chan string

// New makes a string channel with the given buffer depth or unbuffered if 0.
func New(depth int) (In, Out) {
	sch := make(chan string, depth)
	return In(sch), Out(sch)
}

func (in In) Print(args ...interface{}) {
	in <- fmt.Sprint(args...)
}

func (in In) Println(args ...interface{}) {
	in <- fmt.Sprintln(args...)
}

func (in In) Printf(format string, args ...interface{}) {
	in <- fmt.Sprintf(format, args...)
}

func (in In) ReadFrom(r io.Reader) (n int64, err error) {
	for {
		var i int
		buf := make([]byte, 4096)
		i, err = r.Read(buf)
		if err != nil {
			buf = buf[:0]
			break
		}
		buf = buf[:i]
		in <- string(buf)
		n += int64(i)
	}
	if err == io.EOF {
		err = nil
	}
	return
}

func (in In) ReadFromThenClose(r io.ReadCloser) (int64, error) {
	defer r.Close()
	defer close(in)
	return in.ReadFrom(r)
}

func (in In) ReadLinesFrom(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		in <- scanner.Text()
	}
	return scanner.Err()
}

func (in In) ReadLinesFromThenClose(r io.ReadCloser) error {
	defer r.Close()
	defer close(in)
	return in.ReadLinesFrom(r)
}

const (
	withNL    = true
	withoutNL = false
)

func (out Out) writeto(w io.Writer, wantNL bool) (n int64, err error) {
	for s := range out {
		var i int
		i, err = w.Write([]byte(s))
		if err != nil {
			break
		}
		n += int64(i)
		if wantNL {
			_, err = w.Write([]byte("\n"))
			if err != nil {
				break
			}
			n++
		}
	}
	return
}

func (out Out) WriteTo(w io.Writer) (int64, error) {
	return out.writeto(w, withoutNL)
}

func (out Out) PrintLinesTo(w io.Writer) (n int64, err error) {
	return out.writeto(w, withNL)
}

func (out Out) WriteToThenClose(w io.WriteCloser) (int64, error) {
	defer w.Close()
	return out.WriteTo(w)
}

func (out Out) PrintLinesToThenClose(w io.WriteCloser) (int64, error) {
	defer w.Close()
	return out.PrintLinesTo(w)
}
