// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package hdel

import (
	"fmt"

	"github.com/platinasystems/go/goes/redis/internal"
	"github.com/platinasystems/go/redis"
)

const Name = "hdel"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " KEY FIELD" }

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
	ret, err := r.Do("HDEL", args[0], args[1])
	if err != nil {
		return err
	}
	fmt.Println(ret)
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "delete one or more redis hash fields",
	}
}

func (cmd) Complete(args ...string) []string {
	return internal.Complete(args...)
}
