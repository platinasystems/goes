// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package menu

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/external/parms"
	"github.com/platinasystems/goes/internal/shellutils"
	"github.com/platinasystems/goes/lang"
)

var InternalError = errors.New("internal error")
var ErrMenuNotFound = errors.New("Menu not found")
var ErrMenuUnexpectedEOL = errors.New("Unexpected end of line")
var ErrMenuMissingOpenBrace = errors.New("Missing {")
var ErrMenuMissingName = errors.New("Missing menu name")
var ErrMenuUnexpectedText = errors.New("Unexpected text after }")
var ErrMenuNoNesting = errors.New("Nesting not allowed")
var ErrMenuOutOfRange = errors.New("Menu item out of range")

type Root struct {
	RootMenu    *Menu
	CurrentMenu *Menu
}

type Menu struct {
	Entries *[]Entry // Menu entries at this level
}

type Entry struct {
	Name    string
	RunFun  func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error
	Submenu *Menu
}

type Command struct {
	C       string
	R       *Root
	Nesting bool
}

func (r *Root) String() (s string) {
	return r.RootMenu.String()
}

func (m *Menu) String() (s string) {
	for _, e := range *m.Entries {
		s += e.String()
	}
	return
}

func (e *Entry) String() (s string) {
	return e.dumpEntry(0)
}

func (e *Entry) dumpEntry(d int) (s string) {
	s = strings.Repeat(" ", d) +
		fmt.Sprintf("Name: %v RunFun: %v\n", e.Name, &e.RunFun)
	sm := e.Submenu
	if sm != nil {
		for _, en := range *sm.Entries {
			s += en.dumpEntry(d + 1)
		}
	}
	return
}

// NumberedMenu() returns the indicated menu as a numbered list.
func (m *Menu) NumberedMenu() (s string) {
	for i, e := range *m.Entries {
		s += fmt.Sprintf("[%d]  %s\n", i, e.Name)
	}
	return
}

func (c *Command) String() string { return c.C }

func (c *Command) Usage() string {
	return c.C + " [options] [name] { script ... }"
}

func (c *Command) Apropos() lang.Alt {
	nestUs := ""
	if c.Nesting {
		nestUs = " that may contain other menus"
	}
	return lang.Alt{
		lang.EnUS: "define a menu item" + nestUs,
	}
}

func (c *Command) Man() lang.Alt {
	nestUs := ""
	diffUs := ""
	if c.Nesting {
		nestUs = " that may define other menus"
		diffUs = `

	The difference between this and the menuentry command is that
	` + c.String() + ` can contain other submenu and menuentry commands, and
	menuentry can not.`
	}
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Define a menu item` + nestUs + `.

	Options and names are currently ignored. They do not return
	errors for compatibility with existing grub scripts.

	The menu itself is a script which will be run when the menu
	item is selected.` + diffUs,
	}
}

func (c *Command) Block(g *goes.Goes, ls shellutils.List) (*shellutils.List,
	func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error, error) {
	cl := ls.Cmds[0]
	// menuentry name [option...] { definition ... ; }
	if len(cl.Cmds) < 2 {
		return nil, nil, fmt.Errorf("%s: %w", c, ErrMenuUnexpectedEOL)
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
		return nil, nil, fmt.Errorf("%s: %w", c, ErrMenuMissingOpenBrace)
	}

	_, args = parms.New(args, "--class", "--users", "--hotkey", "--id")
	_, args = flags.New(args, "--unrestricted")

	if len(args) < 1 {
		return nil, nil, fmt.Errorf("%s: %w", c, ErrMenuMissingName)
	}

	name := args[0]
	//	fmt.Printf("menuentry: name: %v, parm: %v, flags: %v, args: %v\n",
	//		name, parm, flags, args)

	if len(cl.Cmds) > 0 {
		ls.Cmds[0] = cl
	} else {
		ls.Cmds = ls.Cmds[1:]
	}

	var funList []func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error
	for {
		for len(ls.Cmds) == 0 {
			newls, err := shellutils.Parse(c.String()+">",
				g.Catline)
			if err != nil {
				return nil, nil, err
			}
			ls = *newls
		}
		cl := ls.Cmds[0]
		name := cl.Cmds[0].String()
		if name == "}" {
			if len(cl.Cmds) > 1 {
				return nil, nil, fmt.Errorf("%s: %w", c,
					ErrMenuUnexpectedText)
			}
			break
		}
		if !c.Nesting && (name == "menuentry" || name == "submenu") {
			return nil, nil, fmt.Errorf("%s: %w place %s in submenu",
				c, ErrMenuNoNesting, name)
		}
		nextls, _, runfun, err := g.ProcessList(ls)
		if err != nil {
			return nil, nil, err
		}
		funList = append(funList, runfun)
		ls = *nextls
	}

	sm := &Menu{Entries: &[]Entry{}}
	e := Entry{Name: name}
	if c.Nesting {
		e.Submenu = sm
	}
	e.RunFun = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
		if c.Nesting {
			saved := c.R.CurrentMenu
			c.R.CurrentMenu = sm
			defer func() {
				c.R.CurrentMenu = saved
			}()
		}
		for _, runent := range funList {
			err := runent(stdin, stdout, stderr)
			if err != nil {
				return err
			}
		}
		return nil
	}

	deffun := func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
		m := c.R.CurrentMenu
		*m.Entries = append(*m.Entries, e)

		return nil
	}
	return &ls, deffun, nil
}

// New() - Create a menuentry and submenu command sharing the same
// menu tree.
func New() (*Command, *Command) {
	m := &Menu{Entries: &[]Entry{}}
	r := &Root{RootMenu: m, CurrentMenu: m}
	return &Command{C: "menuentry", R: r},
		&Command{C: "submenu", R: r, Nesting: true}
}

// Reset() resets a menu to be empty
func (c *Command) Reset() {
	m := &Menu{Entries: &[]Entry{}}
	c.R.RootMenu = m
	c.R.CurrentMenu = m
}

func (Command) Main(args ...string) error {
	return InternalError
}

func (c *Command) FindMenu(args ...int) (m *Menu, err error) {
	m = c.R.RootMenu
	for len(args) > 0 {
		i := args[0]
		if m == nil || i >= len(*m.Entries) {
			return nil, ErrMenuNotFound
		}
		m = (*m.Entries)[i].Submenu
		args = args[1:]
	}
	if m == nil {
		return nil, ErrMenuNotFound
	}
	return m, nil
}

func (c *Command) FindEntry(args ...int) (e *Entry, err error) {
	l := len(args)
	mp := args[:l-1]
	m, err := c.FindMenu(mp...)
	if err != nil {
		return nil, err
	}
	i := args[l-1]
	if i >= len(*m.Entries) {
		return nil, ErrMenuNotFound
	}
	e = &(*m.Entries)[i]
	return
}

func (m *Menu) RunMenu(i int,
	stdin io.Reader,
	stdout io.Writer, stderr io.Writer) (sm *Menu, err error) {
	if i >= len((*m.Entries)) {
		return nil, ErrMenuOutOfRange
	}
	me := (*m.Entries)[i]
	return me.Submenu, me.RunFun(stdin, stdout, stderr)
}
