// Copyright 2016-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package daemons

import (
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/external/atsock"
	"github.com/platinasystems/goes/lang"
)

type Server struct {
	// Machines list goes command + args for daemons that run from start,
	// including redisd.  Note that dependent daemons should wait on a
	// respective redis key, e.g.
	//	redis.Hwait(redis.DefaultHash, "redis.ready", "true", TIMEOUT)
	// or
	//	redis.IsReady()
	Init [][]string
	Daemons
}

func (*Server) String() string { return "goes-daemons" }

func (*Server) Usage() string {
	return "goes-daemons [OPTIONS]..."
}

func (*Server) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "start daemons and wait for their exit",
	}
}

func (c *Server) Goes(g *goes.Goes) { c.Daemons.goes = g }

func (*Server) Kind() cmd.Kind { return cmd.Hidden }

func (c *Server) Main(args ...string) error {
	var err error

	c.Daemons.init()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGTERM)
	defer signal.Stop(sig)

	c.rpc, err = atsock.NewRpcServer(sockname())
	if err != nil {
		return err
	}
	defer c.rpc.Close()

	for _, dargs := range c.Init {
		c.Daemons.start(0, dargs...)
	}

	rpc.Register(&c.Daemons)

	for {
		select {
		case <-c.Daemons.done:
			// delay for rpc Stop reply
			time.Sleep(100 * time.Millisecond)
			return nil
		case <-sig:
			c.Daemons.Stop([]int{}, &empty)
		}
	}
}
