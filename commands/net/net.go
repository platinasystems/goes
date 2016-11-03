// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package net provides network admin commands.
package net

import (
	"github.com/platinasystems/go/commands/net/femtocom"
	"github.com/platinasystems/go/commands/net/nsid"
	"github.com/platinasystems/go/commands/net/ping"
	"github.com/platinasystems/go/commands/net/wget"
)

func New() []interface{} {
	return []interface{}{
		femtocom.New(),
		nsid.New(),
		ping.New(),
		wget.New(),
	}
}
