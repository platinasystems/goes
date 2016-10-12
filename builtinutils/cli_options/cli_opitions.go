// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package cli_options

type cliOptions struct{}

func New() cliOptions { return cliOptions{} }

func (cliOptions) String() string { return "cli-options" }
func (cliOptions) Tag() string    { return "builtin" }
func (cliOptions) Usage() string  { return "man cli-options" }

func (cliOptions) Man() map[string]string {
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
