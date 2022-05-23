// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build (!ipc_loopback && !plan9)

package ipc

const Network = "unix"

func (ipc Ipc) Address() (string, error) {
	return string(ipc), nil
}
