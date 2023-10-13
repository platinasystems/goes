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

	"github.com/platinasystems/goes/v2/pkg/crypto/xcert"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
)

var reg struct {
	sync.RWMutex
	x *xcert.X509
}

func regAdmin(ctx context.Context, path []string, args ...string) error {
	if len(args) == 0 {
		return egress.Marked(ErrIncomplete)
	}
	approve := path[len(path)-1] == "approve"
	reg.Lock()
	defer reg.Unlock()
	for cur, prev := reg.x, reg.x; cur != nil; cur = cur.Next {
		if cur.IsMatch(args[0]) {
			if approve {
				Subscribers().Append(cur)
				ClientCAs().Add(cur.Certificate)
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

func regShow(
	ctx context.Context,
	w io.Writer,
	path []string,
	args ...string,
) error {
	reg.RLock()
	defer reg.RUnlock()
	reg.x.Range(func(x *xcert.X509) bool {
		fmt.Fprint(w, x.SKI(), ": ", x.Certificate.DNSNames, "\n")
		return true
	})
	return nil
}

// Register PEM decoded input.
func regSubscribe(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	var data []byte

	if len(args) == 0 {
		return egress.Marked(ErrIncomplete)
	}
	if args[0] == "-" {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, r); err != nil {
			return egress.Marked(err)
		}
		data = buf.Bytes()
	} else {
		data = []byte(args[0])
	}
	x, err := xcert.NewX509(data)
	if err != nil {
		return egress.Marked(err)
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
	self, err := Selfie.MarshalPEM()
	if err != nil {
		return egress.Marked(err)
	}
	_, err = w.Write(self)
	return egress.Marked(err)
}
