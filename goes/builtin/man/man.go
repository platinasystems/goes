// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package man

import (
	"fmt"

	"github.com/platinasystems/go/goes"
)

type man struct{}

func New() man { return man{} }

func (man) String() string { return "man" }
func (man) Tag() string    { return "builtin" }
func (man) Usage() string  { return "man COMMAND..." }

func (m man) Main(args ...string) error {
	n := len(args)
	if n == 0 {
		return fmt.Errorf("COMMAND: missing")
	}
	for i, arg := range args {
		v, found := goes.Man[arg]
		if !found {
			fmt.Print(arg, ": has no man\n")
		} else {
			fmt.Println(v)
			if n > 1 && i < n-1 {
				fmt.Println()
			}
		}
	}
	return nil
}

func (man) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print command documentation",
	}
}
