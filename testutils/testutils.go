// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package testutils provides a collection of test commands and daemons.
package testutils

import (
	"github.com/platinasystems/go/testutils/gohellod"
	"github.com/platinasystems/go/testutils/gopanic"
	"github.com/platinasystems/go/testutils/gopanicd"
	"github.com/platinasystems/go/testutils/hellod"
	"github.com/platinasystems/go/testutils/panic"
	"github.com/platinasystems/go/testutils/panicd"
)

func New() []interface{} {
	return []interface{}{
		gohellod.New(),
		gopanic.New(),
		gopanicd.New(),
		hellod.New(),
		panic.New(),
		panicd.New(),
	}
}
