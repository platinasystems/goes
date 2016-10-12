// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package emptych provides an empty struct channel with encapsulated endpoints
// that are commonly used to stop go routines.
package emptych

type Empty struct{}

type In chan<- Empty
type Out <-chan Empty

func Make() chan Empty {
	return make(chan Empty)
}

func New() (In, Out) {
	ch := Make()
	return In(ch), Out(ch)
}

func (in In) Close() error {
	close(in)
	return nil
}

func (out Out) Wait() {
	<-out
}
