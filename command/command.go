// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package command provides a named reference to bundled commands.
package command

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/platinasystems/go/flags"
	"github.com/platinasystems/go/log"
	"github.com/platinasystems/go/nocomment"
	"github.com/platinasystems/go/parms"
	"github.com/platinasystems/go/pidfile"
	"github.com/platinasystems/go/slice_args"
	"github.com/platinasystems/go/slice_string"
	"github.com/platinasystems/go/url"
)

const (
	DefaultLang = "en_US.UTF-8"
	daemonFlag  = "__GOES_DAEMON__"
)

var (
	// Lang may be set prior to the first Plot for alt preferred languages
	Lang = DefaultLang
	Keys struct {
		Apropos []string
		Main    []string
		Daemon  Daemons
	}
	commands map[string]interface{}
	Daemon   map[string]int
	Apropos  map[string]string
	Man      map[string]string
	Tag      map[string]string
	Usage    map[string]string

	prog string
	path string
)

func Prog() string {
	if len(prog) == 0 {
		var err error
		prog, err = os.Readlink("/proc/self/exe")
		if err != nil {
			prog = "/usr/bin/goes"
		}
	}
	return prog
}

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

type completer interface {
	Complete(...string) []string
}

type daemoner interface {
	Daemon() int // if present, Daemon() should return run level
	mainer
}

type Helper interface {
	Help(...string) string
}

type mainer interface {
	Main(...string) error
}

type mainstringer interface {
	mainer
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
	ilvl := commands[d[i]].(daemoner).Daemon()
	jlvl := commands[d[j]].(daemoner).Daemon()
	return ilvl < jlvl || (ilvl == jlvl && d[i] < d[j])
}

func (d Daemons) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func IsDaemon(name string) bool {
	if commands == nil {
		return false
	}
	cmd, found := commands[name]
	if found {
		_, found = cmd.(daemoner)
	}
	return found
}

//Find returns the command with the given name or nil.
func Find(name string) (interface{}, error) {
	var err error
	v, found := commands[name]
	if !found {
		err = fmt.Errorf("%s: command not found", name)
	}
	return v, err
}

// globs returns a list of directories or files with the given pattern.
// A pattern w/o `*` or `?` is changed to pattern+* to glob matching prefaces.
// For example,
//
//	"some/where/*"
//		returns all of the "some/where" files
//	"some/where*"
//		returns "some/where/" plus all of the contained files
//	"some/where/*.txt"
//		returns all of the ".txt" suffixed files in "some/where"
//	"some/where/file.txt"
//		returns nothing for fully qualified file name
func Globs(pattern string) (c []string) {
	if strings.ContainsAny(pattern, "*?[]") {
		if globs, err := filepath.Glob(pattern); err == nil {
			c = globs
		}
		return
	}
	fi, err := os.Stat(pattern)
	if err == nil && !fi.IsDir() {
		return
	}
	pattern += "*"
	globs, err := filepath.Glob(pattern)
	if err != nil {
		globs = globs[:0]
	}
	for _, name := range globs {
		fi, err := os.Stat(name)
		if err != nil {
			continue
		} else if fi.IsDir() {
			c = append(c, name+string(os.PathSeparator))
		} else {
			t, err := filepath.Match(pattern, name)
			if err == nil && t {
				c = append(c, name)
			}
		}
	}
	if len(c) == 1 && c[0][len(c[0])-1] == os.PathSeparator {
		c = append(c, Globs(c[0])...)
	}
	return
}

// Main runs the arg[0] command in the current context.
// Use Shell() to iterate command input.
//
// If the th remaining args has "-h", "-help", "--help", or "help", the
// command's Help() or Man() methods are used to print text.
//
// Similarly for "-apropos", "-complete", "-man", and "-usage".
//
// If the command is a daemon, this fork exec's itself twice to disassociate
// the daemon from the tty and initiating process.
func Main(args ...string) (err error) {
	if len(args) < 1 {
		return nil
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
	isDaemon := IsDaemon(name)
	daemonFlagValue := os.Getenv(daemonFlag)
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
	if commands == nil {
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
			"REDISD=" + os.Getenv("REDISD"),
			"MACHINED=" + os.Getenv("MACHINED"),
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
	if commands == nil {
		commands = make(map[string]interface{})
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
		if _, found := commands[k]; found {
			panic(fmt.Errorf("%s: duplicate", k))
		}
		commands[k] = cmd
		if _, found := cmd.(mainer); found {
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

// Shell interates command input.
//
// Commands may continue to next line, e.g.:
//
//	echo hello \
//	world
//
// Commands may be pipelined, e.g.:
//
//	ls -lR | more
//	ls -Lr |
//	more
//
// Command comments are ignored, e.g.:
//
//	mount -t tmpfs none /tmp # scratch
//
// Similar for leading whitespace, e.g.:
//
//		echo why\?
//
// However, the shell doesn't have command blocks so there isn't much reason to
// indent input for anything other than "here documents".
//
// A pipeline may redirect input and output of the first and last commands
// respectively, e.g.:
//
//	cat <<-EOF | wc -l > lines.txt
//		...
//	EOF
//
// The redirected files may be URL's, e.g.:
//
//	source https://github.com/MYSTUFF/MYSCRIPT
//
// Redirected output may be tee'd to a truncated or appended file with `>>>`
// and `>>>>` respectively, e.g.:
//
//	dmesg | grep goes >>> goes.log
func Shell(p Prompter) error {
	var (
		rc  io.ReadCloser
		wc  io.WriteCloser
		in  io.Reader
		out io.Writer
		err error

		closers []io.Closer

		pin, pout *os.File
	)
	pl := slice_args.New("|")
	catline := func(prompt string) (line string, err error) {
		for {
			var s string
			s, err = p.Prompt(prompt)
			if err != nil {
				return
			}
			if !strings.HasSuffix(s, "\\") {
				line += s
				return
			}
			line += s[:len(s)-1]
			prompt = "... "
		}
	}

commandLoop:
	for {
		for _, c := range closers {
			c.Close()
		}
		closers = closers[:0]
		if err != nil {
			if err == io.EOF {
				return nil
			}
			if err.Error() != "exit status 1" {
				fmt.Fprintln(os.Stderr, err)
			}
			err = nil
		}
		pl.Reset()
		prompt := filepath.Base(Prog()) + "> "
		if hn, err := os.Hostname(); err == nil {
			prompt = hn + "> "
		}
	pipelineLoop:
		for {
			var s string
			s, err = catline(prompt)
			if err != nil {
				return err
			}
			s = strings.TrimLeft(s, " \t")
			if len(s) == 0 {
				continue pipelineLoop
			}
			s = nocomment.New(s)
			if len(s) == 0 {
				continue pipelineLoop
			}
			pl.Slice(slice_string.New(s)...)
			if pl.More {
				prompt = "| "
			} else {
				break pipelineLoop
			}
		}
		if len(pl.Slices) == 0 {
			continue commandLoop
		}
		end := len(pl.Slices) - 1
		name := pl.Slices[end][0]
		if _, err = Find(name); err != nil {
			continue commandLoop
		}

		if end == 0 &&
			(IsDaemon(name) || Tag[name] == "builtin" ||
				name == os.Args[0]) {
			err = Main(pl.Slices[end]...)
			continue commandLoop
		}

		for i := 1; i <= end; i++ {
			_, found := map[string]struct{}{
				"cli":    struct{}{},
				"cd":     struct{}{},
				"env":    struct{}{},
				"exit":   struct{}{},
				"export": struct{}{},
				"resize": struct{}{},
				"source": struct{}{},
			}[name]
			if found || IsDaemon(name) {
				err = fmt.Errorf("%s: can't pipe\n", name)
				continue commandLoop
			}
		}

		iparm, args := parms.New(pl.Slices[0], "<", "<<", "<<-")
		pl.Slices[0] = args
		oparm, args := parms.New(pl.Slices[end],
			">", ">>", ">>>", ">>>>")
		pl.Slices[end] = args

		in = nil
		if fn := iparm["<"]; len(fn) > 0 {
			rc, err = url.Open(fn)
			if err != nil {
				continue commandLoop
			}
			in = rc
			closers = append(closers, rc)
		} else if len(iparm["<<"]) > 0 || len(iparm["<<-"]) > 0 {
			var trim bool
			lbl := iparm["<<"]
			if len(lbl) == 0 {
				lbl = iparm["<<-"]
				trim = true
			}
			var r, w *os.File
			r, w, err = os.Pipe()
			if err != nil {
				continue commandLoop
			}
			in = r
			closers = append(closers, r)
			go func(w io.WriteCloser, lbl string) {
				defer w.Close()
				prompt := "<<" + fn + " "
				for {
					s, err := catline(prompt)
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
		out = os.Stdout
		if fn := oparm[">"]; len(fn) > 0 {
			wc, err = url.Create(fn)
			if err != nil {
				continue commandLoop
			}
			out = wc
			closers = append(closers, wc)
		} else if fn = oparm[">>"]; len(fn) > 0 {
			wc, err = url.Append(fn)
			if err != nil {
				continue commandLoop
			}
			out = wc
			closers = append(closers, wc)
		} else if fn := oparm[">>>"]; len(fn) > 0 {
			wc, err = url.Create(fn)
			if err != nil {
				continue commandLoop
			}
			out = io.MultiWriter(os.Stdout, wc)
			closers = append(closers, wc)
		} else if fn := oparm[">>"]; len(fn) > 0 {
			wc, err = url.Append(fn)
			if err != nil {
				continue commandLoop
			}
			out = io.MultiWriter(os.Stdout, wc)
			closers = append(closers, wc)
		}

		for i := 0; i < len(pl.Slices); i++ {
			c := exec.Command(Prog(), pl.Slices[i][1:]...)
			c.Args[i] = pl.Slices[i][0]
			c.Stderr = os.Stderr
			if i == 0 {
				c.Stdin = in
			} else {
				c.Stdin = pin
			}
			if i == end {
				c.Stdout = out
			} else {
				pin, pout, err = os.Pipe()
				if err != nil {
					continue commandLoop
				}
				os.Stdout = pout
			}
			if err = c.Start(); err != nil {
				err = fmt.Errorf("child: %v: %v", c.Args, err)
				continue commandLoop
			}
			if i == end {
				err = c.Wait()
			} else {
				go func(c *exec.Cmd) {
					err := c.Wait()
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
					}
					if c.Stdout != os.Stdout {
						m, found := c.Stdout.(closer)
						if found {
							m.Close()
						}
					}
					if c.Stdin != os.Stdin {
						m, found := c.Stdin.(closer)
						if found {
							m.Close()
						}
					}
				}(c)
			}
		}
	}
	return fmt.Errorf("oops, shouldn't be here")
}

// Sort keys
func Sort() {
	sort.Strings(Keys.Apropos)
	sort.Strings(Keys.Main)
	sort.Sort(Keys.Daemon)
}
