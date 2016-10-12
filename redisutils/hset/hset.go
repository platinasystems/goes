// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package hset

import (
	"fmt"

	"github.com/platinasystems/go/flags"
	"github.com/platinasystems/go/redis"
)

type hset struct{}

func New() hset { return hset{} }

func (hset) String() string { return "hset" }
func (hset) Usage() string  { return "hset [-q] KEY FIELD VALUE" }

func (hset) Main(args ...string) error {
	flag, args := flags.New(args, "-q")
	switch len(args) {
	case 0:
		return fmt.Errorf("KEY FIELD VALUE: missing")
	case 1:
		return fmt.Errorf("FIELD VALUE: missing")
	case 2:
		return fmt.Errorf("VALUE: missing")
	case 3:
	default:
		return fmt.Errorf("%v: unexpected", args[3:])
	}
	i, err := redis.Hset(args[0], args[1], args[2])
	if err != nil {
		return err
	}
	if !flag["-q"] {
		fmt.Println(i)
	}
	return nil
}

func (hset) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "set the string value of a redis hash field",
	}
}
