// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package netutils provides network admin commands.
package netutils

import (
	"github.com/platinasystems/go/netutils/femtocom"
	"github.com/platinasystems/go/netutils/nsid"
	"github.com/platinasystems/go/netutils/ping"
)

func New() []interface{} {
	return []interface{}{
		femtocom.New(),
		nsid.New(),
		ping.New(),
	}
}
