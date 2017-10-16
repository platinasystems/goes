// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main_test

import (
	"testing"
	"time"

	. "github.com/platinasystems/go/internal/test"
	main "github.com/platinasystems/go/main/goes-platina-mk1"
)

func Test(t *testing.T) {
	if Goes {
		Exec(main.Goes().Main)
	}
	t.Run("VnetReady", VnetReady)
}

func VnetReady(t *testing.T) {
	assert := Assert{t}
	assert.YoureRoot()
	defer assert.Program(nil,
		"goes", "redisd",
	).Quit(10 * time.Second)
	assert.Program(nil,
		"goes", "hwait", "platina", "redis.ready", "true", "10",
	).Ok()
	defer assert.Program(nil,
		"goes", "vnetd",
	).Gdb().Quit(30 * time.Second)
	assert.Program(nil,
		"goes", "hwait", "platina", "vnet.ready", "true", "30",
	).Ok()
}
