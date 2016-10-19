// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package redisutils provides redis client commands for the local server.
package redisutils

import (
	"github.com/platinasystems/go/redisutils/get"
	"github.com/platinasystems/go/redisutils/hdel"
	"github.com/platinasystems/go/redisutils/hexists"
	"github.com/platinasystems/go/redisutils/hget"
	"github.com/platinasystems/go/redisutils/hgetall"
	"github.com/platinasystems/go/redisutils/hkeys"
	"github.com/platinasystems/go/redisutils/hset"
	"github.com/platinasystems/go/redisutils/keys"
	"github.com/platinasystems/go/redisutils/lrange"
	"github.com/platinasystems/go/redisutils/redisd"
	"github.com/platinasystems/go/redisutils/set"
	"github.com/platinasystems/go/redisutils/subscribe"
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
