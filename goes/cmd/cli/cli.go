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
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/go/internal/shellutils"
	"github.com/platinasystems/go/internal/url"
)

type parsedCommand struct {
	env  []string
	args []string
}

type Command struct {
	Prompt string
	g      *goes.Goes
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

func (c *Command) Goes(g *goes.Goes) { c.g = g }

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
			Close()
		}

		isScript bool
	)

	if c.g == nil {
		panic("cli's goes is nil")
	}

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
			prompter = notliner.New(os.Stdin, nil)
			isScript = true
		case flag.ByName["-no-liner"]:
			prompter = notliner.New(os.Stdin, os.Stdout)
		default:
			if _, found := c.g.ByName["resize"]; !found {
				c.g.ByName["resize"] = resize.Command{}
			}
			prompter = liner.New(c.g)
			defer prompter.Close()
		}
	case 1:
		script, err := url.Open(args[0])
		if err != nil {
			return err
		}
		defer script.Close()
		prompter = notliner.New(script, nil)
		defer prompter.Close()
		isScript = true
	default:
		return fmt.Errorf("%v: unexpected", args[1:])
	}

	c.g.Catline = func(prompt string) (string, error) {
		s, err := prompter.Prompt(prompt)
		if err != nil {
			return "", err
		}
		return s, nil
	}

	signal.Ignore(syscall.SIGINT)
readCommandLoop:
	for {
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
			fmt.Fprintln(os.Stderr, err)
			if isScript && !flag.ByName["-f"] {
				return nil
			}
			continue readCommandLoop
		}
		// loop for each pipeline in command list
		skipNext := false
		for _, pl := range cl.Cmds {
			term := pl.Cmds[len(pl.Cmds)-1].Term
			if !skipNext {
			pipelineLoop:
				for {
					plSlice := make([]parsedCommand, 0)
					// loop for each command in pipeline
					for _, sl := range pl.Cmds {
						envMap, cmdline := sl.Slice(func(k string) string {
							v, def := c.g.EnvMap[k]
							if def {
								return v
							}
							return os.Getenv(k)
						})
						// Add to our context environment if this command only set variables
						if len(cmdline) == 0 {
							if len(envMap) != 0 {
								if c.g.EnvMap == nil {
									c.g.EnvMap = envMap
								} else {
									for k, v := range envMap {
										c.g.EnvMap[k] = v
									}
								}
								c.g.Status = nil // Successfully set variables
							}
							break pipelineLoop
						}
						var envStr []string
						if len(envMap) != 0 {
							envStr = make([]string, 0)
							for k, v := range envMap {
								envStr = append(envStr, fmt.Sprintf("%s=%s", k, v))
							}
						}
						plSlice = append(plSlice, parsedCommand{env: envStr, args: cmdline})

						name := cmdline[0]
						if v := c.g.ByName[name]; v != nil {
							k := cmd.WhatKind(v)
							if k.IsDaemon() {
								err = fmt.Errorf(
									"use `goes-daemons start %s`",
									name)
								break pipelineLoop
							}
							if len(pl.Cmds) > 1 {
								if k.IsCantPipe() {
									err = fmt.Errorf(
										"%s: can't pipe", name)
									break pipelineLoop
								}
							} else if k.IsDontFork() ||
								name == os.Args[0] {
								if flag.ByName["-x"] {
									fmt.Println("+", sl)
								}
								err = c.g.Main(cmdline...)
								break pipelineLoop
							}
						}
					}
					iparm, args := parms.New(plSlice[0].args, "<", "<<", "<<-")
					plSlice[0].args = args
					in = os.Stdin
					if fn := iparm.ByName["<"]; len(fn) > 0 {
						rc, err = url.Open(fn)
						if err != nil {
							break pipelineLoop
						}
						in = rc
						closers = append(closers, rc)
					} else if len(iparm.ByName["<<"]) > 0 ||
						len(iparm.ByName["<<-"]) > 0 {
						var trim bool
						lbl := iparm.ByName["<<"]
						if len(lbl) == 0 {
							lbl = iparm.ByName["<<-"]
							trim = true
						}
						var r, w *os.File
						r, w, err = os.Pipe()
						if err != nil {
							break pipelineLoop
						}
						in = r
						closers = append(closers, r)
						go func(w io.WriteCloser, lbl string) {
							defer w.Close()
							prompt := "<<" + fn + " "
							for {
								s, err := c.g.Catline(prompt)
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

					end := len(plSlice) - 1
					oparm, args := parms.New(plSlice[end].args, ">", ">>", ">>>", ">>>>")
					plSlice[end].args = args
					out = os.Stdout
					if fn := oparm.ByName[">"]; len(fn) > 0 {
						wc, err = url.Create(fn)
						if err != nil {
							break pipelineLoop
						}
						out = wc
						closers = append(closers, wc)
					} else if fn = oparm.ByName[">>"]; len(fn) > 0 {
						wc, err = url.Append(fn)
						if err != nil {
							break pipelineLoop
						}
						out = wc
						closers = append(closers, wc)
					} else if fn := oparm.ByName[">>>"]; len(fn) > 0 {
						wc, err = url.Create(fn)
						if err != nil {
							break pipelineLoop
						}
						out = io.MultiWriter(os.Stdout, wc)
						closers = append(closers, wc)
					} else if fn := oparm.ByName[">>"]; len(fn) > 0 {
						wc, err = url.Append(fn)
						if err != nil {
							break pipelineLoop
						}
						out = io.MultiWriter(os.Stdout, wc)
						closers = append(closers, wc)
					}

					for i, sl := range plSlice {
						if flag.ByName["-x"] {
							fmt.Println("+", strings.Join(sl.env, " "), strings.Join(sl.args, " "))
						}
						x := c.g.Fork(sl.args...)
						if x == nil {
							continue
						}
						if len(sl.env) != 0 {
							x.Env = os.Environ()
							for _, s := range sl.env {
								x.Env = append(x.Env, s)
							}
						}
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
								break pipelineLoop
							}
							x.Stdout = pout
						}
						if err = x.Start(); err != nil {
							err = fmt.Errorf("child: %v: %v", x.Args, err)
							break pipelineLoop
						}
						if i == end {
							err = x.Wait()
							c.g.Status = err
						} else {
							go func(x *exec.Cmd) {
								err := x.Wait()
								if err != nil {
									fmt.Fprintln(os.Stderr, err)
								}
								c.g.Status = err
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
					} // end
					break pipelineLoop
				} // end pipelineLoop
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
					if isScript && !flag.ByName["-f"] {
						return nil
					}
					err = nil
				}
			}
			skipNext = false
			if c.g.Status != nil {
				if term.String() == "&&" {
					skipNext = true
				}
			} else {
				if term.String() == "||" {
					skipNext = true
				}
			}
		} // loop for each pipeline in command list
	} // loop forever reading commands
	return fmt.Errorf("oops, shouldn't be here")
}
