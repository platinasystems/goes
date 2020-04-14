// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package atsock provides an interface to linux abstract sockets named "@NAME"
package atsock

import (
	"net"
	"net/rpc"
)

type RpcServer struct {
	ln    net.Listener
	conns []net.Conn
}

// Dial streaming socket named "@NAME"
// FIXME retry up to 10 seconds?
func Dial(name string) (net.Conn, error) {
	return net.Dial("unix", "@"+name)
}

// Dial datagram socket named "@NAME"
func DialUnixgram(name string) (*net.UnixConn, error) {
	a, err := net.ResolveUnixAddr("unixgram", "@"+name)
	if err != nil {
		return nil, err
	}
	return net.DialUnix("unixgram", nil, a)
}

// Listen on streaming socket named "@NAME"
func Listen(name string) (net.Listener, error) {
	return net.Listen("unix", "@"+name)
}

// Listen for datagrams on a socket named "@NAME"
func ListenUnixgram(name string) (*net.UnixConn, error) {
	a, err := net.ResolveUnixAddr("unixgram", "@"+name)
	if err != nil {
		return nil, err
	}
	return net.ListenUnixgram("unixgram", a)
}

// Dial an rpcserver named "@NAME"
func NewRpcClient(name string) (*rpc.Client, error) {
	conn, err := Dial(name)
	if err != nil {
		return nil, err
	}
	return rpc.NewClient(conn), nil
}

// Start an rpcserver named "@NAME"
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
