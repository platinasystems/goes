// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !ipc_loopback && !ipc_unix_file && (ipc_unix_abstract || linux)

package ipc

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// The prefix of a Unix abstract domain socket is "@".
func Prefix() string { return "@" }

func (ipc Ipc) Listen() (net.Listener, error) {
	ln, err := net.Listen(Network, string(ipc))
	if err != nil {
		return nil, err
	}
	return abstract{ln}, nil
}

type abstract struct {
	net.Listener
}

func (ln abstract) Accept() (conn net.Conn, err error) {
	var peer *syscall.Ucred
	my := syscall.Ucred{
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
		peer, err = syscall.GetsockoptUcred(int(fd),
			syscall.SOL_SOCKET, syscall.SO_PEERCRED)
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
