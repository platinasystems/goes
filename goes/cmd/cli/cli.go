// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/cmd/cli/internal/liner"
	"github.com/platinasystems/go/goes/cmd/cli/internal/notliner"
	"github.com/platinasystems/go/goes/cmd/resize"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/fields"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/nocomment"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/go/internal/pizza"
	"github.com/platinasystems/go/internal/url"
)

const (
	Name    = "cli"
	Apropos = "command line interpreter"
	Usage   = "cli [-x] [-p PROMPT] [URL]"
	Man     = `
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
		EOF`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func New() []cmd.Cmd {
	return []cmd.Cmd{new(Command), resize.New()}
}

type Command struct {
	g *goes.Goes
}

func (*Command) Apropos() lang.Alt   { return apropos }
func (c *Command) Goes(g *goes.Goes) { c.g = g }
func (*Command) Man() lang.Alt       { return man }
func (*Command) String() string      { return Name }
func (*Command) Usage() string       { return Usage }

func (c *Command) Main(args ...string) error {
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

		isScript bool
	)

	defer func() {
		for _, name := range c.g.Names {
			v := c.g.ByName(name)
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
		case flag["-"]:
			prompter = notliner.New(os.Stdin, nil)
			isScript = true
		case flag["-no-liner"]:
			prompter = notliner.New(os.Stdin, os.Stdout)
		default:
			prompter = liner.New(c.g)
		}
	case 1:
		script, err := url.Open(args[0])
		if err != nil {
			return err
		}
		defer script.Close()
		prompter = notliner.New(script, nil)
		isScript = true
	default:
		return fmt.Errorf("%v: unexpected", args[1:])
	}

	pl := pizza.New("|")
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

	signal.Ignore(syscall.SIGINT)
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
			if isScript && !flag["-f"] {
				return nil
			}
			err = nil
		}
		pl.Reset()
		prompt := fmt.Sprint(c.g, "> ")
		if c.g.Parent == nil {
			if hn, err := os.Hostname(); err == nil {
				prompt = fmt.Sprint(hn, "> ")
			}
		}
	pipelineLoop:
		for {
			var s string
			s, err = catline(prompt)
			if err != nil {
				continue commandLoop
			}
			s = strings.TrimLeft(s, " \t")
			if len(s) == 0 {
				continue pipelineLoop
			}
			s = nocomment.New(s)
			if len(s) == 0 {
				continue pipelineLoop
			}
			pl.Slice(fields.New(s)...)
			if pl.More {
				prompt = "| "
			} else {
				break pipelineLoop
			}
		}
		if len(pl.Slices) == 0 {
			continue commandLoop
		}
		for _, sl := range pl.Slices {
			name := sl[0]
			if v := c.g.ByName(name); v != nil {
				k := cmd.WhatKind(v)
				if !k.IsInteractive() {
					err = fmt.Errorf("%s: inoperative",
						name)
					continue commandLoop
				}
				if len(pl.Slices) > 1 {
					if k.IsCantPipe() {
						err = fmt.Errorf(
							"%s: can't pipe", name)
						continue commandLoop
					}
				} else if k.IsDontFork() ||
					name == os.Args[0] {
					if flag["-x"] {
						fmt.Println("+", sl)
					}
					err = c.g.Main(sl...)
					continue commandLoop
				}
			}
		}
		iparm, args := parms.New(pl.Slices[0], "<", "<<", "<<-")
		pl.Slices[0] = args
		in = os.Stdin
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

		for i, sl := range pl.Slices {
			if flag["-x"] {
				fmt.Println("+", strings.Join(sl, " "))
			}
			x := c.g.Fork(sl...)
			x.Stderr = os.Stderr
			if i == 0 {
				x.Stdin = in
			} else {
				x.Stdin = pin
			}
			if i == end {
				x.Stdout = out
			} else {
				pin, pout, err = os.Pipe()
				if err != nil {
					continue commandLoop
				}
				x.Stdout = pout
			}
			if err = x.Start(); err != nil {
				err = fmt.Errorf("child: %v: %v", x.Args, err)
				continue commandLoop
			}
			if i == end {
				err = x.Wait()
			} else {
				go func(x *exec.Cmd) {
					err := x.Wait()
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
					}
					if x.Stdout != os.Stdout {
						m, found := x.Stdout.(io.Closer)
						if found {
							m.Close()
						}
					}
					if x.Stdin != os.Stdin {
						m, found := x.Stdin.(io.Closer)
						if found {
							m.Close()
						}
					}
				}(x)
			}
		}
	}
	return fmt.Errorf("oops, shouldn't be here")
}
