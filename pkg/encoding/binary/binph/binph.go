// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package binph provides internet protocol-headers.
package binph

type Headers interface {
	AppendTo([]byte) []byte
	PullFrom([]byte) []byte
}

// Fill slice with data and return remaining payload.
func Pull(data, slice []byte) []byte {
	n := copy(slice, data)
	return data[n:]
}
