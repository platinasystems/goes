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

	"github.com/platinasystems/go/internal/flags"
)

const (
	InstallName = "/usr/bin/goes"
	DefaultLang = "en_US.UTF-8"
)

const (
	DontFork Kind = 1 << iota
	Daemon
	Hidden
)

var (
	Exit = os.Exit

	// Lang may be set prior to the first Plot for alt preferred languages
	Lang = DefaultLang
)

type ByName map[string]*Goes

type Goes struct {
	Name     string
	ByName   func(ByName)
	Close    func() error
	Complete func(...string) []string
	Help     func(...string) string
	Main     func(...string) error
	Kind     Kind
	Usage    string
	Apropos  map[string]string
	Man      map[string]string
}

type Kind uint16

type aproposer interface {
	Apropos() map[string]string
}

type byNamer interface {
	ByName(ByName)
}

type completer interface {
	Complete(...string) []string
}

type helper interface {
	Help(...string) string
}

type kinder interface {
	Kind() Kind
}

type mainer interface {
	Main(...string) error
}

type manner interface {
	Man() map[string]string
}

type usager interface {
	Usage() string
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
// When run w/o args this uses os.Args and exits instead of returns on error.
// Use cli to iterate command input.
//
// If the args has "-h", "-help", or "--help", this runs
// ByName["help"].Main(args...) to print text.
//
// Similarly for "-apropos", "-complete", "-man", and "-usage".
//
// If the command is a daemon, this fork exec's itself twice to disassociate
// the daemon from the tty and initiating process.
func (byName ByName) Main(args ...string) error {
	if len(args) == 0 {
		args = os.Args
		switch len(args) {
		case 0:
			return nil
		case 1:
			if filepath.Base(args[0]) == ProgBase() {
				args = []string{"cli"}
			}
		}
	}

	if _, found := byName[args[0]]; !found {
		if args[0] == InstallName && len(args) > 2 {
			buf, err := ioutil.ReadFile(args[1])
			if err == nil && utf8.Valid(buf) {
				args = []string{"source", args[1]}
			} else {
				args = args[1:]
			}
		} else {
			args = args[1:]
		}
	}

	name := args[0]
	args = args[1:]
	flag, args := flags.New(args,
		"-h", "-help", "--help",
		"-apropos", "--apropos",
		"-man", "--man",
		"-usage", "--usage",
		"-complete", "--complete")
	flag.Aka("-h", "-help", "--help")
	flag.Aka("-apropos", "--apropos")
	flag.Aka("-complete", "--complete")
	flag.Aka("-man", "--man")
	flag.Aka("-usage", "--usage")
	targs := []string{name}
	switch {
	case flag["-h"]:
		name = "help"
		if len(args) == 0 {
			args = append(targs, args...)
		} else {
			args = targs
		}
	case flag["-apropos"]:
		args = targs
		name = "apropos"
	case flag["-man"]:
		args = targs
		name = "man"
	case flag["-usage"]:
		args = targs
		name = "usage"
	case flag["-complete"]:
		name = "complete"
		if len(args) == 0 {
			args = append(targs, args...)
		} else {
			args = targs
		}
	}
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

// Plot commands on map.
func (byName ByName) Plot(cmds ...interface{}) {
	for _, v := range cmds {
		g, ok := v.(*Goes)
		if ok {
			byName[g.Name] = g
			if g.ByName != nil {
				g.ByName(byName)
			}
			continue
		}
		g = new(Goes)
		if method, found := v.(fmt.Stringer); found {
			g.Name = method.String()
		} else {
			panic(fmt.Errorf("%T: doesn't have String method", v))
		}
		if _, found := byName[g.Name]; found {
			panic(fmt.Errorf("%s: duplicate", g.Name))
		}
		if method, found := v.(mainer); found {
			g.Main = method.Main
		} else {
			panic(fmt.Errorf("%s: doesn't have Main method",
				g.Name))
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
		if method, found := v.(usager); found {
			g.Usage = method.Usage()
		}
		if method, found := v.(aproposer); found {
			g.Apropos = method.Apropos()
		}
		if method, found := v.(manner); found {
			g.Man = method.Man()
		}
		byName[g.Name] = g
	}
}

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
