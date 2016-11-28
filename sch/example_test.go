// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package sch

import "os"

func ExampleString() {
	in, out := New(0)
	go func(in In) {
		defer close(in)
		in <- "hello world\n"
	}(in)
	out.WriteTo(os.Stdout)
	// Output:
	// hello world
}

func ExamplePrintln() {
	in, out := New(0)
	go func(in In) {
		defer close(in)
		in.Println("hello world")
	}(in)
	out.WriteTo(os.Stdout)
	// Output:
	// hello world
}

func ExamplePrint() {
	in, out := New(0)
	go func(in In) {
		defer close(in)
		in.Print("hello world", "\n")
	}(in)
	out.WriteTo(os.Stdout)
	// Output:
	// hello world
}

func ExamplePrintf() {
	in, out := New(0)
	go func(in In) {
		defer close(in)
		in.Printf("%s\n", "hello world")
	}(in)
	out.WriteTo(os.Stdout)
	// Output:
	// hello world
}

func ExampleFile() {
	f, err := os.Open("test.txt")
	if err != nil {
		panic(err)
	}
	in, out := New(3)
	go in.ReadLinesFromThenClose(f)
	out.PrintLinesTo(os.Stdout)
	// Output:
	// Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do
	// eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut
	// enim ad minim veniam, quis nostrud exercitation ullamco laboris
	// nisi ut aliquip ex ea commodo consequat.  Duis aute irure dolor
	// in reprehenderit in voluptate velit esse cillum dolore eu fugiat
	// nulla pariatur. Excepteur sint occaecat cupidatat non proident,
	// sunt in culpa qui officia deserunt mollit anim id est laborum.
}
