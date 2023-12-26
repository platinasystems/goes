// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package parameter

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
)

func Example() {
	var exflags flag.FlagSet
	exflags.Bool("example", true, "")
	if err := exflags.Parse(flag.Args()); err != nil {
		return
	}
	ctx := context.Background()
	ctx = Context(ctx, &flag.CommandLine, &exflags)
	if pr, pw, err := os.Pipe(); err != nil {
		panic(err)
	} else {
		ctx = Context(ctx, &os.Stdin, pr)
		fmt.Fprintln(pw, "hello world")
		pw.Close()
	}

	in := Value(ctx, &os.Stdin)
	out := Value(ctx, &os.Stdout)
	exv := Value(ctx, &flag.CommandLine).Lookup("example").Value

	io.Copy(out, in)
	fmt.Fprintln(out, exv)

	// Output:
	// hello world
	// true
}
