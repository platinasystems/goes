// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package atsock creates and dials servers listening to linux abstract sockets
// named "@MACHINE.NAME"
package atsock

import (
	"fmt"
	"net"
	"net/rpc"

	"github.com/platinasystems/go/internal/machine"
)

type RpcServer struct {
	ln    net.Listener
	conns []net.Conn
}

// Name returns "@MACHINE/NAME"
func Name(name string) string {
	return fmt.Sprint("@", machine.Name, "/", name)
}

// Dial streaming socket named "@MACHINE/NAME"
// FIXME retry up to 10 seconds?
func Dial(name string) (net.Conn, error) {
	return net.Dial("unix", Name(name))
}

// Dial datagram socket named "@MACHINE/NAME"
func DialUnixgram(name string) (*net.UnixConn, error) {
	a, err := net.ResolveUnixAddr("unixgram", Name(name))
	if err != nil {
		return nil, err
	}
	return net.DialUnix("unixgram", nil, a)
}

// Listen on streaming socket named "@MACHINE/NAME"
func Listen(name string) (net.Listener, error) {
	return net.Listen("unix", Name(name))
}

// Listen for datagrams on a socket named "@MACHINE/NAME"
func ListenUnixgram(name string) (*net.UnixConn, error) {
	a, err := net.ResolveUnixAddr("unixgram", Name(name))
	if err != nil {
		return nil, err
	}
	return net.ListenUnixgram("unixgram", a)
}

// Dial an rpcserver named "@MACHINE/NAME"
func NewRpcClient(name string) (*rpc.Client, error) {
	conn, err := Dial(name)
	if err != nil {
		return nil, err
	}
	return rpc.NewClient(conn), nil
}

// Start an rpcserver named "@MACHINE/NAME"
func NewRpcServer(name string) (*RpcServer, error) {
	ln, err := Listen(name)
	if err != nil {
		return nil, err
	}
	srvr := &RpcServer{ln: ln}
	go srvr.listen()
	return srvr, err
}

func (srvr *RpcServer) Close() error {
	err := srvr.ln.Close()
	for _, conn := range srvr.conns {
		xerr := conn.Close()
		if err == nil {
			err = xerr
		}
	}
	return err
}

func (srvr *RpcServer) listen() {
	for {
		conn, err := srvr.ln.Accept()
		if err != nil {
			break
		}
		srvr.conns = append(srvr.conns, conn)
		go rpc.ServeConn(conn)
	}
}
