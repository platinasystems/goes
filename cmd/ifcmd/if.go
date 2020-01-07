// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ifcmd

import (
	"errors"
	"io"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/internal/shellutils"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "if" }

func (Command) Usage() string {
	return "if COMMAND ; then COMMAND else COMMAND endif"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "conditional command",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Conditionally executes statements in a script`,
	}
}

func (c Command) Block(g *goes.Goes, ls shellutils.List) (*shellutils.List, func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error, error) {
	var ifList, thenList, elseList []func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error
	curList := &ifList
	cl := ls.Cmds[0]
	// if <command>
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
			newls, err := shellutils.Parse("if>", g.Catline)
			if err != nil {
				return nil, nil, err
			}
			ls = *newls
		}
		cl := ls.Cmds[0]
		name := cl.Cmds[0].String()
		if name == "then" {
			if curList != &ifList {
				return nil, nil, errors.New("Unexpected 'then'")
			}
			curList = &thenList
			if len(cl.Cmds) > 1 {
				cl.Cmds = cl.Cmds[1:]
				ls.Cmds[0] = cl
			} else {
				ls.Cmds = ls.Cmds[1:]
			}
		}
		if name == "else" {
			if curList != &thenList {
				return nil, nil, errors.New("Unexpected 'else'")
			}
			curList = &elseList
			if len(cl.Cmds) > 1 {
				cl.Cmds = cl.Cmds[1:]
				ls.Cmds[0] = cl
			} else {
				ls.Cmds = ls.Cmds[1:]
			}
		}
		if name == "elif" {
			if curList != &thenList {
				return nil, nil, errors.New("Unexpected 'elif'")
			}
			newls, elifFun, err := c.Block(g, ls)
			if err != nil {
				return nil, nil, err
			}
			runfun := func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
				g.Status = nil
				return elifFun(stdin, stdout, stderr)
			}
			elseList = append(elseList, runfun)
			ls = *newls
			break
		}
		if name == "fi" {
			if curList != &thenList && curList != &elseList {
				return nil, nil, errors.New("Unexpected 'fi'")
			}
			if len(cl.Cmds) > 1 {
				return nil, nil, errors.New("unexpected text after fi")
			}
			break
		}

	}
	blockfun, err := makeBlockFunc(g, ifList, thenList, elseList)

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

func makeBlockFunc(g *goes.Goes, ifList, thenList, elseList []func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error) (func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error, error) {
	runfun := func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
		err := runList(ifList, stdin, stdout, stderr)
		if err == nil && g.Status == nil {
			err = runList(thenList, stdin, stdout, stderr)
		} else {
			err = runList(elseList, stdin, stdout, stderr)
		}
		return err
	}
	return runfun, nil
}

func (Command) Main(args ...string) error {
	return errors.New("internal error")
}
