// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package cli_options

const Name = "cli-options"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Tag() string    { return "builtin" }
func (cmd) Usage() string  { return "man " + Name }

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `OPTIONS
	These common options manipluate the CLI command context.

	> FILE	Redirect stdout to FILE.

	>> FILE
		Append command output to FILE.

	>>> FILE
	>>>> FILE
		Print or append output to FILE in addition to stdout.

	< FILE	Redirect stdin from FILE.

	<<[-] LABEL
		Read command script upto LABEL as stdin. If LABEL is prefaced
		by '-', the leading whitespace is trimmed from each line.

	Note: unlike other shells, there must be a space or equal ('=')
	between the redirection symbols and FILE or LABEL.`,
	}
}
