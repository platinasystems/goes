// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !ipc_abstract && !ipc_unix && (ipc_loopback || plan9)

package ipc

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/os/xdg"
)

const Network = "tcp"

// The prefix of a loopback address file is xdg.RunTimeDir()+"/A."
func Prefix() string {
	return filepath.Join(xdg.RunTimeDir.Value(), "A.")
}

// Returns alocated 127.0.0.1:PORT from Ipc file.
func (ipc Ipc) Address() (string, error) {
	addrdata, err := ioutil.ReadFile(string(ipc))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(addrdata)), err
}

// Listen on the the loopback interface (127.0.0.1) at the next available port
// and record the allocated address in a file named by Ipc.
func (ipc Ipc) Listen() (net.Listener, error) {
	const creatf = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	f, err := os.OpenFile(string(ipc), createf, 0640)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	ln, err := net.Listen(IpcNetwork, "127.0.0.1:")
	if err != nil {
		return nil, err
	}
	fmt.Fprintln(f, ln.Addr())
	return loopback{ln, fn}, err
}

type loopback struct {
	net.Listener
	fn string
}

func (l loopback) Close() error {
	err := l.Listener.Close()
	os.Remove(l.fn)
}
