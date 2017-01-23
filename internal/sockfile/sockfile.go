// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package sockfile creates and dials servers listening to unix socket files in
// /run/goes/socks.
package sockfile

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"time"

	"github.com/platinasystems/go/internal/group"
	"github.com/platinasystems/go/internal/varrun"
)

const Dir = varrun.Dir + "/socks"

type RpcServer struct {
	ln    net.Listener
	conns []net.Conn
}

// Wait up to 10 seconds for the socket then set it to the named group and
// enable group writes.
func Chgroup(path, name string) (err error) {
	fi, err := os.Stat(path)
	if err != nil {
		return
	}
	if gid := group.Parse()[name].Gid(); gid != 0 {
		err = os.Chown(path, os.Geteuid(), gid)
		if err != nil {
			return
		}
	}
	err = os.Chmod(path, fi.Mode()|0020)
	return
}

func Dial(name string) (net.Conn, error) {
	path := Path(name)
	err := WaitFor(path, time.Now().Add(10*time.Second))
	if err != nil {
		return nil, err
	}
	return net.Dial("unix", path)
}

func DialUnixgram(name string) (*net.UnixConn, error) {
	path := Path(name)
	err := WaitFor(path, time.Now().Add(10*time.Second))
	if err != nil {
		return nil, err
	}
	a, err := net.ResolveUnixAddr("unixgram", path)
	if err != nil {
		return nil, err
	}
	return net.DialUnix("unixgram", nil, a)
}

func Listen(name string) (net.Listener, error) {
	path := Path(name)
	err := varrun.New(Dir)
	if err != nil {
		return nil, err
	}
	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	err = WaitFor(path, time.Now().Add(10*time.Second))
	if err == nil {
		err = Chgroup(path, "adm")
	}
	if err != nil {
		ln.Close()
		ln = nil
	}
	return ln, err
}

func ListenUnixgram(name string) (*net.UnixConn, error) {
	path := Path(name)
	err := varrun.New(Dir)
	if err != nil {
		return nil, err
	}
	a, err := net.ResolveUnixAddr("unixgram", path)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUnixgram("unixgram", a)
	if err != nil {
		return nil, err
	}
	err = WaitFor(path, time.Now().Add(10*time.Second))
	if err == nil {
		err = Chgroup(path, "adm")
	}
	if err != nil {
		conn.Close()
		conn = nil
	}
	return conn, err
}

func NewRpcClient(name string) (*rpc.Client, error) {
	conn, err := Dial(name)
	if err != nil {
		return nil, err
	}
	return rpc.NewClient(conn), nil
}

func NewRpcServer(name string) (*RpcServer, error) {
	ln, err := Listen(name)
	if err != nil {
		return nil, err
	}
	srvr := &RpcServer{ln: ln}
	go srvr.listen()
	return srvr, err
}

// Path returns Dir + "/" + Dir(name) if name isn't already prefaced by Dir
func Path(name string) string {
	if filepath.Dir(name) != Dir {
		name = filepath.Join(Dir, filepath.Base(name))
	}
	return name
}

func RemoveAll() {
	socks, err := filepath.Glob(filepath.Join(Dir, "*"))
	if err == nil {
		for _, fn := range socks {
			os.Remove(fn)
		}
		os.Remove(Dir)
	}
}

func WaitFor(path string, timeout time.Time) (err error) {
	for {
		_, err = os.Stat(path)
		if err == nil {
			break
		}
		if time.Now().After(timeout) {
			err = fmt.Errorf("%s: timeout", path)
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
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
	fn := srvr.ln.Addr().String()
	for {
		conn, err := srvr.ln.Accept()
		if err != nil {
			break
		}
		srvr.conns = append(srvr.conns, conn)
		go rpc.ServeConn(conn)
	}
	os.Remove(fn)
}
