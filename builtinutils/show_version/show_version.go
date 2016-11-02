// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package show_version

import (
	"fmt"

	. "github.com/platinasystems/go/version"
)

const Name = "show-version"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Tag() string    { return "builtin" }
func (cmd) Usage() string  { return Name }

func (cmd) Main(args ...string) error {
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}
	fmt.Println(Version)
	return nil
}
