// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

/*
Package accumulate provides a wrapper to sum Reader and Writers.
Use it in a WriteTo like this,

	func (t *TYPE) WriteTo(w) (int64, error) {
		acc := accumulate.New(w)
		defer acc.Fini()
		fmt.Fprint(acc, ...)
		...
		fmt.Fprint(acc, ...)
		return acc.Tuple()
	}

An accumulator will skip subsequent read and writes on error.
*/
package accumulate

import (
	"io"
	"runtime"
	"sync"
)

var pool = sync.Pool{
	New: func() interface{} {
		return new(Accumulator)
	},
}

type Accumulator struct {
	n              int64
	err            error
	ReaderOrWriter interface{}
}

// New returns an Accumulator for the given Reader or Writer.
func New(v interface{}) *Accumulator {
	acc := pool.Get().(*Accumulator)
	runtime.SetFinalizer(acc, (*Accumulator).Fini)
	acc.Reset()
	acc.ReaderOrWriter = v
	return acc
}

// Fini returns the Accumulator to its pool.
// If not called by the New() user, it's called by the GC.
func (acc *Accumulator) Fini() {
	runtime.SetFinalizer(acc, nil)
	*acc = Accumulator{}
	pool.Put(acc)
}

// Error records the first non-nil argument, if any, then returns the first
// error encountered by the accumulator.
func (acc *Accumulator) Error(errs ...error) error {
	for _, err := range errs {
		if acc.err == nil {
			acc.err = err
		}
	}
	return acc.err
}

func (acc *Accumulator) Read(b []byte) (int, error) {
	var i int
	if acc.err != nil {
		return 0, acc.err
	}
	i, acc.err = acc.ReaderOrWriter.(io.Reader).Read(b)
	acc.n += int64(i)
	return i, acc.err
}

func (acc *Accumulator) Reset() {
	acc.n = 0
	acc.err = nil
}

// Total adds any given counts to the accumulator and returns the sum.
func (acc *Accumulator) Total(counts ...int) int64 {
	for _, i := range counts {
		if acc.err == nil {
			acc.n += int64(i)
		}
	}
	return acc.n
}

func (acc *Accumulator) Tuple() (int64, error) {
	return acc.n, acc.err
}

func (acc *Accumulator) Write(b []byte) (int, error) {
	var i int
	if acc.err != nil {
		return 0, acc.err
	}
	i, acc.err = acc.ReaderOrWriter.(io.Writer).Write(b)
	acc.n += int64(i)
	return i, acc.err
}

func (acc *Accumulator) WriteString(s string) (int, error) {
	return acc.Write([]byte(s))
}
