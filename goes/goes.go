// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build linux

// Package goes, combined with a compatibly configured Linux kernel, provides a
// monolithic embedded system.
package goes

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/prog"
)

const (
	DontFork Kind = 1 << iota
	Daemon
	Hidden
	CantPipe
)

type ByName map[string]*Goes

type Cmd interface {
	Apropos() lang.Alt
	Main(...string) error
	// The command's String() is its name.
	String() string
	Usage() string
}

type Goes struct {
	Name     string
	ByName   func(ByName)
	Close    func() error
	Complete func(...string) []string
	Help     func(...string) string
	Main     func(...string) error
	Kind     Kind
	Usage    string
	Apropos  lang.Alt
	Man      lang.Alt
}

type Kind uint16

// optional methods
type byNamer interface {
	ByName(ByName)
}

type completer interface {
	Complete(...string) []string
}

type goeser interface {
	goes() *Goes
}

type helper interface {
	Help(...string) string
}

type kinder interface {
	Kind() Kind
}

type manner interface {
	Man() lang.Alt
}

func New(cmd ...Cmd) ByName {
	m := make(ByName)
	m.Plot(cmd...)
	return m
}

func (byName ByName) Complete(prefix string) (ss []string) {
	for k, g := range byName {
		if strings.HasPrefix(k, prefix) && g.Kind.IsInteractive() {
			ss = append(ss, k)
		}
	}
	if len(ss) > 0 {
		sort.Strings(ss)
	}
	return
}

// Main runs the arg[0] command in the current context.
// When run w/o args, this uses os.Args.
//
// If len(args) == 1 and args[0] doesn't match a mapped command, this will run
// the "cli".
//
// If the args has "-h", "-help", or "--help", this runs
// ByName["help"].Main(args...) to print text.
//
// Similarly for "-apropos", "-complete", "-license", "-man", "-patents",
// "-usage" and "-version".
//
// If the command is a daemon, this fork exec's itself twice to disassociate
// the daemon from the tty and initiating process.
func (byName ByName) Main(args ...string) error {
	if len(args) == 0 {
		if len(os.Args) == 0 {
			return nil
		}
		args = os.Args
	}

	if _, found := byName[args[0]]; !found {
		base := filepath.Base(args[0])
		if _, found = byName[base]; found {
			// e.g. ./COMMAND [ARGS]...
			args[0] = base
		} else if len(args) > 1 {
			if args[0] == prog.Install {
				buf, err := ioutil.ReadFile(args[1])
				if err == nil && utf8.Valid(buf) {
					// e.g. /usr/bin/goes SCRIPT
					args = []string{"source", args[1]}
				} else {
					// e.g. /usr/bin/goes COMMAND [ARGS]...
					args = args[1:]
				}
			} else {
				// e.g. ./goes COMMAND [ARGS]...
				args = args[1:]
			}
		} else {
			// e.g. ./goes
			args = []string{"cli"}
		}
	}

	name, args := byName.pseudonym(args)
	g := byName[name]
	if g == nil {
		return fmt.Errorf("%s: command not found", name)
	}
	if g.Kind.IsDaemon() {
		sig := make(chan os.Signal)
		signal.Notify(sig, syscall.SIGTERM)
		defer func(sig chan os.Signal) {
			sig <- syscall.SIGABRT
		}(sig)
		go g.wait(sig)
	}
	err := g.Main(args...)
	if err == io.EOF {
		err = nil
	}
	if err != nil && !g.Kind.IsDaemon() {
		err = fmt.Errorf("%s: %v", name, err)
	}
	return err
}

// parse pseudo flags, e.g.
//
//	COMMAND {"-license", "-patents", or "-version"}
// returns
//	{"license", "patents", or "version"}
//
//	COMMAND {"-apropos", "-man", or "-usage"}
// returns
//	{"apropos", "man", or "usage"} COMMAND
//
//	COMMAND {"-complete" or "-help"} [ARGS]...
// returns
//	{"complete" or "help"} COMMAND [ARGS]...
func (byName ByName) pseudonym(args []string) (string, []string) {
	name := args[0]
	args = args[1:]
	flag, args := flags.New(args,
		"-apropos", "--apropos",
		"-complete", "--complete",
		"-help", "--help", "-h",
		"-license", "--license",
		"-man", "--man",
		"-patents", "--patents",
		"-usage", "--usage",
		"-version", "--version",
	)
	flag.Aka("-apropos", "--apropos")
	flag.Aka("-complete", "--complete")
	flag.Aka("-help", "--help", "-h")
	flag.Aka("-license", "--license")
	flag.Aka("-man", "--man")
	flag.Aka("-patents", "--patents")
	flag.Aka("-usage", "--usage")
	flag.Aka("-version", "--version")
	for _, x := range []string{
		"-license",
		"-patents",
		"-version",
	} {
		if flag[x] {
			return x[1:], args[:0]
		}
	}
	for _, x := range []string{
		"-apropos",
		"-man",
		"-usage",
	} {
		if flag[x] {
			return x[1:], []string{name}
		}
	}
	for _, x := range []string{
		"-complete",
		"-help",
	} {
		if flag[x] {
			return x[1:], append([]string{name}, args...)
		}
	}
	return name, args
}

// Plot commands on map.
func (byName ByName) Plot(cmds ...Cmd) {
	for _, v := range cmds {
		if method, found := v.(goeser); found {
			g := method.goes()
			byName[g.Name] = g
			if g.ByName != nil {
				g.ByName(byName)
			}
			continue
		}
		name := v.String()
		if _, found := byName[name]; found {
			panic(fmt.Errorf("%s: duplicate", name))
		}
		g := &Goes{
			Name:    name,
			Main:    v.Main,
			Usage:   v.Usage(),
			Apropos: v.Apropos(),
		}
		if strings.HasPrefix(g.Usage, "\n\t") {
			g.Usage = g.Usage[1:]
		}
		if method, found := v.(byNamer); found {
			method.ByName(byName)
		}
		if method, found := v.(io.Closer); found {
			g.Close = method.Close
		}
		if method, found := v.(completer); found {
			g.Complete = method.Complete
		}
		if method, found := v.(helper); found {
			g.Help = method.Help
		}
		if method, found := v.(kinder); found {
			g.Kind = method.Kind()
		}
		if method, found := v.(manner); found {
			g.Man = method.Man()
		}
		byName[g.Name] = g
	}
}

func (g *Goes) goes() *Goes { return g }

func (g *Goes) wait(ch chan os.Signal) {
	for sig := range ch {
		if sig == syscall.SIGTERM {
			if g.Close != nil {
				if err := g.Close(); err != nil {
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

func (k Kind) IsDontFork() bool    { return (k & DontFork) == DontFork }
func (k Kind) IsDaemon() bool      { return (k & Daemon) == Daemon }
func (k Kind) IsHidden() bool      { return (k & Hidden) == Hidden }
func (k Kind) IsInteractive() bool { return (k & (Daemon | Hidden)) == 0 }
func (k Kind) IsCantPipe() bool    { return (k & CantPipe) == CantPipe }

func (k Kind) String() string {
	s := "unknown"
	switch k {
	case DontFork:
		s = "don't fork"
	case Daemon:
		s = "daemon"
	case Hidden:
		s = "hidden"
	}
	return s
}
