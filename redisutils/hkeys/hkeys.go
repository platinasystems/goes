// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package hkeys

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/redis"
	"github.com/platinasystems/go/redisutils/internal"
)

type hkeys struct{}

func New() hkeys { return hkeys{} }

func (hkeys) String() string { return "hkeys" }
func (hkeys) Usage() string  { return "hkeys KEY" }

func (hkeys) Main(args ...string) error {
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
		internal.Fprintln(os.Stdout, s)
	}
	return nil
}

func (hkeys) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "get all the fields in a redis hash",
	}
}
