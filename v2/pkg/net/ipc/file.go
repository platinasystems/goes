// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import (
	"net"
	"os"
	"path/filepath"
)

type File string

// A Unix file socket named Dir()+"/S."+suffix
func NewFile(suffix ...any) Ipc {
	return Ipc{File(filepath.Join(Dir(), join("S.", suffix)))}
}

func (file File) Address() (string, error) { return string(file), nil }

// Listen on the Unix file socket named by File.Address().
func (file File) Listen() (net.Listener, error) {
	address := string(file)
	ln, err := net.Listen(file.Network(), address)
	if err == nil {
		err = os.Chmod(address, 0700)
	}
	return ln, err
}

func (File) Network() string { return "unix" }
