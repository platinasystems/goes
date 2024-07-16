// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import (
	"context"
	"errors"
	"net"
	"os"
	"time"
)

var ShouldRedial = func(err error) bool {
	for _, match := range RedialErrors {
		if errors.Is(err, match) {
			return true
		}
	}
	return false
}

// Redial keeps dialing if error is matched by [ShouldRedial].
func Redial(
	ctx context.Context,
	d net.Dialer,
	network, address string,
) (net.Conn, error) {
	for {
		c, err := d.DialContext(ctx, network, address)
		if err == nil {
			return c, nil
		} else if !ShouldRedial(err) {
			return nil, err
		}
		t := time.NewTimer(1 * time.Second)
		select {
		case <-ctx.Done():
			return nil, context.Canceled
		case <-t.C:
			if !t.Stop() {
				<-t.C
			}
		}
	}
	return nil, os.ErrInvalid
}
