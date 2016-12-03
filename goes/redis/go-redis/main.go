// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/builtin"
	"github.com/platinasystems/go/goes/core"
	"github.com/platinasystems/go/goes/redis"
)

func main() {
	goes.Plot(builtin.New()...)
	goes.Plot(core.New()...)
	goes.Plot(redis.New()...)
	goes.Sort()
	goes.Main()
}
