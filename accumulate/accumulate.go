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
		return acc.N, acc.Err
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
	ReaderOrWriter interface{}
	// Err contains the first error encountered by Read or Write.
	Err error
	// Total Read or Writen
	N int64
}

// New returns an Accumulator for the given Reader or Writer.
func New(v interface{}) *Accumulator {
	a := pool.Get().(*Accumulator)
	runtime.SetFinalizer(a, (*Accumulator).Fini)
	a.ReaderOrWriter = v
	return a
}

// Fini returns the Accumulator to its pool.
// If not called by the New() user, it's called by the GC.
func (a *Accumulator) Fini() {
	runtime.SetFinalizer(a, nil)
	*a = Accumulator{}
	pool.Put(a)
}

func (a *Accumulator) Read(b []byte) (int, error) {
	var i int
	if a.Err != nil {
		return 0, a.Err
	}
	i, a.Err = a.ReaderOrWriter.(io.Reader).Read(b)
	a.N += int64(i)
	return i, a.Err
}

func (a *Accumulator) Write(b []byte) (int, error) {
	var i int
	if a.Err != nil {
		return 0, a.Err
	}
	i, a.Err = a.ReaderOrWriter.(io.Writer).Write(b)
	a.N += int64(i)
	return i, a.Err
}
