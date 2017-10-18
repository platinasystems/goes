// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main_test

import (
	"strings"
	"testing"
	"time"

	. "github.com/platinasystems/go/internal/test"
	main "github.com/platinasystems/go/main/goes-example"
)

func Test(t *testing.T) {
	if Goes {
		Exec(main.Goes().Main)
	}
	t.Run("HelloWorld", HelloWorld)
	t.Run("Pwd", Pwd)
	t.Run("Cat", Cat)
	t.Run("RedisReady", RedisReady)
}

func HelloWorld(t *testing.T) {
	Assert{t}.Program(nil,
		"goes", "echo", "hello", "world",
	).Output(Equal("hello world\n"))

}

func Pwd(t *testing.T) {
	Assert{t}.Program(nil,
		"goes", "pwd",
	).Output(Match(".*/platinasystems/go\n"))

}

func Cat(t *testing.T) {
	Assert{t}.Program(strings.NewReader("HELLO WORLD"),
		"goes", "cat", "-",
	).Output(Equal("HELLO WORLD"))

}

func RedisReady(t *testing.T) {
	assert := Assert{t}
	assert.YoureRoot()
	defer assert.Program(nil, "goes", "redisd").Quit(10 * time.Second)
	assert.Program(nil,
		"goes", "hwait", "platina", "redis.ready", "true", "10",
	).Ok()
}
