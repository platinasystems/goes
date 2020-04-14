// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/cmd/cli/internal/liner"
	"github.com/platinasystems/goes/cmd/cli/internal/notliner"
	"github.com/platinasystems/goes/cmd/resize"
	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/internal/shellutils"
	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/url"
)

type parsedCommand struct {
	env  []string
	args []string
}

type Command struct {
	Prompt       string
	g            *goes.Goes
	promptString string
	prompter     interface {
		Prompt(string) (string, error)
		Close()
	}
	Stdin          io.Reader
	Stdout, Stderr io.Writer
}

func (*Command) String() string { return "cli" }

func (*Command) Usage() string {
	return "cli [-x] [-p PROMPT] [URL]"
}

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "command line interpreter",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
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

	The '-x' flag enables trace of each interpreted command.

	With 'URL', commands are sourced from the reference instead of prompted
	tty input.

COMMENTS
	Hash tag prefaced comments are ignored, e.g.:
		mount -t tmpfs none /tmp # scratch
	or,
		# ignored line...

WHITESPACE
	Leading whitespace is ignored, e.g.:

			echo hello

	However, the cli doesn't have command blocks so there isn't much
	reason to indent input for anything other than "here documents".

ESCAPES
	A COMMAND may extend to multiple lines by escaping the end of
	line with the backslash character ('\').

		echo ..............\
		..............\
		..............

	Similarly, the space between arguments may be escaped.

		echo with\ one\ argument\ having\ five\ spaces

QUOTATION
	Arguments may be single or double quoted.

		echo 'with two arguments' each "having two spaces"
		echo "hello 'beautiful world'"
		echo 'hello \"beautiful world\"'

	But *not*,

		echo 'hello "beautiful world"'

SPECIAL CHARACTERS
.	The command may encode these special characters.

		\a   U+0007 alert or bell
		\b   U+0008 backspace
		\f   U+000C form feed
		\n   U+000A line feed or newline
		\r   U+000D carriage return
		\t   U+0009 horizontal tab
		\v   U+000b vertical tab
		\\   U+005c backslash

	The command may also encode any byte or unicode rune with these.

		\OOO	where OOO are three octal digits
		\xXX	where XX are two hex digits
		\uXXXX
		\UXXXXXXXX

	Finally, the command line may include any unicode rune literal
	supported by Go.
	
		ä 本 日本語

OPTIONS
	These common options manipluate the CLI command context.

	> URL	Redirect stdout to URL.

	>> URL
		Append command output to URL.

	>>> URL
	>>>> URL
		Print or append output to URL in addition to stdout.

	< URL	Redirect stdin from URL.

	<<[-] LABEL
		Read command script upto LABEL as stdin. If LABEL is prefaced
		by '-', the leading whitespace is trimmed from each line.

	Note: unlike other shells, there must be a space or equal ('=')
	between the redirection symbols and URL or LABEL.

PIPES
	The COMMAND output may be piped to the input of another COMMAND, e.g.:
		ls -lR | more
		ls -Lr |
		more

	The COMMAND pipeline may redirect input and output of the first and
	last commands respectively, e.g.:

		cat <<- EOF | wc -l > lines.txt
			...
		EOF`,
	}
}

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (c *Command) Read(p []byte) (n int, err error) {
	s, err := c.prompter.Prompt(c.promptString)
	if err != nil {
		return
	}
	n = copy(p, s)
	if len(s) > len(p) {
		err = errors.New("input too long")
	}
	return
}

func (c *Command) Write(p []byte) (n int, err error) {
	c.promptString = string(p)
	return len(c.promptString), nil
}

func (c *Command) Main(args ...string) error {
	var (
		err      error
		isScript bool
	)

	if c.g == nil {
		panic("cli's goes is nil")
	}

	if c.Stdin == nil {
		c.Stdin = os.Stdin
	}
	if c.Stdout == nil {
		c.Stdout = os.Stdout
	}
	if c.Stderr == nil {
		c.Stderr = os.Stderr
	}
	csig := make(chan os.Signal, 1)
	signal.Notify(csig, os.Interrupt)

	defer func() {
		for _, name := range c.g.Names() {
			v := c.g.ByName[name]
			k := cmd.WhatKind(v)
			if k.IsDontFork() {
				if m, found := v.(io.Closer); found {
					t := m.Close()
					if err == nil {
						err = t
					}
				}
			}
		}
	}()

	flag, args := flags.New(args, "-f", "-x", "-", "-no-liner")
	switch len(args) {
	case 0:
		switch {
		case flag.ByName["-"]:
			c.prompter = notliner.New(c.Stdin, nil)
			isScript = true
		case flag.ByName["-no-liner"]:
			c.prompter = notliner.New(c.Stdin, c.Stdout)
		default:
			if _, found := c.g.ByName["resize"]; !found {
				c.g.ByName["resize"] = resize.Command{}
			}
			c.prompter = liner.New(c.g)
			defer c.prompter.Close()
		}
	case 1:
		script, err := url.Open(args[0])
		if err != nil {
			return err
		}
		defer script.Close()
		c.prompter = notliner.New(script, nil)
		defer c.prompter.Close()
		isScript = true
	default:
		return fmt.Errorf("%v: unexpected", args[1:])
	}

	if flag.ByName["-f"] && c.g.Verbosity < goes.VerboseVerify {
		c.g.Verbosity = goes.VerboseVerify
	}
	if c.g.Catline == nil {
		c.g.Catline = c
	}
readCommandLoop:
	for {
		select {
		case <-csig:
			fmt.Println("\nCommand interrupted")
		default:
		}
		prompt := c.Prompt
		if len(prompt) == 0 {
			prompt = fmt.Sprint(c.g, "> ")
			if len(c.g.Path()) == 0 {
				if hn, err := os.Hostname(); err == nil {
					prompt = fmt.Sprint(hn, "> ")
				}
			}
		}
		cl, err := shellutils.Parse(prompt, c.g.Catline)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			fmt.Fprintln(c.Stderr, err)
			if isScript && !flag.ByName["-f"] {
				return nil
			}
			continue readCommandLoop
		}
		err = c.runList(*cl, flag, isScript)
		if err != nil {
			if isScript && !flag.ByName["-f"] {
				return err
			} else {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
}

func (c *Command) runList(ls shellutils.List, flag *flags.Flags, isScript bool) (err error) {
	// loop for each pipeline in command list
	for len(ls.Cmds) != 0 {
		newls, _, runner, err := c.g.ProcessList(ls)
		if err == nil {
			err = runner(c.Stdin, c.Stdout, c.Stderr)
		}
		if err != nil {
			return err
		}
		if newls != nil {
			ls = *newls
		} else {
			break
		}
	}
	return nil
}
