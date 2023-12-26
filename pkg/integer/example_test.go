// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package integer

import (
	_ "embed"
	"fmt"
	"strings"
	"sync"
)

//go:embed example.txt
var ExampleText string

var ExampleNames = sync.OnceValue(func() []string {
	return strings.Split(ExampleText, "\n")
})

func ExampleZero() {
	fmt.Println(Name(0, ExampleNames()))
	// Output: zero
}

func ExampleOne() {
	fmt.Println(Name(int16(1), ExampleNames()))
	// Output: one
}

func ExampleTwo() {
	fmt.Println(Name(int32(2), ExampleNames()))
	// Output: two
}

func ExampleThree() {
	fmt.Println(Name(int64(3), ExampleNames()))
	// Output: three
}

func ExampleZeroOneTwoThree() {
	fmt.Println(Names(0xf, ExampleNames()))
	// Output: zero,one,two,three
}

func ExampleAdd() {
	i := 1
	Add(&i, int16(2))
	fmt.Println(Name(i, ExampleNames()))
	// Output: three
}

func ExampleAssign() {
	var i int
	Assign(&i, uint64(3))
	fmt.Println(Name(i, ExampleNames()))
	// Output: three
}

func ExampleDiff() {
	fmt.Println(Diff(0, uint64(1)), Diff(1, uint64(1)), Diff(1, uint64(0)))
	// Output: -1 0 1
}

func ExampleEqual() {
	fmt.Println(Equal(0, uint64(0)), Equal(1, int64(0)))
	// Output: true false
}

func ExampleHas() {
	fmt.Println(Has(3, uint64(2)), Equal(6, int64(2)))
	// Output: true false
}

func ExampleMask() {
	fmt.Println(Mask(7, uint8(3)))
	// Output: 4
}

func ExampleReset() {
	u := 3
	Reset(&u, uint8(1))
	fmt.Println(Name(u, ExampleNames()))
	// Output: two
}

func ExampleSet() {
	var u uint
	Set(&u, uint8(1<<0))
	Set(&u, uint16(1<<1))
	Set(&u, uint32(1<<2))
	Set(&u, uint64(1<<3))
	fmt.Println(Names(u, ExampleNames()))
	// Output: zero,one,two,three
}

func ExampleSub() {
	i := 3
	Sub(&i, int8(2))
	fmt.Println(i)
	// Output: 1
}

func ExampleSum() {
	fmt.Println(Sum(1, int8(2)))
	// Output: 3
}

func ExampleToggle() {
	i := 7
	Toggle(&i, int8(2))
	fmt.Println(i)
	// Output: 5
}
