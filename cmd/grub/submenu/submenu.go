// Copyright © 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package submenu

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd/grub/menuentry"
	"github.com/platinasystems/goes/lang"

	"github.com/platinasystems/flags"
	"github.com/platinasystems/parms"
	"github.com/platinasystems/goes/internal/shellutils"
)

type Command struct {
	M *menuentry.Command
}

func (c Command) String() string { return "submenu" }

func (c Command) Usage() string {
	return "NOP"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "NOP",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: Man,
	}
}

const Man = "NOP command for script compatibility\n"

func (c Command) Block(g *goes.Goes, ls shellutils.List) (*shellutils.List, func(stdin io.Reader, stdout io.Writer, stderr io.Writer, isFirst bool, isLast bool) error, error) {
	cl := ls.Cmds[0]
	// submenu name { definition ... ; }
	if len(cl.Cmds) < 2 {
		return nil, nil, errors.New("Submenu: unexpected end of line")
	}

	foundBrace := false
	cl.Cmds = cl.Cmds[1:]
	args := make([]string, 0)

	for len(cl.Cmds) > 0 && !foundBrace {
		cmd := cl.Cmds[0].String()
		cl.Cmds = cl.Cmds[1:]
		if strings.HasSuffix(cmd, "{") {
			foundBrace = true
			cmd = cmd[:len(cmd)-1]
			if len(cmd) == 0 {
				break
			}
		}
		args = append(args, cmd)
	}

	if !foundBrace {
		return nil, nil, errors.New("submenu: missing {")
	}

	_, args = parms.New(args, "--class", "--users", "--hotkey", "--id")
	_, args = flags.New(args, "--unrestricted")

	if len(args) < 1 {
		return nil, nil, errors.New("submenu: missing menu name")
	}

	name := args[0]

	if len(cl.Cmds) > 0 {
		ls.Cmds[0] = cl
	} else {
		ls.Cmds = ls.Cmds[1:]
	}

	var funList []func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error
	for {
		for len(ls.Cmds) == 0 {
			newls, err := shellutils.Parse("submenu>", g.Catline)
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
		if name != "menuentry" {
			return nil, nil, fmt.Errorf("submenu: unexpected %s", name)
		}
		nextls, _, runfun, err := g.ProcessList(ls)
		if err != nil {
			return nil, nil, err
		}
		funList = append(funList, runfun)
		ls = *nextls
	}

	e := menuentry.Entry{Name: name}
	e.RunFun = func(stdin io.Reader, stdout io.Writer, stderr io.Writer, isFirst bool, isLast bool) error {
		for _, runent := range funList {
			err := runent(stdin, stdout, stderr)
			if err != nil {
				return err
			}
		}
		return nil
	}

	deffun := func(stdin io.Reader, stdout io.Writer, stderr io.Writer, isFirst bool, isLast bool) error {
		c.M.Menus = append(c.M.Menus, e)
		return nil
	}
	return &ls, deffun, nil
}

func (Command) Main(args ...string) error {
	return errors.New("internal error")
}
