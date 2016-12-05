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
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"
	"unicode/utf8"

	"github.com/platinasystems/go/goes/internal/flags"
	"github.com/platinasystems/go/goes/pidfile"
	"github.com/platinasystems/go/log"
)

const (
	InstallName = "/usr/bin/goes"
	DefaultLang = "en_US.UTF-8"
	daemonFlag  = "__GOES_DAEMON__"
)

var (
	Exit = os.Exit
	// Lang may be set prior to the first Plot for alt preferred languages
	Lang = DefaultLang
	Keys struct {
		Apropos []string
		Main    []string
		Daemon  Daemons
	}
	ByName  map[string]interface{}
	Daemon  map[string]int
	Apropos map[string]string
	Man     map[string]string
	Tag     map[string]string
	Usage   map[string]string
)

// Main runs the arg[0] command in the current context.
// When run w/o args this uses os.Args and exits instead of returns on error.
// Use Shell() to iterate command input.
//
// If the args has "-h", "-help", or "--help", this runs
// ByName["help"].Main(args...) to print text.
//
// Similarly for "-apropos", "-complete", "-man", and "-usage".
//
// If the command is a daemon, this fork exec's itself twice to disassociate
// the daemon from the tty and initiating process.
func Main(args ...string) (err error) {
	var isDaemon bool

	if len(args) == 0 {
		args = os.Args
		if len(args) == 0 {
			return
		}
		defer func() {
			if err != nil && err != io.EOF {
				fmt.Fprintf(os.Stderr, "%s: %v\n",
					ProgBase(), err)
				Exit(1)
			}
		}()
	} else {
		defer func() {
			if err == io.EOF {
				err = nil
			}
			if isDaemon {
				if err != nil {
					log.Print("daemon", "err", err)
				}
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n",
					filepath.Base(Prog()), err)
			}
		}()
	}
	if _, err := Find(args[0]); err != nil {
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
	if len(args) < 1 {
		args = []string{"cli"}
	}
	name := args[0]
	args = args[1:]
	flag, args := flags.New(args,
		"-h", "-help", "--help", "help",
		"-apropos", "--apropos",
		"-man", "--man",
		"-usage", "--usage",
		"-complete", "--complete")
	flag.Aka("-h", "-help", "--help")
	flag.Aka("-apropos", "--apropos")
	flag.Aka("-complete", "--complete")
	flag.Aka("-man", "--man")
	flag.Aka("-usage", "--usage")
	isDaemon = IsDaemon(name)
	daemonFlagValue := os.Getenv(daemonFlag)
	if ByName == nil {
		err = fmt.Errorf("no commands")
		return
	}
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
		name = "-complete"
		if len(args) == 0 {
			args = append(targs, args...)
		} else {
			args = targs
		}
	}
	cmd, err := Find(name)
	if err != nil {
		return
	}
	ms, found := cmd.(mainstringer)
	if !found {
		err = fmt.Errorf("%s: can't execute", name)
		return
	}
	if !isDaemon {
		err = ms.Main(args...)
		return
	}
	switch daemonFlagValue {
	case "":
		c := exec.Command(Prog(), args...)
		c.Args[0] = name
		c.Stdin = nil
		c.Stdout = nil
		c.Stderr = nil
		c.Env = []string{
			"PATH=" + Path(),
			"TERM=linux",
			daemonFlag + "=child",
		}
		c.Dir = "/"
		c.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
			Pgid:   0,
		}
		err = c.Start()
	case "child":
		syscall.Umask(002)
		pipeOut, waitOut, terr := log.Pipe("info")
		if terr != nil {
			err = terr
			return
		}
		pipeErr, waitErr, terr := log.Pipe("err")
		if terr != nil {
			err = terr
			return
		}
		c := exec.Command(Prog(), args...)
		c.Args[0] = name
		c.Stdin = nil
		c.Stdout = pipeOut
		c.Stderr = pipeErr
		c.Env = []string{
			"PATH=" + Path(),
			"TERM=linux",
			daemonFlag + "=grandchild",
		}
		c.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
			Pgid:   0,
		}
		err = c.Start()
		<-waitOut
		<-waitErr
	case "grandchild":
		pidfn, terr := pidfile.New(name)
		if terr != nil {
			err = terr
			return
		}
		sigch := make(chan os.Signal)
		signal.Notify(sigch, syscall.SIGTERM)
		go terminate(cmd, pidfn, sigch)
		err = ms.Main(args...)
		sigch <- syscall.SIGABRT
	}
	return
}

var prog string

func Prog() string {
	if len(prog) == 0 {
		var err error
		prog, err = os.Readlink("/proc/self/exe")
		if err != nil {
			prog = InstallName
		}
	}
	return prog
}

var progbase string

func ProgBase() string {
	if len(progbase) == 0 {
		progbase = filepath.Base(Prog())
	}
	return progbase
}

var path string

func Path() string {
	if len(path) == 0 {
		path = "/bin:/usr/bin"
		dir := filepath.Dir(Prog())
		if dir != "/bin" && dir != "/usr/bin" {
			path += ":" + dir
		}
	}
	return path
}

type aproposer interface {
	Apropos() map[string]string
}

type closer interface {
	Close() error
}

type Completer interface {
	Complete(...string) []string
}

type daemoner interface {
	Daemon() int // if present, Daemon() should return run level
	Mainer
}

type Helper interface {
	Help(...string) string
}

type Mainer interface {
	Main(...string) error
}

type mainstringer interface {
	Mainer
	stringer
}

type manner interface {
	Man() map[string]string
}

// A Prompter returns the prompted input upto but not including `\n`.
type Prompter interface {
	Prompt(string) (string, error)
}

type stringer interface {
	String() string
}

type tagger interface {
	Tag() string
}

type Usager interface {
	Usage() string
}

// Daemons are sorted by level
type Daemons []string

func (d Daemons) Len() int {
	return len(d)
}

func (d Daemons) Less(i, j int) bool {
	ilvl := ByName[d[i]].(daemoner).Daemon()
	jlvl := ByName[d[j]].(daemoner).Daemon()
	return ilvl < jlvl || (ilvl == jlvl && d[i] < d[j])
}

func (d Daemons) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func IsDaemon(name string) bool {
	if ByName == nil {
		return false
	}
	cmd, found := ByName[name]
	if found {
		_, found = cmd.(daemoner)
	}
	return found
}

//Find returns the command with the given name or nil.
func Find(name string) (interface{}, error) {
	var err error
	v, found := ByName[name]
	if !found {
		err = fmt.Errorf("%s: command not found", name)
	}
	return v, err
}

func terminate(cmd interface{}, pidfn string, ch chan os.Signal) {
	for sig := range ch {
		if sig == syscall.SIGTERM {
			method, found := cmd.(io.Closer)
			if found {
				method.Close()
			}
			os.Remove(pidfn)
			fmt.Println("killed")
			os.Exit(0)
		}
		os.Remove(pidfn)
		break
	}
}

// Plot commands on respective maps and key lists.
func Plot(cmds ...interface{}) {
	lang := os.Getenv("LANG")
	if len(lang) == 0 {
		lang = Lang
	}
	if ByName == nil {
		ByName = make(map[string]interface{})
		Daemon = make(map[string]int)
		Apropos = make(map[string]string)
		Man = make(map[string]string)
		Tag = make(map[string]string)
		Usage = make(map[string]string)
	}
	for _, cmd := range cmds {
		var k string
		if method, found := cmd.(stringer); !found {
			panic("command doesn't have String method")
		} else {
			k = method.String()
		}
		if _, found := ByName[k]; found {
			panic(fmt.Errorf("%s: duplicate", k))
		}
		ByName[k] = cmd
		if _, found := cmd.(Mainer); found {
			Keys.Main = append(Keys.Main, k)
		}
		if method, found := cmd.(daemoner); found {
			Keys.Daemon = append(Keys.Daemon, k)
			Daemon[k] = method.Daemon()
		}
		if method, found := cmd.(aproposer); found {
			m := method.Apropos()
			s, found := m[lang]
			if !found {
				s = m[DefaultLang]
			}
			Apropos[k] = s
			Keys.Apropos = append(Keys.Apropos, k)
		}
		if method, found := cmd.(manner); found {
			m := method.Man()
			s, found := m[lang]
			if !found {
				s = m[DefaultLang]
			}
			Man[k] = s
		}
		if method, found := cmd.(tagger); found {
			Tag[k] = method.Tag()
		}
		if method, found := cmd.(Usager); found {
			Usage[k] = method.Usage()
		}
	}
}

// Sort keys
func Sort() {
	sort.Strings(Keys.Apropos)
	sort.Strings(Keys.Main)
	sort.Sort(Keys.Daemon)
}
