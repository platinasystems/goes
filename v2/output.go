// Copyright © 2015-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"bytes"
	"context"
	"fmt"
	"io"
)

var (
	outputMark int
	outputKey  = &outputMark
)

func WithOutput(ctx context.Context, w io.Writer) context.Context {
	return Output{ctx, w}
}

type Output struct {
	context.Context
	w io.Writer
}

func OutputOf(ctx context.Context) Output {
	if v := ctx.Value(outputKey); v != nil {
		return v.(Output)
	}
	return Output{ctx, nil}
}

func (o Output) Print(args ...interface{}) {
	if o.Err() != nil || o.w == nil {
		return
	}
	fmt.Fprint(o.w, args...)
}

func (o Output) Printf(format string, args ...interface{}) {
	if o.Err() != nil || o.w == nil {
		return
	}
	fmt.Fprintf(o.w, format, args...)
}

func (o Output) Println(args ...interface{}) {
	if o.Err() != nil || o.w == nil {
		return
	}
	fmt.Fprintln(o.w, args...)
}

func (o Output) ReadFrom(r io.Reader) (n int64, err error) {
	if err = o.Err(); err != nil {
		return n, err
	}
	pg := Page.Get().([]byte)
	defer Page.Put(pg)
	for {
		nr, rerr := r.Read(pg)
		if err = o.Err(); err != nil {
			break
		}
		if nr > 0 {
			var nw int
			if o.w != nil {
				if nw, err = o.w.Write(pg[:nr]); err != nil {
					break
				}
			} else {
				nw = nr
			}
			n += int64(nw)
			if err = io.ErrShortWrite; nr != nw {
				break
			}
		}
		if err = rerr; err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
	}
	return
}

func (o Output) Value(k interface{}) interface{} {
	if k == outputKey {
		return o
	}
	return o.Context.Value(k)
}

func (o Output) Write(data []byte) (int, error) {
	err := o.Err()
	if err != nil {
		return 0, err
	}
	if o.w == nil {
		return len(data), nil
	}
	n, err := o.ReadFrom(bytes.NewReader(data))
	return int(n), err
}
