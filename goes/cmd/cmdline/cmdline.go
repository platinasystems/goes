// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cmdline

import (
	"fmt"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/cmdline"
)

type Command struct{}

func (Command) String() string { return "cmdline" }

func (Command) Usage() string { return "cmdline [NAME]..." }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "parse and print /proc/cmdline variables",
	}
}

func (Command) Main(args ...string) error {
	keys, m, err := cmdline.New()
	if err != nil {
		return err
	}
	if len(args) > 0 {
		keys = args
	}
	for _, k := range keys {
		if _, found := m[k]; !found {
			return fmt.Errorf("%s: not found", k)
		}
	}
	if len(keys) == 1 {
		fmt.Println(m[keys[0]])
	} else {
		for _, k := range keys {
			fmt.Print(k, ": ", m[k], "\n")
		}
	}
	return nil
}
