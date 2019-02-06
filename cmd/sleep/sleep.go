// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package sleep

import (
	"fmt"
	"strconv"
	"time"

	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "sleep" }

func (Command) Usage() string { return "sleep SECONDS" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "suspend execution for an interval of time",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	The sleep command suspends execution for a number of SECONDS.`,
	}
}

func (Command) Main(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("SECONDS: missing")
	}

	t, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return fmt.Errorf("%s: %v", args[0], err)
	}

	time.Sleep(time.Second * time.Duration(t))
	return nil
}
