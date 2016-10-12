// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package chmod

import (
	"fmt"
	"os"
	"strconv"
)

type chmod struct{}

func New() chmod { return chmod{} }

func (chmod) String() string { return "chmod" }
func (chmod) Usage() string  { return "chmod MODE FILE..." }

func (chmod) Main(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("MODE: missing")
	}
	if len(args) < 2 {
		return fmt.Errorf("FILE: missing")
	}

	u64, err := strconv.ParseUint(args[0], 0, 32)
	if err != nil {
		return fmt.Errorf("%s: %v", args[0], err)
	}

	mode := os.FileMode(uint32(u64))

	for _, fn := range args[1:] {
		if err = os.Chmod(fn, mode); err != nil {
			return err
		}
	}
	return nil
}

func (chmod) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "change file mode",
	}
}

func (chmod) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	chmod - change file mode"

SYNOPSIS
	chmod MODE FILE...

DESCRIPTION
	Changed each FILE's mode bits to the given octal MODE.`,
	}
}
