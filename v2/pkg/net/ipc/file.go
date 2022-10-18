// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/platinasystems/goes/v2/pkg/os/xdg"
)

type File string

// The prefix of a Unix file socket is xdg.RunTimeDir+"/S."
func NewFile(s string) Ipc {
	fn := filepath.Join(xdg.RunTimeDir.String(), fmt.Sprint("S.", s))
	return Ipc{File(fn)}
}

func (file File) Address() (string, error) { return file.String(), nil }

// Listen on the Unix file socket named by File.Address().
func (file File) Listen() (net.Listener, error) {
	address := file.String()
	ln, err := net.Listen(file.Network(), address)
	if err == nil {
		err = os.Chmod(address, 0770)
	}
	return ln, err
}

func (File) Network() string { return "unix" }

func (file File) String() string { return string(file) }
