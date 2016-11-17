// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/commands/builtin"
	"github.com/platinasystems/go/commands/kernel"
	"github.com/platinasystems/go/goes"
)

func main() {
	command.Plot(builtin.New()...)
	command.Plot(kernel.New()...)
	command.Sort()
	goes.Main()
}
