// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package parameter

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
)

var example = flag.Bool("example", true, "FlagSet parameter")

func Example() {
	var (
		fs     = Parameter[*flag.FlagSet]{flag.CommandLine}
		input  = Parameter[io.Reader]{os.Stdin}
		output = Parameter[io.Writer]{os.Stdout}
	)
	ctx := Background()
	ctx = fs.With(ctx)
	ctx = input.With(ctx, bytes.NewBufferString("hello world\n"))
	ctx = output.With(ctx)
	io.Copy(output.Within(ctx), input.Within(ctx))
	fmt.Println(fs.Within(ctx).Lookup("example").Value)
	// Output:
	// hello world
	// true
}
