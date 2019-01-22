// Copyright © 2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package buildid

import (
	"fmt"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/buildid"
)

type Command struct{}

func (Command) String() string { return "buildid" }

func (Command) Usage() string {
	return "buildid [PROGRAM]..."
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "print buildid of GO program(s)",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Print the buildid of the given GO programs or /proc/self/exe if none.

SEE ALSO 
	go tool buildid`,
	}
}

func (Command) Main(args ...string) error {
	if len(args) == 0 {
		args = []string{"/proc/self/exe"}
	}
	for _, fn := range args {
		s, err := buildid.New(fn)
		if err != nil {
			return err
		}
		if len(args) > 1 {
			fmt.Printf("%s: %s\n", fn, s)
		} else {
			fmt.Println(s)
		}
	}
	return nil
}
