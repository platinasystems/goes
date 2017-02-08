// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package boot

import (
	"fmt"
)

const Name = "boot"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " [OPTIONS]..." }

func (cmd) Main(args ...string) (err error) {
	if len(args) > 0 {
		err = fmt.Errorf("%v: unexpected", args)
	}

	return err
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "boot another operating system",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	boot - Boot another operating system

SYNOPSIS

The boot command finds other operating systems to load, and chooses an
appropriate one to execute.

DESCRIPTION

Boot is a high level interface to the kexec command. While kexec performs
the actual work, boot is a higher level interface that simplifies the process
of selecting a kernel to execute.

OPTIONS`,
	}
}


