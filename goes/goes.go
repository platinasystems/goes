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
	goes [ -x ] [[ -f ][ - | SCRIPT ]]

	HELPER := { apropos | complete | help | man | usage }
	`
	MAN = `
OPTIONS
	-x	print command trace
	-f	don't terminate script on error
	-	execute standard input script
	SCRIPT	execute named script file

SEE ALSO
	goes apropos [COMMAND], goes man COMMAND`
)

type Goes struct {
	name, usage  string
	apropos, man lang.Alt

	byname map[string]cmd.Cmd
	Names  []string

	Parent *Goes
	Path   []string
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

func (g *Goes) Complete(args ...string) []string {
	var ss []string
	pat := "*"

	if len(args) == 0 {
		return g.Names
	}

	cmd.Swap(args)

	if _, found := cmd.Helpers[args[0]]; found {
		n := len(args)
		if n == 1 {
			return g.Names
		}
		for helper := range cmd.Helpers {
			if strings.HasPrefix(helper, args[n-1]) {
				ss = append(ss, helper)
			}
		}
		for _, name := range g.Names {
			if strings.HasPrefix(name, args[n-1]) {
				ss = append(ss, name)
			}
		}
		sort.Strings(ss)
		return ss
	}

	g.Shift(args)

	n := len(args)

	if n == 0 {
		return g.Names
	}

	if v, found := g.byname[args[0]]; found {
		if method, found := v.(completer); found {
			return method.Complete(args[1:]...)
		}
		if n > 1 {
			pat = args[n-1] + pat
		}

		ss, _ = filepath.Glob(pat)
	} else if n == 1 {
		for helper := range cmd.Helpers {
			if strings.HasPrefix(helper, args[0]) {
				ss = append(ss, helper)
			}
		}
		for _, name := range g.Names {
			if strings.HasPrefix(name, args[0]) {
				ss = append(ss, name)
			}
		}
	} else {
		ss, _ = filepath.Glob(args[n-1] + pat)
	}
	return ss
}

// Fork returns an exec.Cmd ready to Run or Output this program with the
// given args.
func (g *Goes) Fork(args ...string) *exec.Cmd {
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
		case strings.HasPrefix(base, "goes-") &&
			strings.HasSuffix(base, "-installer"):
			// e.g. ./goes-MACHINE-installer
			args[0] = "install"
		case base == g.name:
			// e.g. ./goes-MACHINE ...
			fallthrough
		case base == "goes":
			args = args[1:]
		}
	}

	cli := g.byname["cli"]
	cliFlags, cliArgs := flags.New(args, "-f", "-no-liner", "-x")
	if n := len(cliArgs); n == 0 {
		if cli != nil {
			if cliFlags.ByName["-no-liner"] {
				cliArgs = append(cliArgs, "-no-liner")
			}
			if cliFlags.ByName["-x"] {
				cliArgs = append(cliArgs, "-x")
			}
			return cli.Main(cliArgs...)
		} else if def, found := g.byname[""]; found {
			return def.Main()
		}
		fmt.Println(Usage(g))
		return nil
	} else if _, found := g.byname[args[0]]; n == 1 && !found {
		// only check for script if args[0] isn't a command
		buf, err := ioutil.ReadFile(cliArgs[0])
		if cliArgs[0] == "-" || (err == nil && utf8.Valid(buf) &&
			bytes.HasPrefix(buf, []byte("#!/usr/bin/goes"))) {
			// e.g. /usr/bin/goes SCRIPT
			if cli == nil {
				return fmt.Errorf("has no cli")
			}
			for _, t := range []string{"-f", "-x"} {
				if cliFlags.ByName[t] {
					cliArgs = append(cliArgs, t)
				}
			}
			return cli.Main(cliArgs...)
		}
	} else {
		cmd.Swap(args)
	}

	if _, found := cmd.Helpers[args[0]]; found {
		return g.byname[args[0]].Main(args[1:]...)
	}

	g.Shift(args)

	v := g.byname[args[0]]
	if v == nil {
		return fmt.Errorf("%s: command not found", args[0])
	}

	k := cmd.WhatKind(v)
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
		err = fmt.Errorf("%s: %v", args[0], err)
	}
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

// Shift first recognized command to args[0], so,
//
//	OPTIONS... COMMAND [ARGS]...
//
// becomes
//
//	COMMAND OPTIONS... [ARGS]...
func (g *Goes) Shift(args []string) {
	for i := range args {
		if _, found := g.byname[args[i]]; found {
			if i > 0 {
				name := args[i]
				copy(args[1:i+1], args[:i])
				args[0] = name
			}
			break
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
