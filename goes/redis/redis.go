// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package redis provides redis client commands for the local server.
package redis

import (
	"github.com/platinasystems/go/goes/redis/get"
	"github.com/platinasystems/go/goes/redis/hdel"
	"github.com/platinasystems/go/goes/redis/hexists"
	"github.com/platinasystems/go/goes/redis/hget"
	"github.com/platinasystems/go/goes/redis/hgetall"
	"github.com/platinasystems/go/goes/redis/hkeys"
	"github.com/platinasystems/go/goes/redis/hset"
	"github.com/platinasystems/go/goes/redis/keys"
	"github.com/platinasystems/go/goes/redis/lrange"
	"github.com/platinasystems/go/goes/redis/redisd"
	"github.com/platinasystems/go/goes/redis/set"
	"github.com/platinasystems/go/goes/redis/subscribe"
)

func New() []interface{} {
	return []interface{}{
		get.New(),
		hdel.New(),
		hexists.New(),
		hget.New(),
		hgetall.New(),
		hkeys.New(),
		hset.New(),
		keys.New(),
		lrange.New(),
		redisd.New(),
		set.New(),
		subscribe.New(),
	}
}
