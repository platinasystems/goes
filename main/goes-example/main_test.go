// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main_test

import (
	"testing"
	"time"

	"github.com/platinasystems/go/internal/test"
	. "github.com/platinasystems/go/main/goes-example"
)

func Test(t *testing.T) {
	if test.Goes {
		test.Exec(Goes().Main)
	}
	t.Run("HelloWorld", HelloWorld)
	t.Run("Pwd", Pwd)
	t.Run("RedisReady", RedisReady)
}

func HelloWorld(t *testing.T) {
	test.Assert{t}.OutputEqual("hello world\n",
		"goes", "echo", "hello", "world")
}

func Pwd(t *testing.T) {
	test.Assert{t}.OutputMatch(".*/platinasystems/go\n",
		"goes", "pwd")
}

func RedisReady(t *testing.T) {
	assert := test.Assert{t}
	assert.YoureRoot()
	defer assert.Background("goes", "redisd").Quit(10 * time.Second)
	assert.Ok("goes", "hwait", "platina", "redis.ready", "true", "10")
}
