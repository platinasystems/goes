// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build freebsd

package jail

import (
	"context"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

const CanJail = true

// syscall.ForkExec self with ProcAttr.Sys.Jail
func Jail(ctx context.Context, name string) error {
	return xerrors.FIXME("jail", name)
}
