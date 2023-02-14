// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package registry

import (
	"bytes"
	"context"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/container/slice"
	"github.com/platinasystems/goes/v2/pkg/crypto/keycert"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
)

var (
	ErrIncomplete = errors.New("incomplete")
	ErrInvalid    = errors.New("invalid")
	ErrNotFound   = errors.New("not found")
)

var cut = slice.Cut[*keycert.X509]

var reg = struct {
	mutex sync.Mutex
	certs []*keycert.X509
}{}

func Admin(ctx context.Context, path []string, args ...string) error {
	if len(args) == 0 {
		return egress.Marked(ErrIncomplete)
	}
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	approve := path[len(path)-1] == "approve"
	for i, x := range reg.certs {
		if x.Name() == args[0] || x.SKI() == args[0] {
			reg.certs = cut(reg.certs, uint(i), 1)
			if approve {
				err := certs.Subscribers.Append(x)
				if err != nil {
					return egress.Marked(err)
				}
				certs.ClientCAs.Add(x.Certificate)
			}
			return nil
		}
	}
	return fmt.Errorf("%q: %w", args[0], ErrNotFound)
}

func Show(
	ctx context.Context,
	w io.Writer,
	path []string,
	args ...string,
) error {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	for _, x := range reg.certs {
		fmt.Fprint(w, x.SKI(), ": ", x.Certificate.DNSNames, "\n")
	}
	return nil
}

// Register PEM decoded input.
func Subscribe(
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

	blk, _ := pem.Decode(data)
	if blk == nil {
		return egress.Marked(ErrInvalid)
	}
	x := new(keycert.X509)
	err := x.UnmarshalPEM(blk)
	if err != nil {
		return egress.Marked(err)
	}
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	reg.certs = append(reg.certs, x)
	self, err := certs.Self.MarshalPEM()
	if err != nil {
		return egress.Marked(err)
	}
	_, err = w.Write(self)
	return egress.Marked(err)
}
