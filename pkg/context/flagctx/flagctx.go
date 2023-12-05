// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package flagctx

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context"
)

var Parameter = context.NewParameter(flag.CommandLine)

func StringIn(ctx context.Context) string {
	fs := Parameter.In(ctx)
	w := new(strings.Builder)
	fmt.Fprintln(w)
	fs.SetOutput(w)
	fs.PrintDefaults()
	fs.SetOutput(io.Discard)
	return w.String()
}
