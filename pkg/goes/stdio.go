// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"os"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/parameter"
)

func ContextPrintln(ctx context.Context, s string) error {
	s = strings.TrimRight(s, nl)
	if len(s) == 0 {
		return nil
	}
	w := ContextStdout(ctx)
	_, err := w.WriteString(s)
	if err == nil {
		_, err = w.WriteString(nl)
	}
	return err
}

func ContextStderr(ctx context.Context) *os.File {
	return parameter.Value(ctx, &os.Stderr)
}

func StderrContext(ctx context.Context, f *os.File) context.Context {
	return parameter.Context(ctx, &os.Stderr, f)
}

func ContextStdin(ctx context.Context) *os.File {
	return parameter.Value(ctx, &os.Stdin)
}

func StdinContext(ctx context.Context, f *os.File) context.Context {
	return parameter.Context(ctx, &os.Stdin, f)
}

func ContextStdout(ctx context.Context) *os.File {
	return parameter.Value(ctx, &os.Stdout)
}

func StdoutContext(ctx context.Context, f *os.File) context.Context {
	return parameter.Context(ctx, &os.Stdout, f)
}
