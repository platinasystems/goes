// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package stringd

import (
	"fmt"
	"net/rpc"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/redis"
	"github.com/platinasystems/redis/publisher"
	"github.com/platinasystems/redis/rpc/args"
	"github.com/platinasystems/redis/rpc/reply"
	"github.com/platinasystems/go/internal/sockfile"
)

const (
	Name    = "stringd"
	Apropos = "provides a redis settable test string"
	Usage   = "stringd"

	pubkey = "test.string"
)

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}

func New() Command { return make(Command) }

type Command chan struct{}

type Stringd struct {
	s   string
	pub *publisher.Publisher
}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Kind() cmd.Kind    { return cmd.Daemon }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (c Command) Main(...string) error {
	err := redis.IsReady()
	if err != nil {
		return err
	}
	pub, err := publisher.New()
	if err != nil {
		return err
	}
	defer pub.Close()

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
	pub.Print(pubkey, ": ", stringd.s)
	<-c
	return nil
}

func (c Command) Close() error {
	defer close(c)
	return nil
}

func (stringd *Stringd) Hset(args args.Hset, reply *reply.Hset) error {
	stringd.s = string(args.Value)
	stringd.pub.Print(pubkey, ": ", stringd.s)
	*reply = 1
	return nil
}
