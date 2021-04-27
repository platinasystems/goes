// Copyright © 2021 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package forcmd

import (
	"errors"
	"fmt"
	"io"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/internal/shellutils"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "for" }

func (Command) Usage() string {
	return "for VAR in ARGS... ; do COMMAND $VAR; done"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "loop over a set of arguments and run a command",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Iterate over a series of words for a set of commands.`,
	}
}

func (c Command) Block(g *goes.Goes, ls shellutils.List) (*shellutils.List, func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error, error) {
	var doList []func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error
	cl := ls.Cmds[0]

	// for <var> in <commands>
	if len(cl.Cmds) > 1 {
		cl.Cmds = cl.Cmds[1:]
		ls.Cmds[0] = cl
	} else {
		return nil, nil, errors.New("Unexpected `newline'")
	}
	varName := cl.Cmds[0].String()
	if varName == "" {
		return nil, nil, errors.New("Malformed loop variable")
	}

	cl.Cmds = cl.Cmds[1:]
	if len(cl.Cmds) > 1 {
		ls.Cmds[0] = cl
	} else {
		ls.Cmds = ls.Cmds[1:]
	}

	foundIn := false
	foundDo := false
	var wordList []shellutils.Word

	for {
		if len(cl.Cmds) == 0 {
			for len(ls.Cmds) == 0 {
				newls, err := shellutils.Parse("for>", g.Catline)
				if err != nil {
					return nil, nil, err
				}
				ls = *newls
			}
			cl = ls.Cmds[0]
		}
		name := cl.Cmds[0].String()

		if !foundIn {
			if name != "in" {
				return nil, nil, fmt.Errorf("Expected `in' found `%s'",
					name)
			}
			foundIn = true
			cl.Cmds = cl.Cmds[1:]
			if len(cl.Cmds) > 0 {
				ls.Cmds[0] = cl
			} else {
				ls.Cmds = ls.Cmds[1:]
			}
			continue
		}

		if wordList == nil {
			wordList = cl.Cmds
			cl.Cmds = nil
			ls.Cmds = ls.Cmds[1:]
			continue
		}

		if !foundDo {
			if name != "do" {
				return nil, nil, fmt.Errorf("Looking for `do' got `%s'",
					name)
			}
			foundDo = true
			if len(cl.Cmds) > 1 {
				cl.Cmds = cl.Cmds[1:]
				ls.Cmds[0] = cl
			} else {
				ls.Cmds = ls.Cmds[1:]
			}
			continue
		}

		if name == "done" {
			if len(cl.Cmds) > 1 {
				return nil, nil, errors.New("unexpected text after `done'")
			}
			break
		}

		nextls, _, runfun, err := g.ProcessList(ls)
		if err != nil {
			return nil, nil, err
		}
		doList = append(doList, runfun)
		cl.Cmds = nil
		ls = *nextls
		if len(ls.Cmds) != 0 {
			cl = ls.Cmds[0]
		}
	}
	blockfun, err := makeBlockFunc(g, varName, wordList, doList)

	return &ls, blockfun, err
}

func runList(pipeline []func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	for _, runent := range pipeline {
		err := runent(stdin, stdout, stderr)
		if err != nil {
			return err
		}
	}
	return nil
}

func makeBlockFunc(g *goes.Goes, varName string,
	wordList []shellutils.Word,
	doList []func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error) (func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error, error) {
	if g.EnvMap == nil {
		g.EnvMap = make(map[string]string)
	}
	runfun := func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
		for _, word := range wordList {
			for _, str := range word.Expand() {
				g.EnvMap[varName] = str
				err := runList(doList, stdin, stdout, stderr)
				if err != nil {
					fmt.Fprintln(stderr, err)
				}
				if g.Status != nil {
					if g.Status.Error() == "signal: interrupt" {
						return g.Status
					}
				}
			}
		}
		return nil
	}
	return runfun, nil
}

func (Command) Main(args ...string) error {
	return errors.New("internal error")
}
