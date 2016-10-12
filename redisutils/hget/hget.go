// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package hget

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/redis"
	"github.com/platinasystems/go/redisutils/internal"
)

type hget struct{}

func New() hget { return hget{} }

func (hget) String() string { return "hget" }
func (hget) Usage() string  { return "hget KEY FIELD" }

func (hget) Main(args ...string) error {
	switch len(args) {
	case 0:
		return fmt.Errorf("KEY FIELD: missing")
	case 1:
		return fmt.Errorf("FIELD: missing")
	case 2:
	default:
		return fmt.Errorf("%v: unexpected", args[2:])
	}
	s, err := redis.Hget(args[0], args[1])
	if err != nil {
		return err
	}
	internal.Fprintln(os.Stdout, s)
	return nil
}

func (hget) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "get the value of a redis hash field",
	}
}
