// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/go/builtinutils"
	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/kutils"
)

func main() {
	command.Plot(builtinutils.New()...)
	command.Plot(kutils.New()...)
	command.Sort()
	goes.Main()
}
