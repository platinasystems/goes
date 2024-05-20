// Copyright © 2021-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mute

import (
	"log"
	"os"
)

const example = "example\n"

func ExampleOffWrite() {
	l := log.New(os.Stdout, "", 0)
	Off(l).Write([]byte(example))
	// Output: example
}

func ExampleOffPrint() {
	l := log.New(os.Stdout, "", 0)
	Off(l).Print(example)
	// Output: example
}

func ExampleOffOnPrint() {
	l := log.New(os.Stdout, "", 0)
	Off(On(l)).Print(example)
	// Output: example
}

func ExampleOffOffPrint() {
	l := log.New(os.Stdout, "", 0)
	Off(Off(l)).Print(example)
	// Output: example
}

func ExampleOffOnOffPrint() {
	l := log.New(os.Stdout, "", 0)
	Off(On(Off(l))).Print(example)
	// Output: example
}

func ExampleOnWrite() {
	l := log.New(os.Stdout, "", 0)
	On(l).Write([]byte(example))
	// Output:
}

func ExampleOnPrint() {
	l := log.New(os.Stdout, "", 0)
	On(l).Print(example)
	// Output:
}

func ExampleOnOffPrint() {
	l := log.New(os.Stdout, "", 0)
	On(Off(l)).Print(example)
	// Output:
}

func ExampleOnOnPrint() {
	l := log.New(os.Stdout, "", 0)
	On(On(l)).Print(example)
	// Output:
}

func ExampleOnOffOnPrint() {
	l := log.New(os.Stdout, "", 0)
	On(Off(On(l))).Print(example)
	// Output:
}

func ExampleIsOff() {
	l := log.New(os.Stdout, "", 0)
	l.Print(IsOn(Off(l)))
	// Output: false
}

func ExampleIsOffOn() {
	l := log.New(os.Stdout, "", 0)
	l.Print(IsOn(Off(On(l))))
	// Output: false
}

func ExampleIsOn() {
	l := log.New(os.Stdout, "", 0)
	l.Print(IsOn(On(l)))
	// Output: true
}

func ExampleIsOnOff() {
	l := log.New(os.Stdout, "", 0)
	l.Print(IsOn(On(Off(l))))
	// Output: true
}
