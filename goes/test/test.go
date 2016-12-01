// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package test provides a collection of test commands and daemons.
package test

import (
	"github.com/platinasystems/go/goes/test/gohellod"
	"github.com/platinasystems/go/goes/test/gopanic"
	"github.com/platinasystems/go/goes/test/gopanicd"
	"github.com/platinasystems/go/goes/test/hellod"
	"github.com/platinasystems/go/goes/test/panic"
	"github.com/platinasystems/go/goes/test/panicd"
	"github.com/platinasystems/go/goes/test/sleeper"
	"github.com/platinasystems/go/goes/test/stringd"
)

func New() []interface{} {
	return []interface{}{
		gohellod.New(),
		gopanic.New(),
		gopanicd.New(),
		hellod.New(),
		panic.New(),
		panicd.New(),
		sleeper.New(),
		stringd.New(),
	}
}
