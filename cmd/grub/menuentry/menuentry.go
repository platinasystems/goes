// Copyright © 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package menuentry

import (
	"errors"
	"io"

	"github.com/platinasystems/goes"

	"github.com/platinasystems/goes/lang"

	"github.com/platinasystems/flags"
	"github.com/platinasystems/parms"
	"github.com/platinasystems/goes/internal/shellutils"
)

type Entry struct {
	Name   string
	RunFun func(stdin io.Reader, stdout io.Writer, stderr io.Writer, isFirst bool, isLast bool) error
}

type Command struct {
	Menus []Entry
}

func (Command) String() string { return "menuentry" }

func (Command) Usage() string {
	return "NOP"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "NOP",
	}
}

func (c *Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: Man,
	}
}

const Man = "NOP command for script compatibility\n"

func (c *Command) Block(g *goes.Goes, ls shellutils.List) (*shellutils.List, func(stdin io.Reader, stdout io.Writer, stderr io.Writer, isFirst bool, isLast bool) error, error) {
	cl := ls.Cmds[0]
	// menuentry name [option...] { definition ... ; }
	if len(cl.Cmds) < 2 {
		return nil, nil, errors.New("Menuentry: unexpected end of line")
	}

	name := cl.Cmds[1].String()
	cl.Cmds = cl.Cmds[2:]
	args := make([]string, 0)
	foundBrace := false

	for len(cl.Cmds) > 0 {
		cmd := cl.Cmds[0].String()
		cl.Cmds = cl.Cmds[1:]
		if cmd == "{" {
			foundBrace = true
			break
		}
		args = append(args, cmd)
	}

	if !foundBrace {
		return nil, nil, errors.New("menuentry: missing {")
	}

	_, args = parms.New(args, "--class", "--users", "--hotkey", "--id")
	_, args = flags.New(args, "--unrestricted")

	//	fmt.Printf("menuentry: name: %v, parm: %v, flags: %v, args: %v\n",
	//		name, parm, flags, args)

	if len(cl.Cmds) > 0 {
		ls.Cmds[0] = cl
	} else {
		ls.Cmds = ls.Cmds[1:]
	}

	var funList []func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error
	for {
		nextls, _, runfun, err := g.ProcessList(ls)
		if err != nil {
			return nil, nil, err
		}
		funList = append(funList, runfun)
		ls = *nextls
		for len(ls.Cmds) == 0 {
			newls, err := shellutils.Parse("menuentry>", g.Catline)
			if err != nil {
				return nil, nil, err
			}
			ls = *newls
		}
		cl := ls.Cmds[0]
		name := cl.Cmds[0].String()
		if name == "}" {
			if len(cl.Cmds) > 1 {
				return nil, nil, errors.New("unexpected text after }")
			}
			break
		}

	}

	runfun := func(stdin io.Reader, stdout io.Writer, stderr io.Writer, isFirst bool, isLast bool) error {
		for _, runent := range funList {
			err := runent(stdin, stdout, stderr)
			if err != nil {
				return err
			}
		}
		return nil
	}
	e := Entry{Name: name, RunFun: runfun}

	deffun := func(stdin io.Reader, stdout io.Writer, stderr io.Writer, isFirst bool, isLast bool) error {
		c.Menus = append(c.Menus, e)
		return nil
	}
	return &ls, deffun, nil
}

func (Command) Main(args ...string) error {
	return errors.New("internal error")
}
