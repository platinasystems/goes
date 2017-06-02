// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ps

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/proc"
)

const (
	Name    = "ps"
	Apropos = "print process state"
	Usage   = "ps [OPTION]..."
	Man     = `
DESCRIPTION
	Print information for current processes.
	The default list is limitted to processes on controlling TTY.

	-e  Select all processes.
	-f  Full format listing.`
)

type Interface interface {
	Apropos() lang.Alt
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }

func (cmd) Main(args ...string) error {
	var ttynr uint

	flag, args := flags.New(args, "-e", "-f")
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	pid := os.Getpid()
	now := time.Now()
	boy := time.Date(now.Year(), 1, 1, 12, 0, 0, 0, now.Location())
	bod := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0,
		now.Location())
	start := func(t time.Time) string {
		if t.Before(boy) {
			return fmt.Sprintf("%4d", t.Year())
		} else if t.Before(bod) {
			return fmt.Sprintf("%s%02d", t.Month().String()[:3],
				t.Day())
		}
		return fmt.Sprintf("%2d:%02d", t.Hour(), t.Minute())
	}

	fns, err := filepath.Glob("/proc/[0-9]*/stat")
	if err != nil {
		return err
	}

	l := make([]*ps, len(fns))
	for i, fn := range fns {
		l[i], err = newps(fn)
		if err != nil {
			return fmt.Errorf("%s: %v", fn, err)
		}
		if l[i].stat.Pid == pid {
			ttynr = l[i].stat.TtyNr
		}
	}

	sort.Slice(l, func(i, j int) bool {
		if l[i].uid == l[j].uid {
			if l[i].stat.TtyNr == l[j].stat.TtyNr {
				return l[i].stat.Pid < l[j].stat.Pid
			}
			return l[i].stat.TtyNr < l[j].stat.TtyNr
		}
		return l[i].uid < l[j].uid
	})

	switch {
	case flag["-f"]:
		fmt.Printf("%5s %8s %6s %6s %5s %5s %16s %s\n",
			"UID", "TTY", "PID", "PPID", "STATE", "START", "TIME",
			"COMMAND")
	default:
		fmt.Printf("%6s %8s %16s %s\n",
			"PID", "TTY", "TIME", "COMMAND")
	}

	tty := make(tty)

	for _, ps := range l {
		if flag["-e"] || ps.stat.TtyNr == ttynr {
			switch {
			case flag["-f"]:
				cmdline := ps.stat.Comm
				fn := fmt.Sprintf("/proc/%d/cmdline",
					ps.stat.Pid)
				buf, err := ioutil.ReadFile(fn)
				if err == nil && len(buf) > 0 {
					for i, b := range buf {
						if b == 0 {
							buf[i] = ' '
						}
					}
					cmdline = string(buf)
				}
				fmt.Printf("%5d %8s %6d %6d %5s %5s %16s %s\n",
					ps.uid,
					tty.Name(ps.stat.TtyNr),
					ps.stat.Pid,
					ps.stat.Ppid,
					"  "+ps.stat.State+"  ",
					start(ps.stat.StartTime),
					ps.stat.Utime,
					cmdline)
			default:
				fmt.Printf("%6d %8s %16s %s\n",
					ps.stat.Pid,
					tty.Name(ps.stat.TtyNr),
					ps.stat.Utime,
					ps.stat.Comm)
			}
		}
	}
	return nil
}

func (cmd) Man() lang.Alt  { return man }
func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

type ps struct {
	uid  uint32
	gid  uint32
	stat proc.Stat
}

func newps(fn string) (p *ps, err error) {
	fi, err := os.Stat(fn)
	if err != nil {
		return
	}
	f, err := os.Open(fn)
	if err != nil {
		return
	}
	defer f.Close()
	p = &ps{
		uid: fi.Sys().(*syscall.Stat_t).Uid,
		gid: fi.Sys().(*syscall.Stat_t).Gid,
	}
	err = proc.Load(&p.stat).FromFile(fn)
	return
}

type tty map[uint]string

func (tty tty) Name(u uint) string {
	const devname = "DEVNAME="

	name, found := tty[u]
	if found {
		return name
	}

	maj := (u >> 8) & 0xfff
	min := (u & 0xff) | ((u >> 12) & 0xfff00)
	fn := fmt.Sprintf("/sys/dev/char/%d:%d/uevent", maj, min)
	if f, err := os.Open(fn); err == nil {
		sc := bufio.NewScanner(f)
		if !sc.Scan() {
			panic(fmt.Errorf("%s: missing MAJOR=#", fn))
		}
		if !sc.Scan() {
			panic(fmt.Errorf("%s: missing MINOR=#", fn))
		}
		if !sc.Scan() {
			panic(fmt.Errorf("%s: missing DEVNAME=#", fn))
		}
		s := sc.Text()
		if !strings.HasPrefix(s, devname) {
			panic(fmt.Errorf("%s: has %s; expect %s",
				fn, s, devname))
		}
		name = strings.TrimPrefix(s, devname)
	} else {
		fn := fmt.Sprintf("/dev/pts/%d", min)
		if fi, err := os.Stat(fn); err == nil {
			rdev := fi.Sys().(*syscall.Stat_t).Rdev
			if rdev/256 == uint64(maj) && rdev%256 == uint64(min) {
				name = strings.TrimPrefix(fn, "/dev/")
			}
		}
	}
	if len(name) == 0 {
		name = "?"
	}
	tty[u] = name
	return name
}
