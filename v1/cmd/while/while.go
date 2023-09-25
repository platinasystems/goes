// Copyright © 2021 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package whilecmd

import (
	"errors"
	"fmt"
	"io"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/internal/shellutils"
	"github.com/platinasystems/goes/lang"
)

type Command struct {
	IsUntil bool
}

func (c Command) String() string {
	if c.IsUntil {
		return "until"
	}
	return "while"
}

func (c Command) Usage() string {
	return c.String() + " COND ; do COMMAND ; done"
}

func (c Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "Loop " + c.String() + " a condition is true",
	}
}

func (c Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Executes a set of commands ` + c.String() + ` another returns success`,
	}
}

func (c Command) Block(g *goes.Goes, ls shellutils.List) (*shellutils.List, func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error, error) {
	var whileList, doList []func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error
	curList := &whileList
	cl := ls.Cmds[0]
	// while <command>
	if len(cl.Cmds) > 1 {
		cl.Cmds = cl.Cmds[1:]
		ls.Cmds[0] = cl
	} else {
		ls.Cmds = ls.Cmds[1:]
	}
	for {
		nextls, _, runfun, err := g.ProcessList(ls)
		if err != nil {
			return nil, nil, err
		}
		*curList = append(*curList, runfun)
		ls = *nextls
		for len(ls.Cmds) == 0 {
			newls, err := shellutils.Parse("while>", g.Catline)
			if err != nil {
				return nil, nil, err
			}
			ls = *newls
		}
		cl := ls.Cmds[0]
		name := cl.Cmds[0].String()
		if name == "do" {
			if curList != &whileList {
				return nil, nil, errors.New("Unexpected 'do'")
			}
			curList = &doList
			if len(cl.Cmds) > 1 {
				cl.Cmds = cl.Cmds[1:]
				ls.Cmds[0] = cl
			} else {
				ls.Cmds = ls.Cmds[1:]
			}
		}
		if name == "done" {
			if curList != &doList {
				return nil, nil, errors.New("Unexpected 'done'")
			}
			if len(cl.Cmds) > 1 {
				return nil, nil, errors.New("unexpected text after done")
			}
			break
		}

	}
	blockfun, err := c.makeBlockFunc(g, whileList, doList)

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

func (c Command) makeBlockFunc(g *goes.Goes, whileList, doList []func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error) (func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error, error) {
	runfun := func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
		for {
			err := runList(whileList, stdin, stdout, stderr)
			if (err == nil && g.Status == nil) != c.IsUntil {
				err = runList(doList, stdin, stdout, stderr)
				if err != nil {
					fmt.Fprintln(stderr, err)
				}
				if g.Status != nil {
					if g.Status.Error() == "signal: interrupt" {
						return g.Status
					}
				}
			} else {
				return err
			}
		}
	}
	return runfun, nil
}

func (Command) Main(args ...string) error {
	return errors.New("internal error")
}
