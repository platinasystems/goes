// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package sync

import (
	"syscall"
)

type sync struct{}

func New() sync { return sync{} }

func (sync) String() string { return "sync" }
func (sync) Usage() string  { return "sync" }

func (sync) Main(args ...string) error {
	syscall.Sync()
	return nil
}

func (sync) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "flush file system buffers",
	}
}

func (sync) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	sync - flush file system buffers

SYNOPSIS
	sync

DESCRIPTION
	Force changed blocks to disk, update the super block.`,
	}
}
