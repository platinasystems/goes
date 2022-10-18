// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

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

type Loopback string

// The prefix of a loopback address file is xdg.RunTimeDir()+"/A."
func NewLoopback(s string) Ipc {
	fn := filepath.Join(xdg.RunTimeDir.Value(), fmt.Sprint("A.", s))
	return Ipc{Loopback(fn)}
}

// Returns alocated 127.0.0.1:PORT from Ipc file.
func (lb Loopback) Address() (string, error) {
	addrdata, err := ioutil.ReadFile(lb.String())
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(addrdata)), err
}

func (Loopback) Network() string   { return "tcp" }
func (lb Loopback) String() string { return string(lb) }

// Listen on the the loopback interface (127.0.0.1) at the next available port
// and record the allocated address in a file named by Ipc.
func (lb Loopback) Listen() (net.Listener, error) {
	const create = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	f, err := os.OpenFile(string(lb), create, 0640)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	ln, err := net.Listen(lb.Network(), "127.0.0.1:")
	if err != nil {
		return nil, err
	}
	fmt.Fprintln(f, ln.Addr())
	return LoopbackListener{ln, lb}, err
}

type LoopbackListener struct {
	net.Listener
	lb Loopback
}

func (l LoopbackListener) Close() error {
	err := l.Listener.Close()
	os.Remove(string(l.lb))
	return err
}
