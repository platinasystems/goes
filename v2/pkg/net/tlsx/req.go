// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/os/page"
)

const WithInput = "<<<<"

// Send request through connection in context; then write the response to the
// context output and return any error.
//
// The request is made with successive link-value encoded args; then any
// "<<<" prefaced input data; and concludes with an empty value (break).
//
// The response output is the decoded received data up to break; then returns
// following error or nil if break.
//
//	> REQ [ARG]... ["<<<<" INPUT...] ""
//	< OUTPUT... {NACK | ""}
func Req(
	ctx context.Context,
	conn *tls.Conn,
	r io.Reader,
	w io.Writer,
	args ...any,
) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	ictx, cancel := context.WithCancel(context.Background())
	defer cancel()

	enc := lv.NewEncoder(write.With(ctx, conn))

	err := req(enc, args...)
	if err != nil {
		return err
	}

	if r == nil {
		enc.Break()
	} else {
		enc.WriteString(WithInput)
		ienc := lv.NewEncoder(write.With(ictx, conn))
		wg.Add(1)
		go func() {
			defer wg.Done()
			ib := page.New()
			defer page.Free(ib)
			for {
				select {
				case <-ictx.Done():
					return
				default:
				}
				n, err := r.Read(ib)
				if n == 0 {
					ienc.Break()
					break
				}
				_, err = ienc.Write(ib[:n])
				if err != nil {
					break
				}
			}
		}()
	}

	dec := lv.NewDecoder(poll.With(ctx, conn))
	ob := page.New()
	defer page.Free(ob)

	for {
		if n, err := dec.Read(ob); err != nil {
			return err
		} else if n == 0 {
			break
		} else {
			w.Write(ob[:n])
		}
	}

	return nil
}

func ack(enc lv.Encode, args ...any) error {
	return req(enc, append(args, nil))
}

// recurse or iterate if args contains []any or []string.
func req(enc lv.Encode, args ...any) error {
	for _, arg := range args {
		switch t := arg.(type) {
		case nil:
			return enc.Break()
		case []any:
			if err := req(enc, t...); err != nil {
				return err
			}
		case []string:
			for _, s := range t {
				_, err := enc.WriteString(s)
				if err != nil {
					return err
				}
			}
		default:
			_, err := fmt.Fprint(enc, arg)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
