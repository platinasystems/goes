// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main_test

import (
	"regexp"
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
		{"helloworld", helloworld},
		{"pwd", pwd},
		{"cat", cat},
		{"redis", redis},
	}.Run(t)
}

func helloworld(t *testing.T) {
	test.Assert{t}.Program(
		regexp.MustCompile("hello world\n"),
		"goes", "echo", "hello", "world")
}

func pwd(t *testing.T) {
	test.Assert{t}.Program(
		regexp.MustCompile(".*/platinasystems/go\n"),
		"goes", "pwd")
}

func cat(t *testing.T) {
	test.Assert{t}.Program(
		strings.NewReader("HELLO WORLD"),
		regexp.MustCompile("HELLO WORLD"),
		"goes", "cat", "-")
}

func redis(t *testing.T) {
	assert := test.Assert{t}
	assert.YoureRoot()
	defer assert.Background("goes", "redisd").Quit()
	assert.Program(12*time.Second, "goes", "hwait", "platina",
		"redis.ready", "true", "10")
}
