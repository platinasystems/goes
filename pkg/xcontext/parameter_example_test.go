// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package xcontext

import (
	. "context"
	"flag"
	"fmt"
	"io"
	"os"
)

func ExampleParameter() {
	var exflags flag.FlagSet
	exflags.Bool("example", true, "")
	if err := exflags.Parse(flag.Args()); err != nil {
		return
	}
	ctx := Background()
	ctx = ContextParameter(ctx, &flag.CommandLine, &exflags)
	if pr, pw, err := os.Pipe(); err != nil {
		panic(err)
	} else {
		ctx = ContextParameter(ctx, &os.Stdin, pr)
		fmt.Fprintln(pw, "hello world")
		pw.Close()
	}

	in := ParameterValue(ctx, &os.Stdin)
	out := ParameterValue(ctx, &os.Stdout)
	exv := ParameterValue(ctx, &flag.CommandLine).Lookup("example").Value

	io.Copy(out, in)
	fmt.Fprintln(out, exv)

	// Output:
	// hello world
	// true
}
