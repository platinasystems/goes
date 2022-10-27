// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import (
	"github.com/platinasystems/goes/v2/pkg/net/ipc"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

var (
	Exchange = cache.New[ipc.Ipc](func(p *ipc.Ipc) error {
		*p = ipc.New()
		return nil
	}).Value
	Registry = cache.New[ipc.Ipc](func(p *ipc.Ipc) error {
		*p = ipc.New(".registry")
		return nil
	}).Value
)
