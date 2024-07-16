// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build freebsd || netbsd || openbsd

package nettun

import (
	"os"

	"github.com/platinasystems/goes/v2/pkg/netif"
)

const (
	HasPI          = false
	CanTAP         = false
	CanPersist     = false
	CanChangeOwner = false
	CanChangeGroup = false
)

func New(
	unit uint,
	isTAP bool,
	persist bool,
	owner, group int,
	ha netif.HardwareAddr,
) (*os.File, error) {
	return nil, ErrUnsupported
}
