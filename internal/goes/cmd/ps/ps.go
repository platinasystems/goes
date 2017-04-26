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

	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/goes/lang"
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

func (ps cmd) Main(args ...string) error {
	hz = Hz()
	flag, args := flags.New(args, "-e", "-f")
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	epoch = time.Unix(0, 0)
	now := time.Now()
	boy := time.Date(now.Year(), 1, 1, 12, 0, 0, 0, now.Location())
	bod := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0,
		now.Location())

	var si syscall.Sysinfo_t
	err := syscall.Sysinfo(&si)
	if err != nil {
		return err
	}

	thisPid := os.Getpid()
	var thisStat *psStat
	var u64 uint64

	fns, err := filepath.Glob("/proc/[0-9]*/stat")
	if err != nil {
		return err
	}

	stats := make(psStats, 0, 128)
	for _, fn := range fns {
		f, err := os.Open(fn)
		if err != nil {
			return err
		}
		fi, err := f.Stat()
		if err != nil {
			return err
		}
		t := new(psStat)
		t.uid = fi.Sys().(*syscall.Stat_t).Uid
		t.gid = fi.Sys().(*syscall.Stat_t).Gid
		_, err = fmt.Fscanf(f, "%v", &t.pid)
		if err != nil {
			return fmt.Errorf("%s: pid: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.comm)
		if err != nil {
			return fmt.Errorf("%s: comm: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.state)
		if err != nil {
			return fmt.Errorf("%s: state: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.ppid)
		if err != nil {
			return fmt.Errorf("%s: ppid: %v", fn, err)
		}
		t.modComm()
		_, err = fmt.Fscanf(f, "%v", &t.pgrp)
		if err != nil {
			return fmt.Errorf("%s: pgrp: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.session)
		if err != nil {
			return fmt.Errorf("%s: session: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.tty_nr)
		if err != nil {
			return fmt.Errorf("%s: tty_nr: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.tpgid)
		if err != nil {
			return fmt.Errorf("%s: tpgid: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.flags)
		if err != nil {
			return fmt.Errorf("%s: flags: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.minflt)
		if err != nil {
			return fmt.Errorf("%s: minflt: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.cminflt)
		if err != nil {
			return fmt.Errorf("%s: cminflt: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.majflt)
		if err != nil {
			return fmt.Errorf("%s: majflt: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.cmajflt)
		if err != nil {
			return fmt.Errorf("%s: cmajflt: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &u64)
		if err != nil {
			return fmt.Errorf("%s: utime: %v", fn, err)
		}
		t.utime = duration(u64)
		_, err = fmt.Fscanf(f, "%v", &t.stime)
		if err != nil {
			return fmt.Errorf("%s: stime: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.cutime)
		if err != nil {
			return fmt.Errorf("%s: cutime: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.cstime)
		if err != nil {
			return fmt.Errorf("%s: cstime: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.priority)
		if err != nil {
			return fmt.Errorf("%s: priority: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.nice)
		if err != nil {
			return fmt.Errorf("%s: nice: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.num_threads)
		if err != nil {
			return fmt.Errorf("%s: num_threads: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.itrealvalue)
		if err != nil {
			return fmt.Errorf("%s: itrealvalue: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &u64)
		if err != nil {
			return fmt.Errorf("%s: starttime: %v", fn, err)
		}
		t.starttime = startTimeString(now.Add(time.Second*
			-time.Duration(uint64(si.Uptime)-(u64/hz))),
			boy, bod)
		_, err = fmt.Fscanf(f, "%v", &t.vsize)
		if err != nil {
			return fmt.Errorf("%s: vsize: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.rss)
		if err != nil {
			return fmt.Errorf("%s: rss: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.rsslim)
		if err != nil {
			return fmt.Errorf("%s: rsslim: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.startcode)
		if err != nil {
			return fmt.Errorf("%s: startcode: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.endcode)
		if err != nil {
			return fmt.Errorf("%s: endcode: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.startstack)
		if err != nil {
			return fmt.Errorf("%s: startstack: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.kstkesp)
		if err != nil {
			return fmt.Errorf("%s: kstkesp: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.kstkeip)
		if err != nil {
			return fmt.Errorf("%s: kstkeip: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.signal)
		if err != nil {
			return fmt.Errorf("%s: signal: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.blocked)
		if err != nil {
			return fmt.Errorf("%s: blocked: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.sigignore)
		if err != nil {
			return fmt.Errorf("%s: sigignore: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.sigcatch)
		if err != nil {
			return fmt.Errorf("%s: sigcatch: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.wchan)
		if err != nil {
			return fmt.Errorf("%s: wchan: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.nswap)
		if err != nil {
			return fmt.Errorf("%s: nswap: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.cnswap)
		if err != nil {
			return fmt.Errorf("%s: cnswap: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.exit_signal)
		if err != nil {
			return fmt.Errorf("%s: exit_signal: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.processor)
		if err != nil {
			return fmt.Errorf("%s: processor: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.rt_priority)
		if err != nil {
			return fmt.Errorf("%s: rt_priority: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.policy)
		if err != nil {
			return fmt.Errorf("%s: policy: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.delayacct_blkio_ticks)
		if err != nil {
			return fmt.Errorf("%s: delayacct_blkio_ticks: %v",
				fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.guest_time)
		if err != nil {
			return fmt.Errorf("%s: guest_time: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.cguest_time)
		if err != nil {
			return fmt.Errorf("%s: cguest_time: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.start_data)
		if err != nil {
			return fmt.Errorf("%s: start_data: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.end_data)
		if err != nil {
			return fmt.Errorf("%s: end_data: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.start_brk)
		if err != nil {
			return fmt.Errorf("%s: start_brk: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.arg_start)
		if err != nil {
			return fmt.Errorf("%s: arg_start: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.arg_end)
		if err != nil {
			return fmt.Errorf("%s: arg_end: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.env_start)
		if err != nil {
			return fmt.Errorf("%s: env_start: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.env_end)
		if err != nil {
			return fmt.Errorf("%s: env_end: %v", fn, err)
		}
		_, err = fmt.Fscanf(f, "%v", &t.exit_code)
		if err != nil {
			return fmt.Errorf("%s: exit_code: %v", fn, err)
		}

		f.Close()
		stats = append(stats, t)
		if t.pid == thisPid {
			thisStat = t
		}
	}
	sort.Sort(&stats)
	switch {
	case flag["-f"]:
		fmt.Printf("%5s %8s %6s %6s %5s %5s %16s %s\n",
			"UID", "TTY", "PID", "PPID", "STATE", "START", "TIME",
			"COMMAND")
	default:
		fmt.Printf("%6s %8s %16s %s\n",
			"PID", "TTY", "TIME", "COMMAND")
	}
	lastTty := uint(0)
	thisTtyName := "?"
	for _, t := range stats {
		if t.tty_nr != lastTty {
			thisTtyName, err = ttyName(t.tty_nr)
			if err != nil {
				return err
			}
			lastTty = t.tty_nr
		}
		if flag["-e"] || t.tty_nr == thisStat.tty_nr {
			switch {
			case flag["-f"]:
				cmdline := t.comm
				fn := fmt.Sprintf("/proc/%d/cmdline", t.pid)
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
					t.uid,
					thisTtyName,
					t.pid,
					t.ppid,
					"  "+t.state+"  ",
					t.starttime,
					t.utime,
					cmdline)
			default:
				fmt.Printf("%6d %8s %16s %s\n",
					t.pid,
					thisTtyName,
					t.utime,
					t.comm)
			}
		}
	}
	return nil
}

func (cmd) Man() lang.Alt  { return man }
func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

type psStats []*psStat
type psStat struct {
	uid     uint32
	gid     uint32
	pid     int    //  %d
	comm    string //  %s
	state   string //  %c
	ppid    int    //  %d
	pgrp    int    //  %d
	session int    //  %d
	tty_nr  uint   //  %d
	tpgid   int    //  %d
	flags   uint   //  %u
	minflt  uint64 //  %lu
	cminflt uint64 //  %lu
	majflt  uint64 //  %lu
	cmajflt uint64 //  %lu

	utime time.Duration //  %lu

	stime       uint64 //  %lu
	cutime      uint64 //  %ld
	cstime      uint64 //  %ld
	priority    int64  //  %ld
	nice        int64  //  %ld
	num_threads int64  //  %ld
	itrealvalue int64  //  %ld

	starttime string //  %llu

	vsize       uint64 //  %lu
	rss         int64  //  %ld
	rsslim      uint64 //  %lu
	startcode   uint64 //  %lu
	endcode     uint64 //  %lu
	startstack  uint64 //  %lu
	kstkesp     uint64 //  %lu
	kstkeip     uint64 //  %lu
	signal      uint64 //  %lu
	blocked     uint64 //  %lu
	sigignore   uint64 //  %lu
	sigcatch    uint64 //  %lu
	wchan       uint64 //  %lu
	nswap       uint64 //  %lu
	cnswap      uint64 //  %lu
	exit_signal int    //  %d  (since Linux 2.1.22)
	processor   int    //  %d  (since Linux 2.2.8)
	rt_priority uint   //  %u  (since Linux 2.5.19)
	policy      uint   //  %u  (since Linux 2.5.19)

	delayacct_blkio_ticks uint64 //  %llu  (since Linux 2.6.18)

	guest_time  uint64 //  %lu  (since Linux 2.6.24)
	cguest_time uint64 //  %ld  (since Linux 2.6.24)
	start_data  uint64 //  %lu  (since Linux 3.3)
	end_data    uint64 //  %lu  (since Linux 3.3)
	start_brk   uint64 //  %lu  (since Linux 3.3)
	arg_start   uint64 //  %lu  (since Linux 3.5)
	arg_end     uint64 //  %lu  (since Linux 3.5)
	env_start   uint64 //  %lu  (since Linux 3.5)
	env_end     uint64 //  %lu  (since Linux 3.5)
	exit_code   int    //  %d  (since Linux 3.5)
}

var epoch time.Time
var hz uint64

func ttyName(tty_nr uint) (string, error) {
	const devname = "DEVNAME="
	name := "?"
	maj := major(tty_nr)
	min := minor(tty_nr)
	fn := fmt.Sprintf("/sys/dev/char/%d:%d/uevent", maj, min)
	if f, err := os.Open(fn); err == nil {
		sc := bufio.NewScanner(f)
		if !sc.Scan() {
			return "", fmt.Errorf("%s: missing MAJOR=#", fn)
		}
		if !sc.Scan() {
			return "", fmt.Errorf("%s: missing MINOR=#", fn)
		}
		if !sc.Scan() {
			return "", fmt.Errorf("%s: missing DEVNAME=#", fn)
		}
		s := sc.Text()
		if !strings.HasPrefix(s, devname) {
			return "", fmt.Errorf("%s: has %s; expected %s",
				fn, s, devname)
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
	return name, nil

}

func duration(t uint64) time.Duration {
	return unixTime(t).Sub(epoch)
}

func unixTime(t uint64) time.Time {
	return time.Unix(int64(t/hz), int64(t%hz))
}

func major(tty_nr uint) uint {
	return (tty_nr >> 8) & 0xfff
}

func minor(tty_nr uint) uint {
	return (tty_nr & 0xff) | ((tty_nr >> 12) & 0xfff00)
}

func (p *psStats) Len() int {
	return len(*p)
}

func (p *psStats) Less(i, j int) bool {
	if (*p)[i].uid == (*p)[j].uid {
		if (*p)[i].tty_nr == (*p)[j].tty_nr {
			if (*p)[i].pid < (*p)[j].pid {
				return true
			}
		} else if (*p)[i].tty_nr < (*p)[j].tty_nr {
			return true
		}
	} else if (*p)[i].uid < (*p)[j].uid {
		return true
	}
	return false
}

func (p *psStat) modComm() {
	if p.pid == 2 || p.ppid == 2 {
		// replace enclosing parentheses of kernel threads
		p.comm = "[" + p.comm[1:len(p.comm)-1] + "]"
	} else {
		// trim enclosing parentheses of non-kernel threads
		p.comm = p.comm[1 : len(p.comm)-1]
	}
	if p.comm == "goes" || p.comm == "exe" {
		fn := fmt.Sprintf("/proc/%d/cmdline", p.pid)
		buf, err := ioutil.ReadFile(fn)
		if err == nil {
			for i, c := range buf {
				if c == 0 {
					p.comm = string(buf[:i])
					break
				}
			}
		}
	}
}

func (p *psStats) Swap(i, j int) {
	t := (*p)[i]
	(*p)[i] = (*p)[j]
	(*p)[j] = t
}

func startTimeString(start, boy, bod time.Time) (s string) {
	if start.Before(boy) {
		s = fmt.Sprintf("%4d", start.Year())
	} else if start.Before(bod) {
		s = fmt.Sprintf("%s%02d", start.Month().String()[:3],
			start.Day())
	} else {
		s = fmt.Sprintf("%2d:%02d", start.Hour(), start.Minute())
	}
	return s
}

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)
