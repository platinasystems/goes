// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package reg provides an RPC to register redis handlers.
package reg

import (
	"github.com/platinasystems/go/redis/rpc"
	"github.com/platinasystems/go/redis/rpc/args"
	"github.com/platinasystems/go/sockfile"
)

type Reg struct {
	Srvr     *sockfile.RpcServer
	assign   Assigner
	unassign Unassigner
}

type Assigner func(string, interface{}) error
type Unassigner func(string) error

// e.g. name, "redis-reg"
func New(name string, assign Assigner, unassign Unassigner) (*Reg, error) {
	var reg *Reg
	srvr, err := sockfile.NewRpcServer("redis-reg")
	if err != nil {
		reg = &Reg{srvr, assign, unassign}
	}
	return reg, err
}

// Assign an RPC handler for the given redis key.
func (reg *Reg) Assign(a args.Assign, _ *struct{}) error {
	return reg.assign(a.Key, &rpc.Rpc{a.File, a.Name})
}

// Assign the handler for the given redis key.
func (reg *Reg) Unassign(a args.Unassign, _ *struct{}) error {
	return reg.unassign(a.Key)
}
