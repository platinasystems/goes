// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package diagutils provides a collection of diagnostic commands.
package diagutils

import (
	"github.com/platinasystems/go/diagutils/diag"
	"github.com/platinasystems/go/diagutils/gpio"
	"github.com/platinasystems/go/diagutils/i2c"
)

func New() []interface{} {
	return []interface{}{
		diag.New(),
		gpio.New(),
		i2c.New(),
	}
}
