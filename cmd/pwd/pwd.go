// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pwd

import (
	"fmt"
	"os"

	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "pwd" }

func (Command) Usage() string {
	return "pwd [-L]"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "print working directory",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Print the full filename of the process working directory.

	-L  use PWD from environment, even if it contains symlinks;
	    default avoids symlinks

NOTE 
	This may be different than the context directory.`,
	}
}

func (Command) Main(args ...string) error {
	flag, args := flags.New(args, "-L")
	if len(args) != 0 {
		return fmt.Errorf("%v: unexpected", args)
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if flag.ByName["-L"] {
		return fmt.Errorf("FIXME")
	} else {
		fmt.Println(wd)
	}
	return nil
}
