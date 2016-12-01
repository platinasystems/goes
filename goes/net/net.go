// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package net provides network admin commands.
package net

import (
	"github.com/platinasystems/go/goes/net/femtocom"
	"github.com/platinasystems/go/goes/net/nld"
	"github.com/platinasystems/go/goes/net/nsid"
	"github.com/platinasystems/go/goes/net/ping"
	"github.com/platinasystems/go/goes/net/wget"
)

func New() []interface{} {
	return []interface{}{
		femtocom.New(),
		nld.New(),
		nsid.New(),
		ping.New(),
		wget.New(),
	}
}
