// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package chmod

import (
	"fmt"
	"os"
	"strconv"

	"github.com/platinasystems/go/goes/lang"
)

type Command struct{}

func (Command) String() string { return "chmod" }

func (Command) Usage() string { return "chmod MODE FILE..." }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "change file mode",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Changed each FILE's mode bits to the given octal MODE.`,
	}
}

func (Command) Main(args ...string) error {
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
