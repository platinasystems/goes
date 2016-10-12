// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package hdel

import (
	"fmt"

	"github.com/platinasystems/go/redis"
)

type hdel struct{}

func New() hdel { return hdel{} }

func (hdel) String() string { return "hdel" }
func (hdel) Usage() string  { return "hdel KEY FIELD" }

func (hdel) Main(args ...string) error {
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

func (hdel) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "delete one or more redis hash fields",
	}
}
