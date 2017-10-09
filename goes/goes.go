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
	"sort"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/prog"
)

const (
	USAGE = `
	goes COMMAND [ ARGS ]...
	goes COMMAND -[-]HELPER [ ARGS ]...
	goes HELPER [ COMMAND ] [ ARGS ]...
	goes [ -d ] [ -x ] [[ -f ][ - | SCRIPT ]]

	HELPER := { apropos | complete | help | man | usage }
	`
	MAN = `
OPTIONS
	-d	debug block handling
	-x	print command trace
	-f	don't terminate script on error
	-	execute standard input script
	SCRIPT	execute named script file

SEE ALSO
	goes apropos [COMMAND], goes man COMMAND`
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

type Goes struct {
	name, usage  string
	apropos, man lang.Alt

	byname map[string]cmd.Cmd
	Names  []string

	Parent *Goes
	Path   []string

	Catline func(string) (string, error)

	Blocks []block

	Status error
	debug  bool
}

type akaer interface {
	Aka() string
}

type completer interface {
	Complete(...string) []string
}

type goeser interface {
	Goes(*Goes)
}

type helper interface {
	Help(...string) string
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

func New(name, usage string, apropos, man lang.Alt) *Goes {
	if len(usage) == 0 {
		usage = strings.Replace(USAGE, "goes", name, -1)
	}
	if len(man) == 0 {
		man = lang.Alt{
			lang.EnUS: strings.Replace(MAN, "goes", name, -1),
		}
	}
	return &Goes{
		name:    name,
		usage:   usage,
		apropos: apropos,
		man:     man,
		byname:  make(map[string]cmd.Cmd),
	}
}

func (g *Goes) Apropos() lang.Alt { return g.apropos }

func (g *Goes) ByName(name string) cmd.Cmd { return g.byname[name] }

func (g *Goes) Complete(args ...string) (ss []string) {
	n := len(args)
	if n == 0 || len(args[0]) == 0 {
		ss = g.Names
	} else if v, found := g.byname[args[0]]; found {
		if method, found := v.(completer); found {
			ss = method.Complete(args[1:]...)
		} else {
			ss, _ = filepath.Glob(args[n-1] + "*")
		}
	} else if _, found := cmd.Helpers[args[0]]; found {
		if n == 1 || len(args[n-1]) == 0 {
			return g.Names
		}
		for _, name := range g.Names {
			if strings.HasPrefix(name, args[n-1]) {
				ss = append(ss, name)
			}
		}
		if len(ss) > 0 {
			sort.Strings(ss)
		}
	} else if n == 1 {
		for _, name := range g.Names {
			if strings.HasPrefix(name, args[0]) {
				ss = append(ss, name)
			}
		}
		for helper := range cmd.Helpers {
			if strings.HasPrefix(helper, args[0]) {
				ss = append(ss, helper)
			}
		}
		if len(ss) > 0 {
			sort.Strings(ss)
		}
	}
	return
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

	if g.Parent != nil && len(g.Path) == 0 {
		// set Path of sub-goes. e.g. "ip address"
		for p := g; p != nil; p = p.Parent {
			g.Path = append([]string{p.String()}, g.Path...)
		}
	}
	a := append(g.Path, args...)
	x := exec.Command(prog.Name(), a[1:]...)
	x.Args[0] = a[0]
	return x
}

func (g *Goes) Help(args ...string) string {
	cmd.Swap(args)
	g.Shift(args)
	if len(args) > 0 {
		if v, found := g.byname[args[0]]; found {
			if method, found := v.(helper); found {
				return method.Help(args[1:]...)
			}
			return Usage(v)
		}
	}
	return Usage(g)
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
	if g.Parent != nil && len(g.Path) == 0 {
		// set Path of sub-goes. e.g. "ip address"
		for p := g; p != nil; p = p.Parent {
			g.Path = append([]string{p.String()}, g.Path...)
		}
	}
	if len(args) > 0 {
		base := filepath.Base(args[0])
		switch {
		case g.name == "goes-installer":
			if len(args) == 1 {
				args[0] = "install"
			} else {
				args = args[1:]
			}
		case base == g.name:
			// e.g. ./goes-MACHINE ...
			fallthrough
		case base == "goes":
			args = args[1:]
		}
	}

	cli := g.byname["cli"]
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
			err := cli.Main(cliArgs...)
			g.Status = err
			return err
		} else if def, found := g.byname[""]; found {
			err := def.Main()
			g.Status = err
			return err
		}
		fmt.Println(Usage(g))
		g.Status = nil
		return nil
	} else if _, found := g.byname[args[0]]; n == 1 && !found {
		// only check for script if args[0] isn't a command
		buf, err := ioutil.ReadFile(cliArgs[0])
		if cliArgs[0] == "-" || (err == nil && utf8.Valid(buf) &&
			bytes.HasPrefix(buf, []byte("#!/usr/bin/goes"))) {
			// e.g. /usr/bin/goes SCRIPT
			if cli == nil {
				err := fmt.Errorf("has no cli")
				g.Status = err
				return err
			}
			for _, t := range []string{"-f", "-x"} {
				if cliFlags.ByName[t] {
					cliArgs = append(cliArgs, t)
				}
			}
			err := cli.Main(cliArgs...)
			g.Status = err
			return err
		}
	} else {
		cmd.Swap(args)
	}

	if _, found := cmd.Helpers[args[0]]; found {
		err := g.byname[args[0]].Main(args[1:]...)
		g.Status = err
		return err
	}

	g.Shift(args)

	if g.debug {
		fmt.Printf("$=%v %v %v\n", g.Status, g.Blocks, args)
	}

	v, found := g.byname[args[0]]
	if !found {
		if v, found = g.byname[""]; !found {
			err := fmt.Errorf("%s: ambiguous or missing command",
				args[0])
			g.Status = err
			return err
		}
		// e.g. ip -s add [default "show"]
		args = append([]string{""}, args...)
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

func (g *Goes) Man() lang.Alt { return g.man }

// Plot command by name.
func (g *Goes) Plot(cmds ...cmd.Cmd) {
	for _, v := range cmds {
		name := v.String()
		if _, found := g.byname[name]; found {
			panic(fmt.Errorf("%s: duplicate", name))
		}
		g.byname[name] = v
		if _, found := cmd.Helpers[name]; !found {
			g.Names = append(g.Names, name)
		}
		if method, found := v.(goeser); found {
			method.Goes(g)
		}
		if vg, found := v.(*Goes); found {
			vg.Parent = g
		}
	}
	sort.Strings(g.Names)
}

// Shift the first unambiguous longest prefix match command to args[0], so,
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
func (g *Goes) Shift(args []string) {
	for i := range args {
		if _, found := g.byname[args[i]]; found {
			if i > 0 {
				name := args[i]
				copy(args[1:i+1], args[:i])
				args[0] = name
			}
			return
		}
		var matches int
		var last string
		for _, name := range g.Names {
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

func (g *Goes) String() string { return g.name }
func (g *Goes) Usage() string  { return g.usage }

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

func Usage(v Usager) string {
	return fmt.Sprint("usage:\t", strings.TrimSpace(v.Usage()))
}

type Usager interface {
	Usage() string
}
