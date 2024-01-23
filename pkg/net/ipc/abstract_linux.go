// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import (
	"errors"
	"net"
	"os"
)

type Abstract string

// The prefix of a Linux abstract domain socket is "@".
func NewAbstract(suffix ...any) Ipc {
	return Ipc{Abstract(join("@", suffix))}
}

func (Abstract) Err() error         { return nil }
func (Abstract) Network() string    { return "unix" }
func (abs Abstract) String() string { return string(abs) }

func (abs Abstract) Listen() (net.Listener, error) {
	ln, err := net.Listen(abs.Network(), abs.String())
	if err != nil {
		return nil, err
	}
	return AbstractListener{ln}, nil
}

type AbstractListener struct {
	net.Listener
}

func (ln AbstractListener) Accept() (conn net.Conn, err error) {
	var peer *unix.Ucred
	my := unix.Ucred{
		Uid: uint32(os.Geteuid()),
		Gid: uint32(os.Getgid()),
	}
	if conn, err = ln.Listener.Accept(); err != nil {
		return
	}
	defer func() {
		if err != nil {
			conn.Close()
			conn = nil
		}
	}()
	rc, err := conn.(*net.UnixConn).SyscallConn()
	if err != nil {
		return
	}
	cerr := rc.Control(func(fd uintptr) {
		peer, err = unix.GetsockoptUcred(int(fd), unix.SOL_SOCKET,
			unix.SO_PEERCRED)
	})
	if err == nil {
		if err = cerr; err == nil {
			if peer.Uid != my.Uid && peer.Gid != my.Gid {
				err = errors.New("unauthorized")
			}
		}
	}
	return
}
