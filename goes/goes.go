// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build linux

// Package goes, combined with a compatibly configured Linux kernel, provides a
// monolithic embedded system.
package goes

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/prog"
)

var blockNames = [...]string{
	"BlockIf",
	"BlockIfNotTaken",
	"BlockIfThenNotTaken",
	"BlockIfThenTaken",
	"BlockIfElseNotTaken",
	"BlockIfElseTaken"}

type block int

const (
	BlockIf = iota
	BlockIfNotTaken
	BlockIfThenNotTaken
	BlockIfThenTaken
	BlockIfElseNotTaken
	BlockIfElseTaken
)

func (b block) String() string {
	return blockNames[b]
}

type akaer interface {
	Aka() string
}

type goeser interface {
	Goes(*Goes)
}

type Goes struct {
	// These uppercased fields may/should be assigned at instantiation
	NAME, USAGE  string
	APROPOS, MAN lang.Alt

	ByName map[string]cmd.Cmd

	Catline func(string) (string, error)

	Blocks []block

	Status error
	debug  bool

	cache  cache
	parent *Goes

	EnvMap map[string]string
}

func (g *Goes) NotTaken() bool {
	if len(g.Blocks) != 0 &&
		(g.Blocks[len(g.Blocks)-1] == BlockIfNotTaken ||
			g.Blocks[len(g.Blocks)-1] == BlockIfThenNotTaken ||
			g.Blocks[len(g.Blocks)-1] == BlockIfElseNotTaken) {
		return true
	}
	return false
}

func Replace(s, name string) string {
	return strings.Replace(s, "goes", name, -1)
}

func (g *Goes) String() string {
	name := g.NAME
	if len(name) == 0 {
		name = "goes"
	}
	return name
}

func (g *Goes) Goes(parent *Goes) {
	g.parent = parent
}

// Fork returns an exec.Cmd ready to Run or Output this program with the
// given args.
func (g *Goes) Fork(args ...string) *exec.Cmd {
	if g.debug {
		fmt.Printf("F*$=%v %v %v\n", g.Status, g.Blocks, args)
	}
	if g.NotTaken() {
		return nil
	}
	a := append(g.Path(), args...)
	x := exec.Command(prog.Name(), a[1:]...)
	x.Args[0] = a[0]
	return x
}

// Run a command in the current context.
//
// If len(args) == 1 and args[0] doesn't match a mapped command, this will run
// the "cli".
//
// If the args has "-help", or "--help", this runs ByName("help").Main(args...)
// to print text.
//
// Similarly for "-apropos", "-complete", "-man", and "-usage".
//
// If the command is a daemon, this fork exec's itself twice to disassociate
// the daemon from the tty and initiating process.
func (g *Goes) Main(args ...string) error {
	if len(args) > 0 {
		base := filepath.Base(args[0])
		switch {
		case g.NAME == "goes-installer":
			if len(args) == 1 {
				args[0] = "install"
			} else {
				args = args[1:]
			}
		case base == g.NAME:
			// e.g. ./goes-MACHINE ...
			fallthrough
		case base == "goes":
			args = args[1:]
		}
	}

	cli, found := g.ByName["cli"]
	if found {
		cli.(goeser).Goes(g)
	}
	cliFlags, cliArgs := flags.New(args, "-d", "-f", "-no-liner", "-x")
	if cliFlags.ByName["-d"] {
		g.debug = true
	}
	if n := len(cliArgs); n == 0 {
		if cli != nil {
			if cliFlags.ByName["-no-liner"] {
				cliArgs = append(cliArgs, "-no-liner")
			}
			if cliFlags.ByName["-x"] {
				cliArgs = append(cliArgs, "-x")
			}
			g.Status = cli.Main(cliArgs...)
			return g.Status
		} else if def, found := g.ByName[""]; found {
			g.Status = def.Main()
			return g.Status
		}
		fmt.Println(Usage(g))
		g.Status = nil
		return nil
	} else if _, found := g.ByName[args[0]]; n == 1 && !found {
		// only check for script if args[0] isn't a command
		buf, err := ioutil.ReadFile(cliArgs[0])
		if cliArgs[0] == "-" || (err == nil && utf8.Valid(buf) &&
			bytes.HasPrefix(buf, []byte("#!/usr/bin/goes"))) {
			// e.g. /usr/bin/goes SCRIPT
			if cli == nil {
				g.Status = fmt.Errorf("has no cli")
				return g.Status
			}
			for _, t := range []string{"-f", "-x"} {
				if cliFlags.ByName[t] {
					cliArgs = append(cliArgs, t)
				}
			}
			g.Status = cli.Main(cliArgs...)
			return g.Status
		}
	} else {
		g.swap(args)
	}

	if builtin, found := g.Builtins()[args[0]]; found {
		g.Status = builtin(args[1:]...)
		return g.Status
	} else if len(args) == 1 && strings.HasPrefix(args[0], "-") {
		arg0 := strings.TrimLeft(args[0], "-")
		if arg0 == "apropos" {
			fmt.Println(g.Apropos())
			return nil
		} else if builtin, found := g.Builtins()[arg0]; found {
			g.Status = builtin()
			return g.Status
		}
	}

	g.shift(args)

	if g.debug {
		fmt.Printf("$=%v %v %v\n", g.Status, g.Blocks, args)
	}

	v, found := g.ByName[args[0]]
	if !found {
		if v, found = g.ByName[""]; !found {
			g.Status =
				fmt.Errorf("%s: ambiguous or missing command",
					args[0])
			return g.Status
		}
		// e.g. ip -s add [default "show"]
		args = append([]string{""}, args...)
	} else if method, found := v.(goeser); found {
		method.Goes(g)
	}

	k := cmd.WhatKind(v)
	if !k.IsConditional() && g.NotTaken() {
		return nil
	}

	if k.IsDaemon() {
		sig := make(chan os.Signal)
		signal.Notify(sig, syscall.SIGTERM)
		defer func(sig chan os.Signal) {
			sig <- syscall.SIGABRT
		}(sig)
		go wait(v, sig)
	}

	err := v.Main(args[1:]...)
	if err == io.EOF {
		err = nil
	}
	if err != nil && !k.IsDaemon() {
		name := args[0]
		if len(name) == 0 {
			if method, found := v.(akaer); found {
				name = fmt.Sprint("(", method.Aka(), ")")
			}
		}
		err = fmt.Errorf("%s: %v", name, err)
	}
	g.Status = err
	return err
}

// shift the first unambiguous longest prefix match command to args[0], so,
//
//	OPTIONS... COMMAND [ARGS]...
//
// becomes
//
//	COMMAND OPTIONS... [ARGS]...
//
// e.g.
//
//	ip -s li
//
// becomes
//
//	ip link -s
func (g *Goes) shift(args []string) {
	for i := range args {
		if _, found := g.ByName[args[i]]; found {
			if i > 0 {
				name := args[i]
				copy(args[1:i+1], args[:i])
				args[0] = name
			}
			return
		}
		var matches int
		var last string
		for _, name := range g.Names() {
			if strings.HasPrefix(name, args[i]) {
				last = name
				matches++
			}
		}
		if matches == 1 {
			if i > 0 {
				copy(args[1:i+1], args[:i])
			}
			args[0] = last
			return
		}
	}
}

// swap hyphen prefaced helper flags with command, so,
//
//	COMMAND [-[-]]HELPER [ARGS]...
//
// becomes
//
//	HELPER COMMAND [ARGS]...
//
// and
//
//	-[-]HELPER [ARGS]...
//
// becomes
//
//	HELPER [ARGS]...
func (g *Goes) swap(args []string) {
	n := len(args)
	if n > 0 && strings.HasPrefix(args[0], "-") {
		opt := strings.TrimLeft(args[0], "-")
		if _, found := g.Builtins()[opt]; found {
			args[0] = opt
		}
	} else if n > 1 {
		opt := strings.TrimLeft(args[1], "-")
		if _, found := g.Builtins()[opt]; found {
			args[1] = args[0]
			args[0] = opt
		}
	}
}

func wait(v cmd.Cmd, ch chan os.Signal) {
	for sig := range ch {
		if sig == syscall.SIGTERM {
			if method, found := v.(io.Closer); found {
				if err := method.Close(); err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			}
			fmt.Println("killed")
			os.Stdout.Sync()
			os.Stderr.Sync()
			os.Stdout.Close()
			os.Stderr.Close()
			os.Exit(0)
		}
		break
	}
}
