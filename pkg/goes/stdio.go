// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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

// If name is empty or '-' this returns ContextStdin(ctx); otherwise,
// it's the opened the named file.
func ContextInput(ctx context.Context, name string) (*os.File, error) {
	if len(name) == 0 || name == "-" {
		return ContextStdin(ctx), nil
	}
	return os.Open(name)
}

// If filename is empty or '-' this returns ContextStdout(ctx); otherwise,
// this returns the created or truncated file.
func ContextOutput(
	ctx context.Context,
	filename string,
	mode os.FileMode,
) (*os.File, error) {
	const create = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if len(filename) == 0 || filename == "-" {
		return ContextStdout(ctx), nil
	}
	if dir := filepath.Dir(filename); dir != "." {
		if _, err := os.Stat(dir); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
			dmode := 0700 | mode
			if (mode & 0060) == 0060 {
				dmode |= 0010
			}
			if (mode & 0006) == 0006 {
				dmode |= 0001
			}
			if err = os.MkdirAll(dir, dmode); err != nil {
				return nil, err
			}
		}
	}
	return os.OpenFile(filename, create, mode)
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
