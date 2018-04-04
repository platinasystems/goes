// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package rpc provides remote calls to a redis server.
package rpc

import (
	"fmt"

	"github.com/platinasystems/go/internal/atsock"
	"github.com/platinasystems/go/internal/redis/rpc/args"
	"github.com/platinasystems/go/internal/redis/rpc/reply"
)

var empty = struct{}{}

type Rpc string

func New(name string) Rpc { return Rpc(name) }

func (rpc Rpc) String() string          { return string(rpc) }
func (rpc Rpc) Preface(s string) string { return fmt.Sprint(rpc, ".", s) }

func (rpc Rpc) Del(key string, keys ...string) (int, error) {
	cl, err := atsock.NewRpcClient(rpc.String())
	if err != nil {
		return 0, err
	}
	defer cl.Close()
	var r reply.Del
	err = cl.Call(rpc.Preface("Del"), args.Del{key, keys}, &r)
	if err != nil {
		return 0, err
	}
	return r.Redis(), nil
}

func (rpc Rpc) Get(key string) ([]byte, error) {
	cl, err := atsock.NewRpcClient(rpc.String())
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	var r reply.Get
	err = cl.Call(rpc.Preface("Get"), args.Get{key}, &r)
	if err != nil {
		return nil, err
	}
	return r.Redis(), nil
}

func (rpc Rpc) Set(key string, value []byte) error {
	cl, err := atsock.NewRpcClient(rpc.String())
	if err != nil {
		return err
	}
	defer cl.Close()
	return cl.Call(rpc.Preface("Set"), args.Set{key, value}, &empty)
}

func (rpc Rpc) Hdel(key, field string, fields ...string) (int, error) {
	cl, err := atsock.NewRpcClient(rpc.String())
	if err != nil {
		return 0, err
	}
	defer cl.Close()
	var r reply.Hdel
	err = cl.Call(rpc.Preface("Hdel"), args.Hdel{key, field, fields}, &r)
	if err != nil {
		return 0, err
	}
	return r.Redis(), nil
}

func (rpc Rpc) Hexists(key, field string) (int, error) {
	cl, err := atsock.NewRpcClient(rpc.String())
	if err != nil {
		return 0, err
	}
	defer cl.Close()
	var r reply.Hexists
	err = cl.Call(rpc.Preface("Hexists"), args.Hexists{key, field}, &r)
	if err != nil {
		return 0, err
	}
	return r.Redis(), nil
}

func (rpc Rpc) Hget(key, field string) ([]byte, error) {
	cl, err := atsock.NewRpcClient(rpc.String())
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	var r reply.Hget
	err = cl.Call(rpc.Preface("Hget"), args.Hget{key, field}, &r)
	if err != nil {
		return nil, err
	}
	return r.Redis(), nil
}

func (rpc Rpc) Hgetall(key string) ([][]byte, error) {
	cl, err := atsock.NewRpcClient(rpc.String())
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	var r reply.Hgetall
	err = cl.Call(rpc.Preface("Hgetall"), args.Hgetall{key}, &r)
	if err != nil {
		return nil, err
	}
	return r.Redis(), nil
}

func (rpc Rpc) Hkeys(key string) ([][]byte, error) {
	cl, err := atsock.NewRpcClient(rpc.String())
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	var r reply.Hkeys
	err = cl.Call(rpc.Preface("Hkeys"), args.Hkeys{key}, &r)
	if err != nil {
		return nil, err
	}
	return r.Redis(), nil
}

func (rpc Rpc) Hset(key, id string, value []byte) (int, error) {
	cl, err := atsock.NewRpcClient(rpc.String())
	if err != nil {
		return 0, err
	}
	defer cl.Close()
	var r reply.Hset
	err = cl.Call(rpc.Preface("Hset"), args.Hset{key, id, value}, &r)
	if err != nil {
		return 0, err
	}
	return r.Redis(), nil
}

func (rpc Rpc) Lrange(key string, start, stop int) ([][]byte, error) {
	cl, err := atsock.NewRpcClient(rpc.String())
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	var r reply.Lrange
	err = cl.Call(rpc.Preface("Lrange"), args.Lrange{key, start, stop}, &r)
	if err != nil {
		return nil, err
	}
	return r.Redis(), nil
}

func (rpc Rpc) Lindex(key string, index int) ([]byte, error) {
	cl, err := atsock.NewRpcClient(rpc.String())
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	var r reply.Lindex
	err = cl.Call(rpc.Preface("Lindex"), args.Lindex{key, index}, &r)
	if err != nil {
		return nil, err
	}
	return r.Redis(), nil
}

func (rpc Rpc) Blpop(key string, keys ...string) ([][]byte, error) {
	cl, err := atsock.NewRpcClient(rpc.String())
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	var r reply.Blpop
	err = cl.Call(rpc.Preface("Blpop"), args.Blpop{key, keys}, &r)
	if err != nil {
		return nil, err
	}
	return r.Redis(), nil
}

func (rpc Rpc) Brpop(key string, keys ...string) ([][]byte, error) {
	cl, err := atsock.NewRpcClient(rpc.String())
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	var r reply.Brpop
	err = cl.Call(rpc.Preface("Brpop"), args.Brpop{key, keys}, &r)
	if err != nil {
		return nil, err
	}
	return r.Redis(), nil
}

func (rpc Rpc) Lpush(key string, value []byte, values ...[]byte) (int, error) {
	cl, err := atsock.NewRpcClient(rpc.String())
	if err != nil {
		return 0, err
	}
	defer cl.Close()
	var r reply.Lpush
	err = cl.Call(rpc.Preface("Lpush"), args.Lpush{key, value, values}, &r)
	if err != nil {
		return 0, err
	}
	return r.Redis(), nil
}

func (rpc Rpc) Rpush(key string, value []byte, values ...[]byte) (int, error) {
	cl, err := atsock.NewRpcClient(rpc.String())
	if err != nil {
		return 0, err
	}
	defer cl.Close()
	var r reply.Rpush
	err = cl.Call(rpc.Preface("Rpush"), args.Rpush{key, value, values}, &r)
	if err != nil {
		return 0, err
	}
	return r.Redis(), nil
}
