// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package context

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

var parameterExample = flag.Bool("parameter-example", true, "")

func ExampleParameter() {
	var (
		flags  = NewParameter(flag.CommandLine)
		input  = NewParameter(io.Reader(os.Stdin))
		output = NewParameter(io.Writer(os.Stdout))
		notice = NewParameter(log.New(os.Stdout, "", 0))
		trace  = NewParameter(log.New(os.Stdout, "", log.Lshortfile))
	)
	ctx := input.With(Background(), bytes.NewBufferString("hello world\n"))
	fmt.Println(flags.In(ctx).Lookup("parameter-example").Value)
	io.Copy(output.In(ctx), input.In(ctx))
	const konnichiwa = "こんにちは"
	notice.In(ctx).Println(konnichiwa)
	trace.In(ctx).Println("here")
	// Output:
	// true
	// hello world
	// こんにちは
	// example_parameter_test.go:31: here
}
