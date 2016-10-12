// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package hgetall

import (
	"fmt"
	"strings"

	"github.com/platinasystems/go/redis"
	"github.com/platinasystems/go/redisutils/internal"
)

type hgetall struct{}

func New() hgetall { return hgetall{} }

func (hgetall) String() string { return "hgetall" }
func (hgetall) Usage() string  { return "hgetall KEY [PATTERN]" }

func (hgetall) Main(args ...string) error {
	switch len(args) {
	case 0:
		return fmt.Errorf("KEY: missing")
	case 1:
	case 2:
	default:
		return fmt.Errorf("%v: unexpected", args[2:])
	}
	r, err := redis.Connect()
	if err != nil {
		return err
	}
	defer r.Close()
	ret, err := r.Do("HGETALL", args[0])
	if err != nil {
		return err
	}
	list := ret.([]interface{})
	if len(args) == 1 {
		for i := 0; i < len(list); i += 2 {
			fmt.Print(internal.Quotes(string(list[i].([]byte))))
			if list[i+1] != nil {
				fmt.Print(": ")
				fmt.Print(internal.Quotes(string(list[i+1].([]byte))))
			}
			fmt.Println()
		}
	} else {
		for i := 0; i < len(list); i += 2 {
			if strings.Contains(string(list[i].([]byte)), args[1]) {
				fmt.Print(internal.Quotes(string(list[i].([]byte))))
				if list[i+1] != nil {
					fmt.Print(": ")
					fmt.Print(internal.Quotes(string(list[i+1].([]byte))))
				}
				fmt.Println()
			}
		}
	}
	return nil
}

func (hgetall) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "get all the field values in a redis hash",
	}
}
