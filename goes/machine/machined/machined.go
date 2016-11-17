// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package machined

import (
	"fmt"
	"net/rpc"
	"os"
	"sort"
	"strings"

	"github.com/platinasystems/go/info"
	"github.com/platinasystems/go/redis"
	"github.com/platinasystems/go/redis/rpc/args"
	"github.com/platinasystems/go/redis/rpc/reply"
	"github.com/platinasystems/go/sockfile"
)

const Name = "machined"

// Machines should Plot their info providers into this map.
var Info map[string]info.Interface

func Plot(providers ...info.Interface) {
	if Info == nil {
		Info = make(map[string]info.Interface)
	}
	for _, v := range providers {
		Info[v.String()] = v
	}
}

var Hook = func() error { return nil }

type Attrs map[string]interface{}
type Registry []*entry

type entry struct {
	prefix string
	info   info.Interface
}

type cmd struct {
	registry Registry
	sock     *sockfile.RpcServer
	stop     chan<- struct{}
}

func New() *cmd { return &cmd{} }

func (*cmd) Daemon() int    { return -1 }
func (*cmd) String() string { return Name }
func (*cmd) Usage() string  { return Name }

func (cmd *cmd) Main(args ...string) error {
	var i, n int

	stop := make(chan struct{})
	cmd.stop = stop
	var wait <-chan struct{} = stop

	err := Hook()
	if err != nil {
		return err
	}

	for _, v := range Info {
		n += len(v.Prefixes())
	}
	cmd.registry = make(Registry, n)
	for _, v := range Info {
		for _, prefix := range v.Prefixes() {
			cmd.registry[i] = &entry{prefix, v}
			i++
		}
	}
	sort.Sort(cmd)
	rpc.Register(cmd.registry)
	cmd.sock, err = sockfile.NewRpcServer(Name)
	if err != nil {
		return err
	}
	info.PubCh, err = redis.Publish("platina")
	if err != nil {
		return err
	}
	defer close(info.PubCh)
	for _, entry := range cmd.registry {
		key := fmt.Sprint("platina:", entry.prefix)
		err = redis.Assign(key, Name, "Registry")
		if err != nil {
			return err
		}
	}
	for _, v := range Info {
		go func(v info.Interface) {
			if err := v.Main(); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}(v)
	}
	<-wait
	return nil
}

func (cmd *cmd) Close() error {
	err := cmd.sock.Close()
	for _, v := range Info {
		if v.Close != nil {
			xerr := v.Close()
			if err == nil {
				err = xerr
			}
		}
	}
	close(cmd.stop)
	return err
}

// Interface to sort machine info registry
func (cmd *cmd) Len() int {
	return len(cmd.registry)
}

func (cmd *cmd) Less(i, j int) (t bool) {
	ni, nj := len(cmd.registry[i].prefix), len(cmd.registry[j].prefix)
	switch {
	case ni < nj:
		t = true
	case ni > nj:
		t = false
	case ni == nj:
		t = cmd.registry[i].prefix < cmd.registry[j].prefix
	default:
		panic("oops")
	}
	return t
}

func (cmd *cmd) Swap(i, j int) {
	cmd.registry[i], cmd.registry[j] = cmd.registry[j], cmd.registry[i]
}

func (r Registry) Hdel(args args.Hdel, reply *reply.Hdel) error {
	for _, entry := range r {
		if strings.HasPrefix(args.Field, entry.prefix) {
			if entry.info.Del == nil {
				break
			}
			err := entry.info.Del(args.Field)
			if err == nil {
				*reply = 1
			}
			return err
		}
	}
	return fmt.Errorf("can't delete %s", args.Field)
}

func (r Registry) Hset(args args.Hset, reply *reply.Hset) error {
	value := string(args.Value)
	for _, entry := range r {
		if strings.HasPrefix(args.Field, entry.prefix) {
			if entry.info.Set == nil {
				break
			}
			err := entry.info.Set(args.Field, value)
			if err == nil {
				*reply = 1
			}
			return err
		}
	}
	return fmt.Errorf("can't set %s", args.Field)
}
