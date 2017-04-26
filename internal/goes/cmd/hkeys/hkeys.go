// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hkeys

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/internal/goes/lang"
	"github.com/platinasystems/go/internal/redis"
)

const (
	Name    = "hkeys"
	Apropos = "get all the fields in a redis hash"
	Usage   = "hkeys KEY"
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
		return fmt.Errorf("KEY: missing")
	case 1:
	default:
		return fmt.Errorf("%v: unexpected", args[1:])
	}
	keys, err := redis.Hkeys(args[0])
	if err != nil {
		return err
	}
	for _, s := range keys {
		redis.Fprintln(os.Stdout, s)
	}
	return nil
}

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
