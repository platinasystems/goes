// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main_test

import (
	"strings"
	"testing"
	"time"

	"github.com/platinasystems/go/internal/test"
	main "github.com/platinasystems/go/main/goes-example"
)

func Test(t *testing.T) {
	if test.Goes {
		test.Exec(main.Goes().Main)
	}
	test.Suite{
		{"hello", hello},
		{"pwd", pwd},
		{"cat", cat},
		{"redis", redis},
	}.Run(t)
}

func hello(t *testing.T) {
	test.Assert{t}.Program(nil,
		"goes", "echo", "hello", "world",
	).Output("hello world\n").Done()

}

func pwd(t *testing.T) {
	test.Assert{t}.Program(nil,
		"goes", "pwd",
	).Output("/.*/platinasystems/go\n/").Done()

}

func cat(t *testing.T) {
	test.Assert{t}.Program(strings.NewReader("HELLO WORLD"),
		"goes", "cat", "-",
	).Output("HELLO WORLD").Done()

}

func redis(t *testing.T) {
	assert := test.Assert{t}
	assert.YoureRoot()
	defer assert.Program(nil, "goes", "redisd").Quit(10 * time.Second)
	assert.Program(nil,
		"goes", "hwait", "platina", "redis.ready", "true", "10",
	).Ok().Done()
}
