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
	"sort"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/platinasystems/go/goes/internal/flags"
	"github.com/platinasystems/go/goes/pidfile"
)

const (
	InstallName = "/usr/bin/goes"
	DefaultLang = "en_US.UTF-8"
)

const (
	Builtin Kind = 1 << iota
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

var keys []string // cache

func (byName ByName) Keys() []string {
	// this assumes different Goes maps would have different lengths
	if len(keys) == len(byName) {
		return keys
	}
	keys = make([]string, 0, len(byName))
	for k := range byName {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (byName ByName) Complete(args ...string) (ss []string) {
	if len(args) < 1 {
		return
	}
	for _, k := range byName.Keys() {
		if strings.HasPrefix(k, args[len(args)-1]) {
			ss = append(ss, k)
		}
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
func (byName ByName) Main(args ...string) (err error) {
	var sig chan os.Signal
	defer func() {
		if err == io.EOF {
			err = nil
		}
		if err != nil {
			if sig != nil {
				fmt.Fprintln(os.Stderr, err)
			} else {
				fmt.Fprintf(os.Stderr, "%s: %v\n",
					ProgBase(), err)
			}
		}
		if sig != nil {
			sig <- syscall.SIGABRT
			os.Stdout.Close()
			os.Stderr.Close()
		}
	}()

	if len(args) == 0 {
		args = os.Args
		switch len(args) {
		case 0:
			return
		case 1:
			if args[0] == ProgBase() {
				args[0] = "cli"
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
		pidfn, terr := pidfile.New(name)
		if terr != nil {
			err = terr
			return
		}
		signal.Notify(sig, syscall.SIGTERM)
		go g.wait(pidfn, sig)
	}
	err = g.Main(args...)
	return
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

func (g *Goes) wait(pidfn string, ch chan os.Signal) {
	for sig := range ch {
		if sig == syscall.SIGTERM {
			if g.Close != nil {
				g.Close()
			}
			os.Remove(pidfn)
			fmt.Println("killed")
			os.Exit(0)
		}
		os.Remove(pidfn)
		break
	}
}

func (k Kind) IsBuiltin() bool { return (k & Builtin) == Builtin }
func (k Kind) IsDaemon() bool  { return (k & Daemon) == Daemon }

func (k Kind) String() string {
	s := "unknown"
	switch k {
	case Builtin:
		s = "builtin"
	case Daemon:
		s = "daemon"
	}
	return s
}
