// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ps

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/internal/proc"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "ps" }

func (Command) Usage() string {
	return "ps [OPTION]..."
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "print process state",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Print information for current processes.
	The default list is limitted to processes on controlling TTY.

	-e  Select all processes.
	-f  Full format listing.`,
	}
}

func (Command) Main(args ...string) error {
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
	stime := func(t time.Time) string {
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
	case flag.ByName["-f"]:
		fmt.Println("UID        PID  PPID  C STIME TTY          TIME CMD")
	default:
		fmt.Println("  PID TTY          TIME CMD")
	}

	tty := make(tty)
	uid := make(uid)

	for _, ps := range l {
		if flag.ByName["-e"] || ps.stat.TtyNr == ttynr {
			switch {
			case flag.ByName["-f"]:
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
				fmt.Printf("%-8s %5d %5d %2s %5s %-7s %9s %s\n",
					uid.Name(ps.uid),
					ps.stat.Pid,
					ps.stat.Ppid,
					ps.stat.State,
					tty.Name(ps.stat.TtyNr),
					stime(ps.stat.StartTime),
					ps.stat.Utime,
					cmdline)
			default:
				fmt.Printf("%5d %-7s %9s %s\n",
					ps.stat.Pid,
					tty.Name(ps.stat.TtyNr),
					ps.stat.Utime,
					ps.stat.Comm)
			}
		}
	}
	return nil
}

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

type uid map[uint32]string

func (uid uid) Name(u uint32) string {
	name, found := uid[u]
	if found {
		return name
	}
	name = fmt.Sprint(u)
	if f, err := os.Open("/etc/passwd"); err == nil {
		defer f.Close()
		scan := bufio.NewScanner(f)
		for scan.Scan() {
			field := bytes.Split(scan.Bytes(), []byte(":"))
			var n uint32
			fmt.Sscan(string(field[2]), &n)
			uid[n] = string(field[0])
			if u == n {
				name = uid[n]
			}
		}
	}
	return name
}
