// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cat

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/context/read"
	"github.com/platinasystems/goes/v2/pkg/context/write"
)

func Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	switch path[1] {
	case "complete":
		return nil
	case "help":
		copy(path[1:], path[2:])
		path = path[:len(path)-1]
		return template.Must(template.New("usage").Parse(`
usage: {{.}} [<file(s)>|-]
Concatenate file(s) or standard in (-) to output.
`[1:])).Execute(w, strings.Join(path, " "))
	}
	if len(args) == 0 {
		args = append(args, "-")
	}
	cr := read.With(ctx, r)
	cw := write.With(ctx, w)
	for _, fn := range args {
		if fn == "-" {
			io.Copy(cw, cr)
		} else if fi, err := os.Stat(fn); err == nil && fi.IsDir() {
			return fmt.Errorf("%s: is a directory", fn)
		} else if f, err := os.Open(fn); err == nil {
			io.Copy(cw, f)
			f.Close()
		} else {
			return fmt.Errorf("%s: %w", fn, errors.Unwrap(err))
		}
	}
	return ctx.Err()
}
