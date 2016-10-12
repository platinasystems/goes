// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package get

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/redis"
	"github.com/platinasystems/go/redisutils/internal"
)

type get struct{}

func New() get { return get{} }

func (get) String() string { return "get" }
func (get) Usage() string  { return "get KEY" }

func (get) Main(args ...string) error {
	switch len(args) {
	case 0:
		return fmt.Errorf("KEY: missing")
	case 1:
	default:
		return fmt.Errorf("%v: unexpected", args[1:])
	}
	s, err := redis.Get(args[0])
	if err != nil {
		return err
	}
	internal.Fprintln(os.Stdout, s)
	return nil
}

func (get) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "get the value of a redis key",
	}
}
