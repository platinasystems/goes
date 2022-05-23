// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !ipc_loopback && !ipc_unix_abstract && (ipc_unix_file || (!linux && !plan9))

package ipc

import (
	"net"
	"os"
	"path/filepath"

	"github.com/platinasystems/goes/v2/pkg/os/xdg"
)

// The prefix of a Unix file socket is xdg.RunTimeDir+"/S."
func Prefix() string {
	return filepath.Join(xdg.RunTimeDir.String(), "S.")
}

// Listen on the Unix file socket named by Ipc.Address().
func (ipc Ipc) Listen() (net.Listener, error) {
	ln, err := net.Listen(Network, string(ipc))
	if err == nil {
		err = os.Chmod(string(ipc), 0770)
	}
	return ln, err
}
