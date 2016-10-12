// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package initutils provides /init, /sbin/init and /usr/sbin/goesd commands.
package initutils

import (
	"github.com/platinasystems/go/initutils/goesd"
	"github.com/platinasystems/go/initutils/sbininit"
	"github.com/platinasystems/go/initutils/slashinit"
)

func New() []interface{} {
	return []interface{}{
		goesd.New(),
		sbininit.New(),
		slashinit.New(),
	}
}
