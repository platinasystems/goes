// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mon

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
)

type Command struct{}

func (Command) String() string { return "mon" }

func (Command) Usage() string {
	return "ip netns monitor"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "network namespace",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
SEE ALSO
	ip man netns || ip netns -man
	man ip || ip -man`,
	}
}

func (Command) Main(args ...string) error {
	opt, args := options.New(args)
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	fmt.Println("FIXME", "mon")

	_ = opt

	return nil
}
