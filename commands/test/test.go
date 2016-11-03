// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package test provides a collection of test commands and daemons.
package test

import (
	"github.com/platinasystems/go/commands/test/gohellod"
	"github.com/platinasystems/go/commands/test/gopanic"
	"github.com/platinasystems/go/commands/test/gopanicd"
	"github.com/platinasystems/go/commands/test/hellod"
	"github.com/platinasystems/go/commands/test/panic"
	"github.com/platinasystems/go/commands/test/panicd"
	"github.com/platinasystems/go/commands/test/sleeper"
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
	}
}
