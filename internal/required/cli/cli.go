// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package cli provides a command line interface.
//
// Commands may continue to next line, e.g.:
//
//	echo hello \
//	world
//
// Commands may be pipelined, e.g.:
//
//	ls -lR | more
//	ls -Lr |
//	more
//
// Command comments are ignored, e.g.:
//
//	mount -t tmpfs none /tmp # scratch
//
// Similar for leading whitespace, e.g.:
//
//		echo why\?
//
// However, this cli doesn't have command blocks so there isn't much reason to
// indent input for anything other than "here documents".
//
// A pipeline may redirect input and output of the first and last commands
// respectively, e.g.:
//
//	cat <<-EOF | wc -l > lines.txt
//		...
//	EOF
//
// The redirected files may be URL's, e.g.:
//
//	source https://github.com/MYSTUFF/MYSCRIPT
//
// Redirected output may be tee'd to a truncated or appended file with `>>>`
// and `>>>>` respectively, e.g.:
//
//	dmesg | grep goes >>> goes.log
package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/go/internal/required/cli/internal/liner"
	"github.com/platinasystems/go/internal/required/cli/internal/nocomment"
	"github.com/platinasystems/go/internal/required/cli/internal/notliner"
	"github.com/platinasystems/go/internal/required/cli/internal/slice_args"
	"github.com/platinasystems/go/internal/required/cli/internal/slice_string"
	"github.com/platinasystems/go/internal/url"
)

const Name = "cli"

type cmd goes.ByName

func New() *cmd { return new(cmd) }

func (*cmd) Kind() goes.Kind { return goes.DontFork | goes.CantPipe }
func (*cmd) String() string  { return Name }
func (*cmd) Usage() string   { return "cli [-x] [URL]" }

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (c *cmd) Main(args ...string) error {
	var (
		rc  io.ReadCloser
		wc  io.WriteCloser
		in  io.Reader
		out io.Writer
		err error

		closers []io.Closer

		pin, pout *os.File

		prompter interface {
			Prompt(string) (string, error)
		}
	)
	defer func() {
		for _, g := range goes.ByName(*c) {
			if g.Kind.IsDontFork() && g.Close != nil {
				t := g.Close()
				if err == nil {
					err = t
				}
			}
		}
	}()
	flag, args := flags.New(args, "-x", "-no-liner")
	switch len(args) {
	case 0:
		if flag["-no-liner"] {
			prompter = notliner.New(os.Stdin, os.Stdout)
		} else {
			prompter = liner.New(goes.ByName(*c))
		}
	case 1:
		script, err := url.Open(args[0])
		if err != nil {
			return err
		}
		defer script.Close()
		prompter = notliner.New(script, nil)
	default:
		return fmt.Errorf("%v: unexpected", args[1:])
	}

	pl := slice_args.New("|")
	catline := func(prompt string) (string, error) {
		var line string
		for {
			s, err := prompter.Prompt(prompt)
			if err != nil {
				return "", err
			}
			if !strings.HasSuffix(s, "\\") {
				return line + s, nil
			}
			line += s[:len(s)-1]
			prompt = "... "
		}
	}

commandLoop:
	for {
		for _, c := range closers {
			c.Close()
		}
		closers = closers[:0]
		if err != nil {
			if err == io.EOF {
				return nil
			}
			if err.Error() != "exit status 1" {
				fmt.Fprintln(os.Stderr, err)
			}
			err = nil
		}
		pl.Reset()
		prompt := filepath.Base(goes.Prog()) + "> "
		if hn, err := os.Hostname(); err == nil {
			prompt = hn + "> "
		}
	pipelineLoop:
		for {
			var s string
			s, err = catline(prompt)
			if err != nil {
				return err
			}
			s = strings.TrimLeft(s, " \t")
			if len(s) == 0 {
				continue pipelineLoop
			}
			s = nocomment.New(s)
			if len(s) == 0 {
				continue pipelineLoop
			}
			pl.Slice(slice_string.New(s)...)
			if pl.More {
				prompt = "| "
			} else {
				break pipelineLoop
			}
		}
		if len(pl.Slices) == 0 {
			continue commandLoop
		}
		for i := 0; i < len(pl.Slices); i++ {
			name := pl.Slices[i][0]
			g := goes.ByName(*c)[name]
			if g == nil {
				err = fmt.Errorf("%s: not found", name)
				continue commandLoop
			}
			if !g.Kind.IsInteractive() {
				err = fmt.Errorf("%s: inoperative", name)
				continue commandLoop
			}
			if len(pl.Slices) > 0 {
				if g.Kind.IsCantPipe() {
					err = fmt.Errorf("%s: can't pipe", name)
					continue commandLoop
				}
			}
		}
		if len(pl.Slices) == 1 {
			name := pl.Slices[0][0]
			g := goes.ByName(*c)[name]
			if g.Kind.IsDontFork() || name == os.Args[0] {
				if flag["-x"] {
					fmt.Println("+", pl.Slices[0])
				}
				err = goes.ByName(*c).Main(pl.Slices[0]...)
				continue commandLoop
			}
		}

		iparm, args := parms.New(pl.Slices[0], "<", "<<", "<<-")
		pl.Slices[0] = args
		in = nil
		if fn := iparm["<"]; len(fn) > 0 {
			rc, err = url.Open(fn)
			if err != nil {
				continue commandLoop
			}
			in = rc
			closers = append(closers, rc)
		} else if len(iparm["<<"]) > 0 || len(iparm["<<-"]) > 0 {
			var trim bool
			lbl := iparm["<<"]
			if len(lbl) == 0 {
				lbl = iparm["<<-"]
				trim = true
			}
			var r, w *os.File
			r, w, err = os.Pipe()
			if err != nil {
				continue commandLoop
			}
			in = r
			closers = append(closers, r)
			go func(w io.WriteCloser, lbl string) {
				defer w.Close()
				prompt := "<<" + fn + " "
				for {
					s, err := catline(prompt)
					if err != nil || s == lbl {
						break
					}
					if trim {
						s = strings.TrimLeft(s, " \t")
					}
					fmt.Fprintln(w, s)
				}
			}(w, lbl)
		}

		end := len(pl.Slices) - 1
		oparm, args := parms.New(pl.Slices[end],
			">", ">>", ">>>", ">>>>")
		pl.Slices[end] = args
		out = os.Stdout
		if fn := oparm[">"]; len(fn) > 0 {
			wc, err = url.Create(fn)
			if err != nil {
				continue commandLoop
			}
			out = wc
			closers = append(closers, wc)
		} else if fn = oparm[">>"]; len(fn) > 0 {
			wc, err = url.Append(fn)
			if err != nil {
				continue commandLoop
			}
			out = wc
			closers = append(closers, wc)
		} else if fn := oparm[">>>"]; len(fn) > 0 {
			wc, err = url.Create(fn)
			if err != nil {
				continue commandLoop
			}
			out = io.MultiWriter(os.Stdout, wc)
			closers = append(closers, wc)
		} else if fn := oparm[">>"]; len(fn) > 0 {
			wc, err = url.Append(fn)
			if err != nil {
				continue commandLoop
			}
			out = io.MultiWriter(os.Stdout, wc)
			closers = append(closers, wc)
		}

		for i := 0; i < len(pl.Slices); i++ {
			if flag["-x"] {
				fmt.Println("+", pl.Slices[i])
			}
			c := exec.Command(goes.Prog(), pl.Slices[i][1:]...)
			c.Args[0] = pl.Slices[i][0]
			c.Stderr = os.Stderr
			if i == 0 {
				c.Stdin = in
			} else {
				c.Stdin = pin
			}
			if i == end {
				c.Stdout = out
			} else {
				pin, pout, err = os.Pipe()
				if err != nil {
					continue commandLoop
				}
				c.Stdout = pout
			}
			if err = c.Start(); err != nil {
				err = fmt.Errorf("child: %v: %v", c.Args, err)
				continue commandLoop
			}
			if i == end {
				err = c.Wait()
			} else {
				go func(c *exec.Cmd) {
					err := c.Wait()
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
					}
					if c.Stdout != os.Stdout {
						m, found := c.Stdout.(io.Closer)
						if found {
							m.Close()
						}
					}
					if c.Stdin != os.Stdin {
						m, found := c.Stdin.(io.Closer)
						if found {
							m.Close()
						}
					}
				}(c)
			}
		}
	}
	return fmt.Errorf("oops, shouldn't be here")
}

func (*cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "command line interpreter",
	}
}

func (*cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	cli - command line interpreter

SYNOPSIS
	cli [-x] [URL]

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
	The command may encode these special characters.

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
