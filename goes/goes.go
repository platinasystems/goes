// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build linux

// Package goes, combined with a compatibly configured Linux kernel, provides a
// monolithic embedded system.
package goes

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/go/internal/prog"
	"github.com/platinasystems/go/internal/shellutils"
	"github.com/platinasystems/go/internal/url"
)

const (
	VerboseQuiet = iota
	VerboseVerify
	VerboseDebug
)

type Blocker interface {
	Block(*Goes, shellutils.List) (*shellutils.List, func(io.Reader, io.Writer, io.Writer, bool, bool) error, error)
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

	Status    error
	Verbosity int

	cache  cache
	parent *Goes

	EnvMap map[string]string

	FunctionMap map[string]Function
}

type Function struct {
	Name       string
	Definition []string
	RunFun     func(stdin io.Reader, stdout io.Writer, stderr io.Writer, isFirst bool, isLast bool) error
}

func (g *Goes) ProcessPipeline(ls shellutils.List) (*shellutils.List, *shellutils.Word, func(io.Reader, io.Writer, io.Writer) error, error) {
	var (
		closers []io.Closer
		term    shellutils.Word
	)
	isLast := false
	pipeline := make([]func(io.Reader, io.Writer, io.Writer, bool, bool) error, 0)
	for len(ls.Cmds) != 0 && !isLast {
		cl := ls.Cmds[0]
		term = cl.Term
		if term.String() != "|" {
			isLast = true
		}

		var runfun func(stdin io.Reader, stdout io.Writer, stderr io.Writer, isFirst bool, isLast bool) error
		name := cl.Cmds[0].String()
		if v := g.ByName[name]; v != nil {
			if method, found := v.(Blocker); found {
				var (
					newls *shellutils.List
					err   error
				)
				newls, runfun, err = method.Block(g, ls)
				if err != nil {
					return nil, nil, nil, err
				}

				ls = *newls
				cl = ls.Cmds[0]
				ls.Cmds = ls.Cmds[1:]
				term = cl.Term
				if term.String() != "|" {
					isLast = true
				}
				pipeline = append(pipeline, runfun)
				continue
			}
		}
		runfun, err := g.ProcessCommand(cl, &closers)
		if err != nil {
			return nil, nil, nil, err
		}
		ls.Cmds = ls.Cmds[1:]
		pipeline = append(pipeline, runfun)
	}

	pipefun, err := g.MakePipefun(pipeline, &closers)
	return &ls, &term, pipefun, err
}

func (g *Goes) ProcessCommand(cl shellutils.Cmdline, closers *[]io.Closer) (func(stdin io.Reader, stdout io.Writer, stderr io.Writer, isFirst bool, isLast bool) error, error) {
	runfun := func(stdin io.Reader, stdout io.Writer, stderr io.Writer, isFirst bool, isLast bool) error {
		envMap, args := cl.Slice(func(k string) string {
			v, def := g.EnvMap[k]
			if def {
				return v
			}
			return os.Getenv(k)
		})
		// Add to our context environment if this command only set variables
		if len(args) == 0 {
			if len(envMap) != 0 {
				if g.EnvMap == nil {
					g.EnvMap = envMap
				} else {
					for k, v := range envMap {
						g.EnvMap[k] = v
					}
				}
				g.Status = nil // Successfully set variables
			}
			return nil
		}
		name := args[0]
		// check for function invocation
		if f, x := g.FunctionMap[name]; x {
			return f.RunFun(stdin, stdout, stderr, isFirst, isLast)
		}
		// check for built in command
		if v := g.ByName[name]; v != nil {
			k := cmd.WhatKind(v)
			if k.IsDaemon() {
				return fmt.Errorf(
					"use `goes-daemons start %s`",
					name)
			}
			if !isFirst || !isLast {
				if k.IsCantPipe() {
					return fmt.Errorf(
						"%s: can't pipe", name)
				}
			} else if k.IsDontFork() ||
				name == os.Args[0] {
				if method, found := v.(goeser); found {
					method.Goes(g)
				}
				return g.Main(args...)
			}
		} else if builtin, found := g.Builtins()[name]; found {
			return builtin(args[1:]...)
		} else {
			return fmt.Errorf("%s: command not found", name)
		}
		in := stdin
		if isFirst {
			var iparm *parms.Parms
			isFirst = false
			iparm, args = parms.New(args, "<", "<<", "<<-")
			if fn := iparm.ByName["<"]; len(fn) > 0 {
				rc, err := url.Open(fn)
				if err != nil {
					return err
				}
				in = rc
				*closers = append(*closers, rc)
			} else if len(iparm.ByName["<<"]) > 0 ||
				len(iparm.ByName["<<-"]) > 0 {
				var trim bool
				lbl := iparm.ByName["<<"]
				if len(lbl) == 0 {
					lbl = iparm.ByName["<<-"]
					trim = true
				}
				r, w, err := os.Pipe()
				if err != nil {
					return err
				}
				in = r
				*closers = append(*closers, r)
				go func(w io.WriteCloser, lbl string) {
					defer w.Close()
					prompt := "<<" + fn + " "
					for {
						s, err := g.Catline(prompt)
						if err != nil || s == lbl {
							break
						}
						if trim {
							s = strings.TrimLeft(s, " \t")
						}
						fmt.Fprintln(w, s)
					}
				}(w, lbl)
			}
		}
		out := stdout
		if isLast {
			var oparm *parms.Parms
			oparm, args = parms.New(args, ">", ">>", ">>>", ">>>>")
			if fn := oparm.ByName[">"]; len(fn) > 0 {
				wc, err := url.Create(fn)
				if err != nil {
					return err
				}
				out = wc
				*closers = append(*closers, wc)
			} else if fn = oparm.ByName[">>"]; len(fn) > 0 {
				wc, err := url.Append(fn)
				if err != nil {
					return err
				}
				out = wc
				*closers = append(*closers, wc)
			} else if fn := oparm.ByName[">>>"]; len(fn) > 0 {
				wc, err := url.Create(fn)
				if err != nil {
					return err
				}
				out = io.MultiWriter(os.Stdout, wc)
				*closers = append(*closers, wc)
			} else if fn := oparm.ByName[">>"]; len(fn) > 0 {
				wc, err := url.Append(fn)
				if err != nil {
					return err
				}
				out = io.MultiWriter(os.Stdout, wc)
				*closers = append(*closers, wc)
			}
		}
		var envStr []string
		if len(envMap) != 0 {
			envStr = make([]string, 0)
			for k, v := range envMap {
				envStr = append(envStr, fmt.Sprintf("%s=%s", k, v))
			}
		}
		if g.Verbosity >= VerboseVerify {
			fmt.Println("+", strings.Join(envStr, " "), strings.Join(args, " "))
		}
		x := g.Fork(args...)
		if len(envStr) != 0 {
			x.Env = os.Environ()
			for _, s := range envStr {
				x.Env = append(x.Env, s)
			}
		}
		x.Stdin = in
		x.Stdout = out
		x.Stderr = stderr

		if err := x.Start(); err != nil {
			err = fmt.Errorf("child: %v: %v", x.Args, err)
			return err
		}
		if isLast {
			err := x.Wait()
			g.Status = err
		} else {
			go func(x *exec.Cmd) {
				err := x.Wait()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
				if x.Stdout != os.Stdout {
					m, found := x.Stdout.(io.Closer)
					if found {
						m.Close()
					}
				}
				if x.Stdin != os.Stdin {
					m, found := x.Stdin.(io.Closer)
					if found {
						m.Close()
					}
				}
			}(x)
		}
		return nil
	}
	return runfun, nil
}

func (g *Goes) MakePipefun(pipeline []func(io.Reader, io.Writer, io.Writer, bool, bool) error, closers *[]io.Closer) (func(io.Reader, io.Writer, io.Writer) error, error) {
	pipefun := func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
		var (
			err error
			pin *os.File
		)
		defer func() {
			for _, c := range *closers {
				c.Close()
			}
		}()
		in := stdin
		end := len(pipeline) - 1
		for i, runfun := range pipeline {
			out := stdout
			if i != end {
				var pout *os.File
				pin, pout, err = os.Pipe()
				if err != nil {
					break
				}
				out = pout
			}
			err = runfun(in, out, stderr, i == 0, i == end)
			if err != nil {
				break
			}
			in = pin
		}
		return err
	}
	return pipefun, nil
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
	if g.Verbosity >= VerboseDebug {
		fmt.Printf("F*$=%v %v\n", g.Status, args)
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

	var v cmd.Cmd
	var k cmd.Kind
	var found bool
	if len(args) > 0 {
		v, found = g.ByName[args[0]]
		if found {
			k = cmd.WhatKind(v)
		}
	}
	if !found || !k.IsNoCLIFlags() {
		cli, clifound := g.ByName["cli"]
		if clifound {
			cli.(goeser).Goes(g)
		}
		cliFlags, cliArgs := flags.New(args, "-d", "-f", "-no-liner", "-x")
		if cliFlags.ByName["-d"] && g.Verbosity < VerboseDebug {
			g.Verbosity = VerboseDebug
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
		} else if n == 1 && !found {
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

	if g.shift(args) {
		v, found = g.ByName[args[0]]
	}

	if g.Verbosity >= VerboseDebug {
		fmt.Printf("$=%v %v\n", g.Status, args)
	}

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
func (g *Goes) shift(args []string) bool {
	for i := range args {
		if _, found := g.ByName[args[i]]; found {
			if i > 0 {
				name := args[i]
				copy(args[1:i+1], args[:i])
				args[0] = name
			}
			return true
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
			return true
		}
	}
	return false
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
				time.Sleep(500 * time.Millisecond)
			}
			os.Stdout.Sync()
			os.Stderr.Sync()
			os.Stdout.Close()
			os.Stderr.Close()
			os.Exit(0)
		}
		break
	}
}

func (g *Goes) ensureTerminated(ls shellutils.List) (*shellutils.List, error) {
	for {
		term := ""
		for _, cl := range ls.Cmds {
			term = cl.Term.String()
			if term != "||" && term != "&&" && term != "|" {
				return &ls, nil
			}
		}
		newls, err := shellutils.Parse(fmt.Sprintf("%s>>", term), g.Catline)
		if err != nil {
			return nil, err
		}
		for _, cl := range (*newls).Cmds {
			ls.Cmds = append(ls.Cmds, cl)
		}
	}
	return nil, errors.New("ensureTerminated: internal error")
}

type piperun struct {
	f func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error
	t shellutils.Word
}

func (g *Goes) ProcessList(ls shellutils.List) (*shellutils.List, *shellutils.Word, func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error, error) {
	var (
		pipeline []piperun
		term     shellutils.Word
	)

	newls, err := g.ensureTerminated(ls)
	if err != nil {
		return nil, nil, nil, err
	}
	ls = *newls
	for len(ls.Cmds) != 0 {
		nextls, term, runner, err := g.ProcessPipeline(ls)
		if err != nil {
			return nil, nil, nil, err
		}
		ls = *nextls
		pipeline = append(pipeline, piperun{f: runner, t: *term})
		if term.String() != "&&" && term.String() != "||" {
			break
		}
	}

	listfun, err := g.MakeListFunc(pipeline)

	return &ls, &term, listfun, nil
}

func (g *Goes) MakeListFunc(pipeline []piperun) (func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error, error) {
	listfun := func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
		var err error
		skipNext := false
		for _, runfun := range pipeline {
			term := runfun.t
			if !skipNext {
				err = runfun.f(stdin, stdout, stderr)
				if err != nil {
					fmt.Fprintln(stderr, err)
					g.Status = err
				}
				skipNext = false
			}
			if g.Status != nil {
				if term.String() == "&&" {
					skipNext = true
				}
			} else {
				if term.String() == "||" {
					skipNext = true

				}
			}
		}
		return nil
	}
	return listfun, nil
}
