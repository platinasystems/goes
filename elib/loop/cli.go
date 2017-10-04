// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loop

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/iomux"

	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Cli struct {
	r Noder
	n *Node
	cli.Main
}

func (l *Loop) CliAdd(c *cli.Command) { l.Cli.AddCommand(c) }
func (c *Cli) SetEventNode(r Noder) {
	c.r = r
	c.n = r.GetNode()
}

type fileEvent struct {
	c *Cli
	*cli.File
}

func (c *Cli) rxReady(f *cli.File) {
	c.n.SignalEvent(&fileEvent{c: c, File: f}, c.r)
}

func (c *fileEvent) EventAction() {
	if err := c.RxReady(); err == cli.ErrQuit {
		c.c.n.SignalEvent(ErrQuit, c.c.r)
	}
}

func (c *fileEvent) String() string { return "rx-ready " + c.File.String() }

func (c *Cli) LoopInit(l *Loop) {
	if len(c.Main.Prompt) == 0 {
		c.Main.Prompt = "cli# "
	}
	c.Main.Start()
}

func (c *Cli) LoopExit(l *Loop) {
	c.Main.End()
}

type loggerMain struct {
	once sync.Once
	w    io.Writer
	l    *log.Logger
}

func (l *Loop) loggerInit() {
	m := &l.loggerMain
	if m.w = l.LogWriter; m.w == nil {
		m.w = os.Stdout
	}
	m.l = log.New(m.w, "", log.Lmicroseconds)
	return
}

func (l *Loop) Logf(format string, args ...interface{}) {
	m := &l.loggerMain
	m.once.Do(l.loggerInit)
	m.l.Printf(format, args...)
}
func (l *Loop) Logln(args ...interface{}) {
	m := &l.loggerMain
	m.once.Do(l.loggerInit)
	m.l.Println(args...)
}
func (m *loggerMain) Fatalf(format string, args ...interface{}) { panic(fmt.Errorf(format, args...)) }

func (l *Loop) showRuntimeStats(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	show_detail := false
	show_events := false
	colMap := map[string]bool{
		"State": false,
	}
	for !in.End() {
		switch {
		case in.Parse("d%*etail"):
			colMap["State"] = true
			show_detail = true
		case in.Parse("e%*vent"):
			show_events = true
		default:
			in.ParseError()
		}
	}

	l.flushAllActivePollerStats()

	if show_events {
		return l.showRuntimeEvents(w)
	}

	type node struct {
		Name     string  `format:"%-30s"`
		State    string  `align:"center"`
		Calls    uint64  `format:"%16d"`
		Vectors  uint64  `format:"%16d"`
		Suspends uint64  `format:"%16d"`
		Clocks   float64 `format:"%16.2f"`
	}

	ns := []node{}
	var inputSummary stats
	for _, n := range l.nodes {
		var s [2]stats
		s[0].add(&n.inputStats)
		inputSummary.add(&n.inputStats)
		inputSummary.clocks += n.outputStats.clocksSinceLastClear()
		s[1].add(&n.outputStats)
		name := n.name
		_, isIn := n.noder.(inLooper)
		_, isOut := n.noder.(outLooper)
		_, isInOut := n.noder.(inOutLooper)
		for j := range s {
			if j == 0 && !isIn && !isInOut {
				continue
			}
			if j == 1 && !isOut && !isInOut {
				continue
			}
			io := ""
			if (isIn && isOut) || isInOut {
				if j == 0 {
					io = " in"
				} else {
					io = " out"
				}
			}
			if s[j].calls > 0 || show_detail {
				state := ""
				if j == 0 {
					state = n.s.String()
				}
				ns = append(ns, node{
					Name:     name + io,
					State:    state,
					Calls:    s[j].calls,
					Vectors:  s[j].vectors,
					Suspends: s[j].suspends,
					Clocks:   s[j].clocksPerVector(),
				})
			}
		}
	}

	// Summary
	if s := inputSummary; s.calls > 0 {
		dt := time.Since(l.timeLastRuntimeClear).Seconds()
		vecsPerSec := float64(s.vectors) / dt
		clocksPerVec := float64(s.clocks) / float64(s.vectors)
		vecsPerCall := float64(s.vectors) / float64(s.calls)
		fmt.Fprintf(w, "Vectors: %d, Vectors/sec: %.2e, Clocks/vector: %.2f, Vectors/call %.2f\n",
			s.vectors, vecsPerSec, clocksPerVec, vecsPerCall)
	}

	sort.Slice(ns, func(i, j int) bool { return ns[i].Name < ns[j].Name })
	elib.Tabulate(ns).WriteCols(w, colMap)
	return
}

func (l *Loop) clearRuntimeStats(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	l.flushAllActivePollerStats()
	l.timeLastRuntimeClear = time.Now()
	for _, n := range l.nodes {
		n.inputStats.clear()
		n.outputStats.clear()
		n.e.eventStats.clear()
	}
	return
}

func (l *Loop) showEventLog(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	detail := false
	graphic := false
	matching := ""
	showFilters := false
	summary := false
	for !in.End() {
		switch {
		case in.Parse("de%*tail"):
			detail = true
		case in.Parse("f%*ilter"):
			showFilters = true
		case in.Parse("gr%*aphic"):
			graphic = true
		case in.Parse("m%*atching %v", &matching):
		case in.Parse("s%*ummary"):
			summary = true
		default:
			in.ParseError()
		}
	}

	if showFilters {
		elog.PrintFilters(w)
		return
	}

	v := elog.NewView()

	if summary {
		fmt.Fprintln(w, v.NumEvents(), "events in log")
		fmt.Fprintln(w, elog.GetSequence(), "events logged")
		return
	}

	if matching != "" {
		var eis []uint
		if eis, err = v.EventsMatching(matching, eis); err == nil {
			ne, te := len(eis), v.NumEvents()
			fmt.Fprintf(w, "%d matching of total %d\n", ne, te)
			if detail {
				v.PrintEvents(w, eis, detail)
			}
		} else {
			fmt.Fprintln(w, "bad regexp:", matching)
		}
		return
	}

	if graphic {
		l.ViewEventLog(v)
	} else {
		v.Print(w, detail)
	}
	return
}

func (l *Loop) clearEventLog(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	elog.Clear()
	return
}

func (l *Loop) configEventLog(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	var (
		s        string
		n_events uint
	)
	for !in.End() {
		switch {
		case in.Parse("f%*ilter r%*eset"):
			elog.ResetFilters()
		case in.Parse("f%*ilter d%*elete %v", &s):
			elog.AddDelEventFilter(s, true)
		case in.Parse("f%*ilter a%*dd %v", &s) || in.Parse("f%*ilter %v", &s):
			elog.AddDelEventFilter(s, false)
		case in.Parse("re%*size %d", &n_events):
			elog.Resize(n_events)
		case in.Parse("disable-after %d", &n_events):
			elog.DisableAfter(uint64(n_events))
		case in.Parse("s%*ave %s", &s):
			err = elog.SaveFile(s)
		case in.Parse("d%*ump %s", &s):
			var v elog.View
			if err = v.LoadFile(s); err == nil {
				v.Print(w, false)
			}
		default:
			in.ParseError()
		}
	}
	return
}

func (l *Loop) exec(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	var files []*os.File
	for !in.End() {
		var (
			pattern string
			names   []string
			f       *os.File
		)
		in.Parse("%s", &pattern)
		if names, err = filepath.Glob(pattern); err != nil {
			return
		}
		if len(names) == 0 {
			err = fmt.Errorf("no files matching pattern: `%s'", pattern)
			return
		}
		for _, name := range names {
			if f, err = os.OpenFile(name, os.O_RDONLY, 0); err != nil {
				return
			}
			files = append(files, f)
		}
	}
	for _, f := range files {
		var i [2]cli.Input
		i[0].Init(f)
		for !i[0].End() {
			i[1].Init(nil)
			if !i[0].Parse("%l", &i[1].Input) {
				err = i[0].Error()
				return
			}
			if err = l.Cli.ExecInput(w, &i[1]); err != nil {
				return
			}
		}
		f.Close()
	}
	return
}

func (l *Loop) comment(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	in.Skip()
	return
}

func (l *Loop) cliInit() {
	l.RegisterEventPoller(iomux.Default)
	c := &l.Cli
	c.Main.RxReady = c.rxReady
	c.AddCommand(&cli.Command{
		Name:      "show runtime",
		ShortHelp: "show main loop runtime statistics",
		Action:    l.showRuntimeStats,
	})
	c.AddCommand(&cli.Command{
		Name:      "clear runtime",
		ShortHelp: "clear main loop runtime statistics",
		Action:    l.clearRuntimeStats,
	})
	c.AddCommand(&cli.Command{
		Name:      "show event-log",
		ShortHelp: "show events in event log",
		Action:    l.showEventLog,
	})
	c.AddCommand(&cli.Command{
		Name:      "clear event-log",
		ShortHelp: "clear events in event log",
		Action:    l.clearEventLog,
	})
	c.AddCommand(&cli.Command{
		Name:      "event-log",
		ShortHelp: "event log commands",
		Action:    l.configEventLog,
	})
	c.AddCommand(&cli.Command{
		Name:      "exec",
		ShortHelp: "execute cli commands from given file(s)",
		Action:    l.exec,
	})
	c.AddCommand(&cli.Command{
		Name:      "//",
		ShortHelp: "comment",
		Action:    l.comment,
	})
}
