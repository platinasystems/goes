// Copyright © 2015-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build unix

package sync

import (
	"context"

	"golang.org/x/sys/unix"
)

const SyncUsage = `
usage: {{.Name}}
Force changed blocks to disk, update the super block.
`

func Sync(ctx context.Context, args []string) error {
	unix.Sync()
	return nil
}
