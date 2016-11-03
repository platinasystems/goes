// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package fs provides file system admin commands.
package fs

import (
	"github.com/platinasystems/go/commands/fs/mount"
	"github.com/platinasystems/go/commands/fs/umount"
)

func New() []interface{} {
	return []interface{}{
		mount.New(),
		umount.New(),
	}
}
