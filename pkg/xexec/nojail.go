// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !freebsd

package xexec

import (
	"context"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

const CanJail = false

func Jail(ctx context.Context, name string) error {
	return xerrors.Broken("jail", name)
}
