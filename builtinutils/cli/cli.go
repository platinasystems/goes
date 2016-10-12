// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package cli

import (
	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/liner"
)

type cli struct{}

func New() cli { return cli{} }

func (cli) String() string { return "cli" }
func (cli) Tag() string    { return "builtin" }
func (cli) Usage() string  { return "man cli" }

func (cli) Main(args ...string) error {
	return command.Shell(liner.New().GetLine)
}

func (cli) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	cli - command line interpreter

DESCRIPTION
	The go-es command line interpreter is an incomplete shell with just
	this basic syntax:
		COMMAND [OPTIONS]... [ARGS]...

	The COMMAND and each option or argument are separated with one or more
	spaces. Leading and trailing spaces are ignored.
	
	Each command has an execution context that may be manipulated by
	options described later. Some commands may also change the context of
	associatated commands to provide semantics without altering the basic
	syntax.

SEE ALSO
	cli escapes, cli options`,
	}
}
