// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"time"

	"github.com/platinasystems/go/goes/cmd/start"
	"github.com/platinasystems/go/internal/redis"
)

func init() {
	start.ConfHook = func() error {
		return redis.Hwait(redis.DefaultHash, "vnet.ready", "true",
			10*time.Second)
	}
}
