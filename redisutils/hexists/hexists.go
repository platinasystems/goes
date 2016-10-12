// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package hexists

import (
	"fmt"

	"github.com/platinasystems/go/redis"
)

type hexists struct{}

func New() hexists { return hexists{} }

func (hexists) String() string { return "hexists" }
func (hexists) Usage() string  { return "hexists KEY FIELD" }

func (hexists) Main(args ...string) error {
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

func (hexists) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "determine if the redis hash field exists",
	}
}
