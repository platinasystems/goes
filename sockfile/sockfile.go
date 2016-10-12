// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package sockfile

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/platinasystems/go/emptych"
	"github.com/platinasystems/go/group"
)

const Dir = "/run/goes/socks"

type RpcServer struct {
	sig chan os.Signal
	emptych.Out
}

func Path(name string) string {
	if strings.HasPrefix(name, Dir) {
		return name
	}
	return filepath.Join(Dir, name)
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
	dir := filepath.Dir(path)
	adm := group.Parse()["adm"].Gid()
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, os.FileMode(0755))
			if adm > 0 {
				os.Chown(dir, os.Geteuid(), adm)
			}
		}
		if err != nil {
			return nil, err
		}
	}
	if _, err := os.Stat(path); err == nil || os.IsExist(err) {
		return nil, fmt.Errorf("%s: %v", name, os.ErrExist)
	}
	ln, err := net.Listen("unix", path)
	if err == nil && adm > 0 {
		os.Chown(path, os.Geteuid(), adm)
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
	done := emptych.Make()
	ln, err := Listen(name)
	if err != nil {
		return nil, err
	}
	srvr := &RpcServer{
		make(chan os.Signal),
		emptych.Out(done),
	}
	signal.Notify(srvr.sig, syscall.SIGTERM)
	go func(done emptych.In) {
		fmt.Println("listen to ", ln.Addr().String())
		for {
			conn, err := ln.Accept()
			if err != nil {
				fn := ln.Addr().String()
				fmt.Println("remove ", fn)
				os.Remove(fn)
				break
			}
			go rpc.ServeConn(conn)
		}
		done.Close()
	}(emptych.In(done))
	go func() {
		for sig := range srvr.sig {
			if sig == syscall.SIGTERM {
				fmt.Println("close ", ln.Addr().String())
				ln.Close()
				break
			}
		}
		close(srvr.sig)
	}()
	return srvr, err
}

func (p *RpcServer) Terminate() {
	p.sig <- syscall.SIGTERM
}
