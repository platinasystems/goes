// Copyright © 2015-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"io"
	"io/ioutil"
	"strings"
)

var (
	inputMark int
	inputKey  = &inputMark
)

func WithInput(ctx context.Context, r io.Reader) context.Context {
	return Input{ctx, r}
}

type Input struct {
	context.Context
	r io.Reader
}

func InputOf(ctx context.Context) Input {
	if v := ctx.Value(inputKey); v != nil {
		return v.(Input)
	}
	return Input{ctx, nil}
}

func (in Input) Read(buf []byte) (int, error) {
	if err := in.Err(); err != nil {
		return 0, err
	}
	if in.r == nil {
		return 0, io.EOF
	}
	return in.r.Read(buf)
}

func (in Input) Value(k interface{}) interface{} {
	if k == inputKey {
		return in
	}
	return in.Context.Value(k)
}

// Hereby strips `<<< CONTENT` args to push context input.
func Hereby(ctx context.Context, args []string) (context.Context, []string) {
	n := len(args)
	for i, s := range args {
		if s == "<<<" {
			if i == n-1 {
				args = args[:i]
				break
			}
			ctx = WithInput(ctx, strings.NewReader(args[i+1]))
			copy(args[i:], args[i+2:])
			args = args[:n-2]
			break
		}
	}
	return ctx, args
}

// Herein replaces a `<[FILE]` arg with `<<< CONTENT` of FILE or input context.
func Herein(ctx context.Context, args []string) (context.Context, []string) {
	n := len(args)
	for i, s := range args {
		if strings.HasPrefix(s, "<") {
			s = strings.TrimPrefix(s, "<")
			r := func() ([]byte, error) {
				return ioutil.ReadFile(s)
			}
			if len(s) == 0 {
				r = func() ([]byte, error) {
					return ioutil.ReadAll(InputOf(ctx))
				}
			} else if strings.HasPrefix(s, "<") {
				continue
			}
			copy(args[i:], args[i+1:])
			args = args[:n-1]
			if b, err := r(); err == nil {
				args = append(args, "<<<", string(b))
			}
			break
		}
	}
	return ctx, args
}
