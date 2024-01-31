// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/goes"
)

var reg struct {
	sync.RWMutex
	x *X509
}

func regAdmin(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return egress.Mark(ErrIncomplete)
	}
	if err := InitClientCAs(); err != nil {
		return err
	}
	branch := goes.ContextBranch(ctx)
	approve := branch[len(branch)-1] == "approve"
	reg.Lock()
	defer reg.Unlock()
	for cur, prev := reg.x, reg.x; cur != nil; cur = cur.Next {
		if cur.IsMatch(args[0]) {
			if approve {
				Subscribers.Append(cur)
				ClientCAs.Add(cur.Certificate)
			}
			if cur == prev {
				reg.x = cur.Next
			} else {
				prev.Next = cur.Next
			}
			return nil
		}
		prev = cur
	}
	return fmt.Errorf("%q: %w", args[0], ErrNotFound)
}

func regShow(ctx context.Context, args []string) error {
	reg.RLock()
	defer reg.RUnlock()
	w := goes.ContextStdout(ctx)
	reg.x.Range(func(x *X509) bool {
		fmt.Fprint(w, x.SKI(), ": ", x.Certificate.DNSNames, "\n")
		return true
	})
	return nil
}

// Register PEM decoded input.
func regSubscribe(ctx context.Context, args []string) error {
	var data []byte

	r := goes.ContextStdin(ctx)
	if len(args) == 0 {
		return egress.Mark(ErrIncomplete)
	}
	if args[0] == "-" {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, r); err != nil {
			return egress.Mark(err)
		}
		data = buf.Bytes()
	} else {
		data = []byte(args[0])
	}
	err := InitSelf()
	if err != nil {
		return err
	}
	x, err := NewX509(data)
	if err != nil {
		return egress.Mark(err)
	}
	func() {
		reg.Lock()
		defer reg.Unlock()
		if reg.x == nil {
			reg.x = x
		} else {
			reg.x.Append(x)
		}
	}()
	selfpem, err := Self.MarshalPEM()
	if err != nil {
		return err
	}
	_, err = goes.ContextStdout(ctx).Write(selfpem)
	return egress.Mark(err)
}
