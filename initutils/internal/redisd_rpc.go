// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package internal

import (
	"github.com/platinasystems/go/redis"
	"github.com/platinasystems/go/sockfile"
)

type RedisReg struct{ Srvr *sockfile.RpcServer }
type RedisRpc struct{ File, Name string }

var empty = struct{}{}

func (p *RedisReg) main() (err error) {
	p.Srvr, err = sockfile.NewRpcServer("redis-reg")
	return
}

// Assign an RPC handler for the given redis key.
func (*RedisReg) Assign(args redis.AssignArgs, _ *struct{}) error {
	Init.assign(args.Key, &RedisRpc{args.File, args.Name})
	return nil
}

// Assign the handler for the given redis key.
func (*RedisReg) Unassign(args redis.UnassignArgs, _ *struct{}) error {
	Init.unassign(args.Key)
	return nil
}

func (r *RedisRpc) Del(key string, keys ...string) (int, error) {
	cl, err := sockfile.NewRpcClient(r.File)
	if err != nil {
		return 0, err
	}
	defer cl.Close()
	var reply redis.DelReply
	err = cl.Call(r.Name+".Del", redis.DelArgs{key, keys}, &reply)
	if err != nil {
		return 0, err
	}
	return reply.Redis(), nil
}

func (r *RedisRpc) Get(key string) ([]byte, error) {
	cl, err := sockfile.NewRpcClient(r.File)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	var reply redis.GetReply
	err = cl.Call(r.Name+".Get", redis.GetArgs{key}, &reply)
	if err != nil {
		return nil, err
	}
	return reply.Redis(), nil
}

func (r *RedisRpc) Set(key string, value []byte) error {
	cl, err := sockfile.NewRpcClient(r.File)
	if err != nil {
		return err
	}
	defer cl.Close()
	return cl.Call(r.Name+".Set", redis.SetArgs{key, value}, &empty)
}

func (r *RedisRpc) Hdel(key, field string, fields ...string) (int, error) {
	cl, err := sockfile.NewRpcClient(r.File)
	if err != nil {
		return 0, err
	}
	defer cl.Close()
	var reply redis.HdelReply
	err = cl.Call(r.Name+".Hdel", redis.HdelArgs{key, field, fields},
		&reply)
	if err != nil {
		return 0, err
	}
	return reply.Redis(), nil
}

func (r *RedisRpc) Hexists(key, field string) (int, error) {
	cl, err := sockfile.NewRpcClient(r.File)
	if err != nil {
		return 0, err
	}
	defer cl.Close()
	var reply redis.HexistsReply
	err = cl.Call(r.Name+".Hexists", redis.HexistsArgs{key, field}, &reply)
	if err != nil {
		return 0, err
	}
	return reply.Redis(), nil
}

func (r *RedisRpc) Hget(key, field string) ([]byte, error) {
	cl, err := sockfile.NewRpcClient(r.File)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	var reply redis.HgetReply
	err = cl.Call(r.Name+".Hget", redis.HgetArgs{key, field}, &reply)
	if err != nil {
		return nil, err
	}
	return reply.Redis(), nil
}

func (r *RedisRpc) Hgetall(key string) ([][]byte, error) {
	cl, err := sockfile.NewRpcClient(r.File)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	var reply redis.HgetallReply
	err = cl.Call(r.Name+".Hgetall", redis.HgetallArgs{key}, &reply)
	if err != nil {
		return nil, err
	}
	return reply.Redis(), nil
}

func (r *RedisRpc) Hkeys(key string) ([][]byte, error) {
	cl, err := sockfile.NewRpcClient(r.File)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	var reply redis.HkeysReply
	err = cl.Call(r.Name+".Hkeys", redis.HkeysArgs{key}, &reply)
	if err != nil {
		return nil, err
	}
	return reply.Redis(), nil
}

func (r *RedisRpc) Hset(key, id string, value []byte) (int, error) {
	cl, err := sockfile.NewRpcClient(r.File)
	if err != nil {
		return 0, err
	}
	defer cl.Close()
	var reply redis.HsetReply
	err = cl.Call(r.Name+".Hset", redis.HsetArgs{key, id, value},
		&reply)
	if err != nil {
		return 0, err
	}
	return reply.Redis(), nil
}

func (r *RedisRpc) Lrange(key string, start, stop int) ([][]byte, error) {
	cl, err := sockfile.NewRpcClient(r.File)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	var reply redis.LrangeReply
	err = cl.Call(r.Name+".Lrange", redis.LrangeArgs{key, start, stop},
		&reply)
	if err != nil {
		return nil, err
	}
	return reply.Redis(), nil
}

func (r *RedisRpc) Lindex(key string, index int) ([]byte, error) {
	cl, err := sockfile.NewRpcClient(r.File)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	var reply redis.LindexReply
	err = cl.Call(r.Name+".Lindex", redis.LindexArgs{key, index}, &reply)
	if err != nil {
		return nil, err
	}
	return reply.Redis(), nil
}

func (r *RedisRpc) Blpop(key string, keys ...string) ([][]byte, error) {
	cl, err := sockfile.NewRpcClient(r.File)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	var reply redis.BlpopReply
	err = cl.Call(r.Name+".Blpop", redis.BlpopArgs{key, keys}, &reply)
	if err != nil {
		return nil, err
	}
	return reply.Redis(), nil
}

func (r *RedisRpc) Brpop(key string, keys ...string) ([][]byte, error) {
	cl, err := sockfile.NewRpcClient(r.File)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	var reply redis.BrpopReply
	err = cl.Call(r.Name+".Brpop", redis.BrpopArgs{key, keys}, &reply)
	if err != nil {
		return nil, err
	}
	return reply.Redis(), nil
}

func (r *RedisRpc) Lpush(key string, value []byte, values ...[]byte) (int, error) {
	cl, err := sockfile.NewRpcClient(r.File)
	if err != nil {
		return 0, err
	}
	defer cl.Close()
	var reply redis.LpushReply
	err = cl.Call(r.Name+".Lpush", redis.LpushArgs{key, value, values},
		&reply)
	if err != nil {
		return 0, err
	}
	return reply.Redis(), nil
}

func (r *RedisRpc) Rpush(key string, value []byte, values ...[]byte) (int, error) {
	cl, err := sockfile.NewRpcClient(r.File)
	if err != nil {
		return 0, err
	}
	defer cl.Close()
	var reply redis.RpushReply
	err = cl.Call(r.Name+".Rpush", redis.RpushArgs{key, value, values},
		&reply)
	if err != nil {
		return 0, err
	}
	return reply.Redis(), nil
}
