// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package function

import (
	"errors"
	"fmt"
	"io"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/lang"

	"github.com/platinasystems/goes/internal/shellutils"
)

const (
	Name    = "function"
	Apropos = "function name { definition }"
	Usage   = "function name { definition }"
	Man     = `
DESCRIPTION
	Define a function.
`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Block(g *goes.Goes, ls shellutils.List) (*shellutils.List, func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error, error) {
	cl := ls.Cmds[0]
	// function name { definition ... ; }
	if len(cl.Cmds) < 2 {
		return nil, nil, errors.New("Function: unexpected end of line")
	}

	name := cl.Cmds[1].String()
	cl.Cmds = cl.Cmds[2:]
	for len(cl.Cmds) < 1 {
		ls.Cmds = ls.Cmds[1:]
		for len(ls.Cmds) == 0 {
			newls, err := shellutils.Parse("function>", g.Catline)
			if err != nil {
				return nil, nil, err
			}
			ls = *newls
		}
		cl = ls.Cmds[0]
	}
	if cl.Cmds[0].String() != "{" {
		return nil, nil, fmt.Errorf("Function: unexpected %s",
			cl.Cmds[0].String())
	}
	if len(cl.Cmds) > 1 {
		cl.Cmds = cl.Cmds[1:]
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
			newls, err := shellutils.Parse("function>", g.Catline)
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

	runfun := func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
		for _, runent := range funList {
			err := runent(stdin, stdout, stderr)
			if err != nil {
				return err
			}
		}
		return nil
	}
	f := goes.Function{Name: name, RunFun: runfun}

	deffun := func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
		if g.FunctionMap == nil {
			g.FunctionMap = make(map[string]goes.Function, 0)
		}
		g.FunctionMap[name] = f
		return nil
	}
	return &ls, deffun, nil
}

func (Command) Main(args ...string) error {
	return errors.New("internal error")
}
