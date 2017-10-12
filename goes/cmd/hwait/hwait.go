// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hwait

import (
	"fmt"
	"time"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/redis"
)

const (
	Name    = "hwait"
	Apropos = "wait until the redis hash field has given value"
	Usage   = "hwait KEY FIELD VALUE [TIMEOUT(seconds)]"
)

type Interface interface {
	Apropos() lang.Alt
	Complete(...string) []string
	Main(...string) error
	String() string
	Usage() string
}

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Complete(args ...string) []string {
	return redis.Complete(args...)
}

func (Command) Main(args ...string) error {
	n := time.Duration(3)
	switch len(args) {
	case 0:
		return fmt.Errorf("KEY FIELD: missing")
	case 1:
		return fmt.Errorf("FIELD: missing")
	case 2:
		return fmt.Errorf("VALUE: missing")
	case 3:
	case 4:
		if _, err := fmt.Sscan(args[3], &n); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%v: unexpected", args[4:])
	}
	return redis.Hwait(args[0], args[1], args[2], n*time.Second)
}
