// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package usage

import (
	"fmt"
	"strings"

	"github.com/platinasystems/go/goes"
)

type usage struct{}

func New() usage { return usage{} }

func (usage) String() string { return "usage" }
func (usage) Tag() string    { return "builtin" }
func (usage) Usage() string  { return "usage  COMMAND..." }

func (u usage) Main(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("COMMAND: missing")
	}
	for _, arg := range args {
		s, found := goes.Usage[arg]
		if !found {
			return fmt.Errorf("%s: not found", arg)
		}
		if strings.IndexRune(s, '\n') >= 0 {
			fmt.Print("usage:\t", s, "\n")
		} else {
			fmt.Println("usage:", s)
		}
	}
	return nil
}
