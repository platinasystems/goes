// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package stringd

import (
	"fmt"
	"net/rpc"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/internal/redis"
	"github.com/platinasystems/go/goes/internal/redis/rpc/args"
	"github.com/platinasystems/go/goes/internal/redis/rpc/reply"
	"github.com/platinasystems/go/goes/internal/sockfile"
)

const Name = "stringd"
const pubkey = "test.string"

type cmd chan struct{}

type Stringd struct {
	s   string
	pub chan<- string
}

func New() cmd { return cmd(make(chan struct{})) }

func (cmd) Kind() goes.Kind { return goes.Daemon }
func (cmd) String() string  { return Name }
func (cmd) Usage() string   { return Name }

func (cmd cmd) Main(...string) error {
	pub, err := redis.Publish(redis.DefaultHash)
	if err != nil {
		return err
	}
	defer close(pub)
	stringd := &Stringd{
		s:   "hello world",
		pub: pub,
	}
	rpc.Register(stringd)
	sock, err := sockfile.NewRpcServer(Name)
	if err != nil {
		return err
	}
	defer sock.Close()
	key := fmt.Sprintf("%s:%s", redis.DefaultHash, pubkey)
	err = redis.Assign(key, Name, "Stringd")
	if err != nil {
		return err
	}
	stringd.pub <- fmt.Sprint(pubkey, ": ", stringd.s)
	<-cmd
	return nil
}

func (cmd cmd) Close() error {
	defer close(cmd)
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "provides a redis settable test string",
	}
}

func (stringd *Stringd) Hset(args args.Hset, reply *reply.Hset) error {
	stringd.s = string(args.Value)
	stringd.pub <- fmt.Sprint(pubkey, ": ", stringd.s)
	*reply = 1
	return nil
}
