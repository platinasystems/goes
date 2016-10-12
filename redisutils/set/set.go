// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package set

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/flags"
	"github.com/platinasystems/go/redis"
	"github.com/platinasystems/go/redisutils/internal"
)

type set struct{}

func New() set { return set{} }

func (set) String() string { return "set" }
func (set) Usage() string  { return "set [-q] KEY VALUE" }

func (set) Main(args ...string) error {
	flag, args := flags.New(args, "-q")
	switch len(args) {
	case 0:
		return fmt.Errorf("KEY VALUE: missing")
	case 1:
		return fmt.Errorf("VALUE: missing")
	case 2:
	default:
		return fmt.Errorf("%v: unexpected", args[2:])
	}
	s, err := redis.Set(args[0], args[1])
	if err != nil {
		return err
	}
	if !flag["-q"] {
		internal.Fprintln(os.Stdout, s)
	}
	return nil
}

func (set) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "set the string value of a redis key",
	}
}
