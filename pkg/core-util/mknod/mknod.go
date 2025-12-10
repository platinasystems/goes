// Copyright © 2015-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// go:build unix

package mknod

import (
	"context"
	"flag"
	"strconv"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"golang.org/x/sys/unix"
)

const MknodUsage = `
usage: {{.Name}} [flags] <name> <type> [<major> <minor>]
Make block, character, or pipe device files.",

{{flags .}}
Types

  b	Block
  c, u	Character (unbuffered)
  d	Directory
  p	FIFO
`

var (
	Mknod_m xflag.FileMode
)

var MknodFlags = xflag.Labels{
	{"m", "Octal file permission bits.", &Mknod_m},
}

func Mknod(ctx context.Context, args []string) error {
	var filetype uint32 = 0
	xflag.TemplateUsage(MknodUsage)
	err := MknodFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}

	args = flag.Args()
	nargs := len(args)
	if nargs == 0 {
		return xerrors.Incomplete("name")
	} else if nargs < 2 {
		return xerrors.Incomplete("type")
	}
	switch args[1] {
	case "b":
		filetype = unix.S_IFBLK
	case "c", "u":
		filetype = unix.S_IFCHR
	case "p":
		filetype = unix.S_IFIFO
	case "d":
		filetype = unix.S_IFDIR
	case "r":
		filetype = unix.S_IFREG
	}
	filetype |= uint32(Mknod_m.Mode())
	var maj, min uint64
	if nargs > 2 {
		if maj, err = strconv.ParseUint(args[2], 10, 8); err != nil {
			return xerrors.Label(err, "major")
		}
	}
	if nargs > 3 {
		if min, err = strconv.ParseUint(args[3], 10, 8); err != nil {
			return xerrors.Label(err, "minor")
		}
	}
	return unix.Mknod(args[0], filetype, int((maj*256)+min))
}
