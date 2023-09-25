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

func (File) Err() error       { return nil }
func (File) Network() string  { return "unix" }
func (f File) String() string { return string(f) }

// Listen on the Unix file socket named by File.Address().
func (f File) Listen() (net.Listener, error) {
	address := string(f)
	ln, err := net.Listen(f.Network(), address)
	if err == nil {
		err = os.Chmod(address, 0700)
	}
	return ln, err
}
