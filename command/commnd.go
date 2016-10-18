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
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/platinasystems/go/flags"
	"github.com/platinasystems/go/log"
	"github.com/platinasystems/go/nocomment"
	"github.com/platinasystems/go/parms"
	"github.com/platinasystems/go/pidfile"
	"github.com/platinasystems/go/recovered"
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

	Prog string
	Path string
)

func init() {
	s, err := os.Readlink("/proc/self/exe")
	if err == nil {
		Prog = s
	} else {
		Prog = "/usr/bin/goes"
	}
	Path = "/bin:/usr/bin"
	if s = filepath.Dir(Prog); s != "/bin" && s != "/usr/bin" {
		Path += ":" + s
	}
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

// A GetLiner will return the prompted input upto but not including `\n`.
type GetLiner func(prompt string) (string, error)

type helper interface {
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

type stringer interface {
	String() string
}

type tagger interface {
	Tag() string
}

type usager interface {
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
	cmd, found := commands[name]
	if found {
		_, found = cmd.(daemoner)
	}
	return found
}

//Find returns the command with the given name or nil.
func Find(name string) interface{} { return commands[name] }

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
func Main(args ...string) error {
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
	flag.Aka("-h", "-help", "--help", "help")
	flag.Aka("-apropos", "--apropos")
	flag.Aka("-complete", "--complete")
	flag.Aka("-man", "--man")
	flag.Aka("-usage", "--usage")
	if commands == nil {
		return fmt.Errorf("no commands")
	}
	cmd, found := commands[name]
	if !found {
		return fmt.Errorf("%s: command not found", name)
	}
	ms, found := cmd.(mainstringer)
	if !found {
		return fmt.Errorf("%s: can't execute", name)
	}
	switch {
	case flag["-h"]:
		s := Man[args[0]]
		if method, found := cmd.(helper); found {
			s = method.Help(args...)
		}
		if len(s) == 0 {
			return fmt.Errorf("%s: has no help", name)
		}
		fmt.Println(s)
	case flag["-apropos"]:
		s := Apropos[args[0]]
		if len(s) == 0 {
			return fmt.Errorf("%s: has no apropos", name)
		}
		fmt.Println(s)
	case flag["-man"]:
		s := Man[args[0]]
		if len(s) == 0 {
			return fmt.Errorf("%s: has no man", name)
		}
		fmt.Println(s)
	case flag["-usage"]:
		if s := Usage[args[0]]; len(s) == 0 {
			return fmt.Errorf("%s: has no usage", name)
		} else if strings.IndexRune(s, '\n') >= 0 {
			fmt.Print("usage:\t", s, "\n")
		} else {
			fmt.Println("usage:", s)
		}
	case flag["-complete"]:
		if m, found := cmd.(completer); found {
			for _, c := range m.Complete(args...) {
				fmt.Println(c)
			}
		} else {
			pattern := "*"
			if len(args) > 0 {
				pattern = args[len(args)-1]
			}
			for _, c := range Globs(pattern) {
				fmt.Println(c)
			}
		}
	default:
		if IsDaemon(name) {
			switch os.Getenv(daemonFlag) {
			case "":
				c := exec.Command(Prog, args[1:]...)
				c.Args[0] = name
				c.Stdin = nil
				c.Stdout = nil
				c.Stderr = nil
				c.Env = []string{
					"PATH=" + Path,
					"TERM=linux",
					daemonFlag + "=child",
				}
				c.Dir = "/"
				c.SysProcAttr = &syscall.SysProcAttr{
					Setsid: true,
					Pgid:   0,
				}
				err := c.Start()
				if err != nil {
					return fmt.Errorf("child: %v: %v",
						c.Args, err)
				}
				return c.Wait()
			case "child":
				syscall.Umask(002)
				c := exec.Command(Prog, args[1:]...)
				c.Args[0] = name
				c.Stdin = nil
				if p, err := log.Pipe("info"); err == nil {
					c.Stdout = p
				} else {
					c.Stdout = nil
				}
				if p, err := log.Pipe("err"); err == nil {
					c.Stderr = p
				} else {
					c.Stderr = nil
				}
				c.Env = []string{
					"PATH=" + Path,
					"TERM=linux",
					daemonFlag + "=grandchild",
				}
				c.SysProcAttr = &syscall.SysProcAttr{
					Setsid: true,
					Pgid:   0,
				}
				return c.Start()
			case "grandchild":
				pidfn, err := pidfile.New()
				if err == nil {
					err = recovered.New(ms).Main(args...)
					os.Remove(pidfn)
				}
				return err
			}
		} else {
			return recovered.New(ms).Main(args...)
		}
	}
	return nil
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
		if method, found := cmd.(usager); found {
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
func Shell(getline GetLiner) error {
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
			s, err = getline(prompt)
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
			fmt.Fprintln(os.Stderr, err)
			err = nil
		}
		pl.Reset()
		prompt := filepath.Base(Prog) + "> "
		if hn, err := os.Hostname(); err == nil {
			prompt = hn + "> "
		}
	pipelineLoop:
		s, err := catline(prompt)
		if err != nil {
			return err
		}
		s = strings.TrimLeft(s, " \t")
		s = nocomment.New(s)
		pl.Slice(slice_string.New(s)...)
		if pl.More {
			prompt = "| "
			goto pipelineLoop
		}
		if len(pl.Slices) < 1 {
			continue commandLoop
		}
		end := len(pl.Slices) - 1
		name := pl.Slices[end][0]
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
			c := exec.Command(Prog, pl.Slices[i][1:]...)
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
