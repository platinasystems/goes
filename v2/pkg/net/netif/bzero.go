// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

func Bzero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
