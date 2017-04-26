// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package sleep

import (
	"fmt"
	"strconv"
	"time"

	"github.com/platinasystems/go/internal/goes/lang"
)

const (
	Name    = "sleep"
	Apropos = "suspend execution for an interval of time"
	Usage   = "sleep SECONDS"
	Man     = `
DESCRIPTION
	The sleep command suspends execution for a number of SECONDS.`
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
		return fmt.Errorf("SECONDS: missing")
	}

	t, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return fmt.Errorf("%s: %v", args[0], err)
	}

	time.Sleep(time.Second * time.Duration(t))
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
