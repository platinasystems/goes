// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hexists

import (
	"fmt"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/redis"
)

const (
	Name    = "hexists"
	Apropos = "determine if the redis hash field exists"
	Usage   = "hexists KEY FIELD"
)

type Interface interface {
	Apropos() lang.Alt
	Complete(...string) []string
	Main(...string) error
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }

func (cmd) Complete(args ...string) []string {
	return redis.Complete(args...)
}

func (cmd) Main(args ...string) error {
	switch len(args) {
	case 0:
		return fmt.Errorf("KEY FIELD: missing")
	case 1:
		return fmt.Errorf("FIELD: missing")
	case 2:
	default:
		return fmt.Errorf("%v: unexpected", args[2:])
	}
	r, err := redis.Connect()
	if err != nil {
		return err
	}
	defer r.Close()
	ret, err := r.Do("HEXISTS", args[0], args[1])
	if err != nil {
		return err
	}
	fmt.Println(ret)
	return nil
}

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
