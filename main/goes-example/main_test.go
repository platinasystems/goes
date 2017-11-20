// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/platinasystems/go/internal/test"
)

func Test(t *testing.T) {
	test.Main(main)
	test.Suite{
		{"hello", func(t *testing.T) {
			test.Assert{t}.Program(
				regexp.MustCompile("hello world\n"),
				test.Self{}, "echo", "hello", "world")
		}},
		{"pwd", func(t *testing.T) {
			test.Assert{t}.Program(test.Self{}, "pwd")
		}},
		{"cat", func(t *testing.T) {
			test.Assert{t}.Program(
				strings.NewReader("HELLO WORLD"),
				regexp.MustCompile("HELLO WORLD"),
				test.Self{}, "cat", "-")
		}},
		{"redis", func(t *testing.T) {
			assert := test.Assert{t}
			assert.YoureRoot()
			defer assert.Background(test.Self{}, "redisd").Quit()
			assert.Program(12*time.Second, test.Self{},
				"hwait", "platina", "redis.ready", "true", "10")
		}},
	}.Run(t)
}
