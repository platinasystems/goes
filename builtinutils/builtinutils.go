// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package builtinutils provides commands that run in the same context as the
// goes cli or script instead of as a sub-process.
package builtinutils

import (
	"github.com/platinasystems/go/builtinutils/apropos"
	"github.com/platinasystems/go/builtinutils/cd"
	"github.com/platinasystems/go/builtinutils/cli"
	"github.com/platinasystems/go/builtinutils/cli_escapes"
	"github.com/platinasystems/go/builtinutils/cli_options"
	"github.com/platinasystems/go/builtinutils/complete"
	"github.com/platinasystems/go/builtinutils/env"
	"github.com/platinasystems/go/builtinutils/exit"
	"github.com/platinasystems/go/builtinutils/export"
	"github.com/platinasystems/go/builtinutils/man"
	"github.com/platinasystems/go/builtinutils/resize"
	"github.com/platinasystems/go/builtinutils/show_commands"
	"github.com/platinasystems/go/builtinutils/source"
	"github.com/platinasystems/go/builtinutils/usage"
)

func New() []interface{} {
	return []interface{}{
		apropos.New(),
		cli.New(),
		cli_escapes.New(),
		cli_options.New(),
		cd.New(),
		complete.New(),
		env.New(),
		exit.New(),
		export.New(),
		man.New(),
		resize.New(),
		show_commands.New(),
		source.New(),
		usage.New(),
	}
}
