// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package chmod

import (
	"fmt"
	"os"
	"strconv"

	"github.com/platinasystems/go/internal/goes/lang"
)

const (
	Name    = "chmod"
	Apropos = "change file mode"
	Usage   = "chmod MODE FILE..."
	Man     = `
DESCRIPTION
	Changed each FILE's mode bits to the given octal MODE.`
)

type Interface interface {
	Apropos() lang.Alt
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }

func (cmd) Main(args ...string) error {
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

func (cmd) Man() lang.Alt  { return man }
func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)
