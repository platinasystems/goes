// Copyright © 2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package menu

import (
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/platinasystems/goes/internal/shellutils"
)

func TestMain(t *testing.T) {
	m, s := New()
	testCmdWithArgs(m, t, InternalError)
	testCmdWithArgs(s, t, InternalError)
}

func ExampleMain() {
	m, _ := New()
	fmt.Println(m.String())
	fmt.Println(m.Usage())
	fmt.Println(m.Apropos())
	fmt.Println(m.Man())
	// Output:
	// menuentry
	// menuentry [options] [name] { script ... }
	// define a menu item
	//
	// DESCRIPTION
	//	Define a menu item.
	//
	//	Options and names are currently ignored. They do not return
	//	errors for compatibility with existing grub scripts.
	//
	//	The menu itself is a script which will be run when the menu
	//	item is selected.
}

func ExampleMain_submenuCommand() {
	_, s := New()
	fmt.Println(s.String())
	fmt.Println(s.Usage())
	fmt.Println(s.Apropos())
	fmt.Println(s.Man())
	// Output:
	// submenu
	// submenu [options] [name] { script ... }
	// define a menu item that may contain other menus
	//
	// DESCRIPTION
	//	Define a menu item that may define other menus.
	//
	//	Options and names are currently ignored. They do not return
	//	errors for compatibility with existing grub scripts.
	//
	//	The menu itself is a script which will be run when the menu
	//	item is selected.
	//
	//	The difference between this and the menuentry command is that
	//	submenu can contain other submenu and menuentry commands, and
	//	menuentry can not.
}

func TestScript(t *testing.T) {
	m, _, g := newgoes(&ts{
		script: []string{
			"menuentry menu0 {",
			"    testpoint 0",
			"}",
			"menuentry menu1 {",
			"    testpoint 1",
			"}",
			"submenu menu2 {",
			"    testpoint 2",
			"    menuentry menu21 {",
			"        testpoint 21",
			"    }",
			"}",
		},
	})

	err := g.Main("test", "cli", "-")
	if err != nil {
		t.Errorf("Main returned %s", err)
	}

	fmt.Printf("Menu tree:\n%s", m.R)

	_, err = m.FindEntry(1, 2, 3)
	if err != ErrMenuNotFound {
		if err == nil {
			t.Errorf("FindEntry(1, 2, 3) returned success expecting %s",
				ErrMenuNotFound)
		}
		t.Errorf("FindEntry(1, 2, 3) returned %s expecting %s",
			err, ErrMenuNotFound)
	}

	e0, err := m.FindEntry(0)
	if err != nil {
		t.Errorf("FindEntry(0) returned %s", err)
	}
	if e0.Submenu != nil {
		t.Errorf("FindEntry(0) shouldn't have submenus, but got %s",
			e0)
	}
	fmt.Printf("E0: %s", e0)
	e1, err := m.FindEntry(1)
	if err != nil {
		t.Errorf("FindEntry(1) returned %s", err)
	}
	if e1.Submenu != nil {
		t.Errorf("FindEntry(1) shouldn't have submenus, but got %s",
			e1)
	}
	fmt.Printf("E1: %s", e1)
	e2, err := m.FindEntry(2)
	if err != nil {
		t.Errorf("FindEntry(2) returned %s", err)
	}
	fmt.Printf("E2: %s\n", e2)
	if e2.Submenu == nil {
		t.Errorf("FindEntry(2) didn't return submenu")
	}
	_, err = m.FindEntry(2, 0)
	if err != ErrMenuNotFound {
		if err == nil {
			t.Errorf("FindEntry(2,0) returned success expecting %s",
				ErrMenuNotFound)
		}
		t.Errorf("FindEntry(2,0) returned %s expecting %s",
			err, ErrMenuNotFound)
	}
	err = e2.RunFun(os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		t.Errorf("FindEntry(2).RunFun returned %s", err)
	}
	fmt.Printf("E2 (after running menu): %s", e2)
	e21, err := m.FindEntry(2, 0)
	if err != nil {
		t.Errorf("FindEntry(2,0) returned %s", err)
	}
	fmt.Printf("E21: %s", e21)
	fmt.Println(e2.Submenu.NumberedMenu())
	_, err = m.FindEntry(2, 0, 0)
	if err != ErrMenuNotFound {
		if err == nil {
			t.Errorf("FindEntry(2,0,0) returned success expecting %s",
				ErrMenuNotFound)
		}
		t.Errorf("FindEntry(2,0,0) returned %s expecting %s",
			err, ErrMenuNotFound)
	}
	m.Reset()
	_, err = m.FindEntry(0)
	if err != ErrMenuNotFound {
		if err == nil {
			t.Errorf("After Reset(), FindEntry(0) returned success expecting %s",
				ErrMenuNotFound)
		}
		t.Errorf("After Reset(), FindEntry(0) returned %s expecting %s",
			err, ErrMenuNotFound)
	}
}

func TestScriptIncomplete(t *testing.T) {
	_, _, g := newgoes(&ts{
		script: []string{
			"menuentry menu0 {",
			"testpoint",
		},
	})
	err := g.Main("test", "cli", "-")
	if !errors.Is(err, io.EOF) {
		if err == nil {
			t.Errorf("Incomplete script returned no error")
		}
		t.Errorf("Incomplete script returned error %s", err)
	}
}

func TestOneline(t *testing.T) {
	_, _, g := newgoes(&ts{
		script: []string{
			"menuentry menu0 { testpoint; }",
		},
	})
	err := g.Main("test", "cli", "-")
	if err != nil {
		t.Errorf("Oneline script returned %s", err)
	}
}

func TestErrorInMenu(t *testing.T) {
	m, _, g := newgoes(&ts{
		script: []string{
			"menuentry menu0 {",
			"testpoint fail",
			"}",
		},
	})
	err := g.Main("test", "cli", "-")
	if err != nil {
		t.Errorf("TestErrorInMenu script returned %s", err)
	}
	menu := m.R.RootMenu
	if m == nil {
		t.Errorf("Can't find root menu")
	}
	menu0, err := menu.RunMenu(0, os.Stdin, os.Stdout, os.Stderr)
	if !errors.Is(err, ErrTestpointFailed) {
		if err == nil {
			t.Errorf("RunMenu(0) returned success")
		} else {
			t.Errorf("RunMenu(0) returned %s", err)
		}
	}
	if menu0 != nil {
		t.Errorf("RunMenu(0) returned submenus!")
	}
	_, err = menu.RunMenu(1, os.Stdin, os.Stdout, os.Stderr)
	if !errors.Is(err, ErrMenuOutOfRange) {
		if err == nil {
			t.Errorf("RunMenu(0) returned success")
		} else {
			t.Errorf("RunMenu(0) returned %s", err)
		}
	}

}

func TestUnterminatedString(t *testing.T) {
	_, _, g := newgoes(&ts{
		script: []string{
			"submenu foo {",
			"menuentry bar {",
			`testpoint "unterminated quoted string`,
		},
	})
	err := g.Main("test", "cli", "-")
	if !errors.Is(err, shellutils.ErrMissingEndQuote) {
		if err == nil {
			t.Errorf("Got success expecting %s",
				shellutils.ErrMissingEndQuote)
		} else {
			t.Errorf("Got %s expecting %s",
				err, shellutils.ErrMissingEndQuote)
		}
	}
}

func TestMenuentryNoArgs(t *testing.T) {
	s1 := &ts{
		script: []string{
			"menuentry",
		},
	}
	testScriptNoArgs(s1, t)
}

func TestSubmenuNoArgs(t *testing.T) {
	s1 := &ts{
		script: []string{
			"submenu",
		},
	}
	testScriptNoArgs(s1, t)
}

func TestMenuentryNoOpenBrace(t *testing.T) {
	s1 := &ts{
		script: []string{
			"menuentry foo",
		},
	}
	testScriptNoOpenBrace(s1, t)
}

func TestSubmenuNoOpenBrace(t *testing.T) {
	s1 := &ts{
		script: []string{
			"submenu foo",
		},
	}
	testScriptNoOpenBrace(s1, t)
}

func TestMenuentryMissingMenuName(t *testing.T) {
	s1 := &ts{
		script: []string{
			"menuentry {",
		},
	}
	testScriptMissingMenuName(s1, t)
}

func TestSubmenuMissingMenuName(t *testing.T) {
	s1 := &ts{
		script: []string{
			"submenu {",
		},
	}
	testScriptMissingMenuName(s1, t)
}

func TestMenuentryUnexpectedText(t *testing.T) {
	s1 := &ts{
		script: []string{
			"menuentry test {",
			"testpoint",
			"} foo",
		},
	}
	testScriptUnexpectedText(s1, t)
}

func TestSubmenuUnexpectedText(t *testing.T) {
	s1 := &ts{
		script: []string{
			"submenu test {",
			"testpoint",
			"} foo",
		},
	}
	testScriptUnexpectedText(s1, t)
}

func TestMenuentryNestedMenus(t *testing.T) {
	s1 := &ts{
		script: []string{
			"menuentry test {",
			"menuentry test1 {",
			"testpoint",
			"}",
			"}",
		},
	}
	testScriptNestedMenus(s1, t, false)
}

func TestSubmenuNestedMenus(t *testing.T) {
	s1 := &ts{
		script: []string{
			"submenu test {",
			"menuentry test1 {",
			"testpoint",
			"}",
			"}",
		},
	}
	testScriptNestedMenus(s1, t, true)
}
