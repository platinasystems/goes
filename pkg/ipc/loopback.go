// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type Loopback struct {
	fn  string
	err error
}

// A loopback address file named Dir()+"/A."+suffix
func NewLoopback(suffix ...any) Ipc {
	return Ipc{&Loopback{
		filepath.Join(Dir(), join("A.", suffix)),
		nil,
	}}
}

// Returns any error trying to read file containing allocated address:port
func (lb *Loopback) Err() error { return lb.err }

// Listen on the the loopback interface (127.0.0.1) at the next available port
// and record the allocated address in a file named by Ipc.
func (lb *Loopback) Listen() (net.Listener, error) {
	const create = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	f, err := os.OpenFile(lb.fn, create, 0640)
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

func (*Loopback) Network() string { return "tcp" }

// Returns allocated 127.0.0.1:PORT from Ipc file.
func (lb *Loopback) String() string {
	addrdata, err := os.ReadFile(lb.fn)
	if err != nil {
		lb.err = err
		return ""
	}
	return strings.TrimSpace(string(addrdata))
}

type LoopbackListener struct {
	net.Listener
	lb *Loopback
}

func (l LoopbackListener) Close() error {
	err := l.Listener.Close()
	os.Remove(string(l.lb.fn))
	return err
}
