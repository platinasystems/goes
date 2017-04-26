// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package keys

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/internal/goes/lang"
	"github.com/platinasystems/go/internal/redis"
)

const (
	Name    = "keys"
	Apropos = "find all redis keys matching the given pattern"
	Usage   = "keys [PATTERN]"
)

type Interface interface {
	Apropos() lang.Alt
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
	var pattern string
	switch len(args) {
	case 0:
		pattern = ".*"
	case 1:
		pattern = args[0]
	default:
		return fmt.Errorf("%v: unexpected", args[1:])
	}
	keys, err := redis.Keys(pattern)
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
