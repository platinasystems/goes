// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package keys

import (
	"fmt"
	"os"

	"github.com/platinasystems/go/redis"
	"github.com/platinasystems/go/redisutils/internal"
)

type keys struct{}

func New() keys { return keys{} }

func (keys) String() string { return "keys" }
func (keys) Usage() string  { return "keys [PATTERN]" }

func (keys) Main(args ...string) error {
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
		internal.Fprintln(os.Stdout, s)
	}
	return nil
}

func (keys) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "find all redis keys matching the given pattern",
	}
}
