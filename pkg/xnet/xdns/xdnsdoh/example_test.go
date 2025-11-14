// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsdoh

import (
	"context"
	"fmt"
)

func ExampleLookupAddrPort() {
	ctx := context.Background()
	for _, x := range []struct{ nw, s string }{
		{"tcp", "127.0.0.1"},
		{"tcp4", "localhost"},
		{"tcp", "::1"},
		{"tcp", "127.0.0.1:80"},
		{"tcp", "[::1]:80"},
		{"tcp", "127.0.0.1:http"},
		{"tcp", "[::1]:http"},
		{"tcp4", "localhost:http"},
	} {
		if aps, err := LookupAddrPort(ctx, x.nw, x.s); err != nil {
			fmt.Println(err)
		} else {
			for _, ap := range aps {
				fmt.Println(ap)
			}
		}
	}
	// Output:
	// 127.0.0.1:0
	// 127.0.0.1:0
	// [::1]:0
	// 127.0.0.1:80
	// [::1]:80
	// 127.0.0.1:80
	// [::1]:80
	// 127.0.0.1:80
}
