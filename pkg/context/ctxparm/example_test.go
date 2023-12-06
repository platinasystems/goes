// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package ctxparm

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
)

var example = flag.Bool("example", true, "")

func ExampleParameter() {
	ctx := context.Background()
	ctx = Reader.With(ctx, bytes.NewBufferString("hello world\n"))
	ctx = Writer.With(ctx, os.Stdout)
	io.Copy(Writer.In(ctx), Reader.In(ctx))
	fmt.Println(Flags.In(ctx).Lookup("example").Value)
	// Output:
	// hello world
	// true
}
