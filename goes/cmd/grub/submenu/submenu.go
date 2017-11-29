// Copyright © 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package submenu

import (
	"errors"
	"fmt"
	"io"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"

	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/go/internal/shellutils"
)

type Command struct{}

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

func (Command) Block(g *goes.Goes, ls shellutils.List) (*shellutils.List, func(stdin io.Reader, stdout io.Writer, stderr io.Writer, isFirst bool, isLast bool) error, error) {
	cl := ls.Cmds[0]
	// submenu name { definition ... ; }
	if len(cl.Cmds) < 2 {
		return nil, nil, errors.New("Submenu: unexpected end of line")
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
		return nil, nil, errors.New("submenu: missing {")
	}

	_, args = parms.New(args, "--class", "--users", "--hotkey", "--id")
	_, args = flags.New(args, "--unrestricted")

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

	runfun := func(stdin io.Reader, stdout io.Writer, stderr io.Writer, isFirst bool, isLast bool) error {
		for _, runent := range funList {
			err := runent(stdin, stdout, stderr)
			if err != nil {
				return err
			}
		}
		return nil
	}
	fmt.Printf("Submenu %s defined\n", name)
	return &ls, runfun, nil
}

func (Command) Main(args ...string) error {
	return errors.New("internal error")
}
