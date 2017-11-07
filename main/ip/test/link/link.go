// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package link

import (
	"fmt"
	"os"
	"testing"

	"github.com/platinasystems/go/internal/test"
)

// Suite of 'ip link' tests
var Suite = test.Suite{
	{"default", func(t *testing.T) {
		test.Assert{t}.Program(nil, "goes", "ip", "link").Ok()
	}},
	{"show", show},
	{"add", add},
}.Run

var show = test.Suite{
	{"default", func(t *testing.T) {
		test.Assert{t}.Program(nil, "goes", "ip", "link", "show").Ok()
	}},
	{"lo", func(t *testing.T) {
		test.Assert{t}.Program(nil, "goes", "ip", "link", "show",
			"lo").Ok()
	}},
}.Run

func add(t *testing.T) {
	test.Assert{t}.YoureRoot()
	test.Suite{
		{"dummy", dummy},
		{"ipip", test.Suite{
			{"fou", fou},
		}.Run},
	}.Run(t)
}

func dummy(t *testing.T) {
	assert := test.Assert{t}
	name := fmt.Sprint("dummy", os.Getpid())
	assert.Program(nil, "goes", "ip", "link", "add", "type", "dummy",
		"name", name).Ok()
	assert.Program(nil, "goes", "ip", "link", "show", name).Ok()
	assert.Program(nil, "goes", "ip", "link", "delete", name).Ok()
}

func fou(t *testing.T) {
	assert := test.Assert{t}
	name := fmt.Sprint("fou", os.Getpid())
	assert.Program(nil, "goes",
		"ip", "link", "add",
		"type", "ipip",
		"name", name,
		"dev", "eth0",
		"encap", "fou",
		"encap-sport", "any",
		"encap-dport", "7777").Ok()
	assert.Program(nil, "goes", "ip", "link", "show", name).Ok()
	assert.Program(nil, "goes", "ip", "link", "delete", name).Ok()
}
