// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package sockfile

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"time"

	"github.com/platinasystems/go/group"
	"github.com/platinasystems/go/rundir"
)

const Dir = "/run/goes/socks"

type RpcServer struct {
	ln    net.Listener
	conns []net.Conn
}

func Dial(name string) (net.Conn, error) {
	path := Path(name)
	for i := 0; ; i++ {
		conn, err := net.Dial("unix", path)
		if err == nil {
			return conn, err
		}
		if i == 30 {
			return nil, err
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func Listen(name string) (net.Listener, error) {
	path := Path(name)
	_, err := os.Stat(path)
	if err == nil || os.IsExist(err) {
		return nil, fmt.Errorf("%s: %v", name, os.ErrExist)
	}
	err = rundir.New(Dir)
	if err != nil {
		return nil, err
	}
	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	if adm := group.Parse()["adm"].Gid(); adm > 0 {
		err = os.Chown(path, os.Geteuid(), adm)
	}
	return ln, err
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
