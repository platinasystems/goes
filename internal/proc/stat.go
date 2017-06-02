// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package proc

import (
	"fmt"
	"io"
	"io/ioutil"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/sysconf"
)

// Linux /proc/<PID|"self">/stat
type Stat struct {
	Pid     int    //  %d
	Comm    string //  %s
	State   string //  %c
	Ppid    int    //  %d
	Pgrp    int    //  %d
	Session int    //  %d
	TtyNr   uint   //  %d
	Tpgid   int    //  %d
	Flags   uint   //  %u
	MinFlt  uint64 //  %lu
	CminFlt uint64 //  %lu
	MajFlt  uint64 //  %lu
	CmajFlt uint64 //  %lu

	Utime time.Duration //  %lu

	Stime       uint64 //  %lu
	Cutime      uint64 //  %ld
	Cstime      uint64 //  %ld
	Priority    int64  //  %ld
	Nice        int64  //  %ld
	NumThreads  int64  //  %ld
	ItRealValue int64  //  %ld

	StartTime string //  %llu

	Vsize      uint64 //  %lu
	Rss        int64  //  %ld
	RssLim     uint64 //  %lu
	StartCode  uint64 //  %lu
	EndCode    uint64 //  %lu
	StartStack uint64 //  %lu
	KstkESP    uint64 //  %lu
	KstkEIP    uint64 //  %lu
	Signal     uint64 //  %lu
	Blocked    uint64 //  %lu
	SigIgnore  uint64 //  %lu
	SigCatch   uint64 //  %lu
	Wchan      uint64 //  %lu
	Nswap      uint64 //  %lu
	Cnswap     uint64 //  %lu
	ExitSignal int    //  %d  (since Linux 2.1.22)
	Processor  int    //  %d  (since Linux 2.2.8)
	RtPriority uint   //  %u  (since Linux 2.5.19)
	Policy     uint   //  %u  (since Linux 2.5.19)

	DelayAcctBlkioTicks uint64 //  %llu  (since Linux 2.6.18)

	GuestTime  uint64 //  %lu  (since Linux 2.6.24)
	CguestTime uint64 //  %ld  (since Linux 2.6.24)
	StartData  uint64 //  %lu  (since Linux 3.3)
	EndData    uint64 //  %lu  (since Linux 3.3)
	StartBrk   uint64 //  %lu  (since Linux 3.3)
	ArgStart   uint64 //  %lu  (since Linux 3.5)
	ArgEnd     uint64 //  %lu  (since Linux 3.5)
	EnvStart   uint64 //  %lu  (since Linux 3.5)
	EnvEnd     uint64 //  %lu  (since Linux 3.5)
	ExitCode   int    //  %d  (since Linux 3.5)
}

var hz uint64

func (p *Stat) ReadFrom(r io.Reader) error {
	var utime, starttime uint64
	var si syscall.Sysinfo_t

	err := syscall.Sysinfo(&si)
	if err != nil {
		return err
	}

	if hz == 0 {
		hz = sysconf.Hz()
	}

	now := time.Now()
	boy := time.Date(now.Year(), 1, 1, 12, 0, 0, 0, now.Location())
	bod := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0,
		now.Location())

	for _, x := range []struct {
		s string
		v interface{}
	}{
		{"Pid", &p.Pid},
		{"Comm", &p.Comm},
		{"State", &p.State},
		{"Ppid", &p.Ppid},
		{"Pgrp", &p.Pgrp},
		{"Session", &p.Session},
		{"TtyNr", &p.TtyNr},
		{"Tpgid", &p.Tpgid},
		{"Flags", &p.Flags},
		{"MinFlt", &p.MinFlt},
		{"CminFlt", &p.CminFlt},
		{"MajFlt", &p.MajFlt},
		{"CmajFlt", &p.CmajFlt},
		{"Utime", &utime},
		{"Stime", &p.Stime},
		{"Cutime", &p.Cutime},
		{"Cstime", &p.Cstime},
		{"Priority", &p.Priority},
		{"Nice", &p.Nice},
		{"NumThreads", &p.NumThreads},
		{"ItRealValue", &p.ItRealValue},
		{"StartTime", &starttime},
		{"Vsize", &p.Vsize},
		{"Rss", &p.Rss},
		{"RssLim", &p.RssLim},
		{"StartCode", &p.StartCode},
		{"EndCode", &p.EndCode},
		{"StartStack", &p.StartStack},
		{"KstkESP", &p.KstkESP},
		{"KstkEIP", &p.KstkEIP},
		{"Signal", &p.Signal},
		{"Blocked", &p.Blocked},
		{"SigIgnore", &p.SigIgnore},
		{"SigCatch", &p.SigCatch},
		{"Wchan", &p.Wchan},
		{"Nswap", &p.Nswap},
		{"Cnswap", &p.Cnswap},
		{"ExitSignal", &p.ExitSignal},
		{"Processor", &p.Processor},
		{"RtPriority", &p.RtPriority},
		{"Policy", &p.Policy},
		{"DelayAcctBlkioTicks", &p.DelayAcctBlkioTicks},
		{"GuestTime", &p.GuestTime},
		{"CguestTime", &p.CguestTime},
		{"StartData", &p.StartData},
		{"EndData", &p.EndData},
		{"StartBrk", &p.StartBrk},
		{"ArgStart", &p.ArgStart},
		{"ArgEnd", &p.ArgEnd},
		{"EnvStart", &p.EnvStart},
		{"EnvEnd", &p.EnvEnd},
		{"ExitCode", &p.ExitCode},
	} {
		if _, err = fmt.Fscanf(r, "%v", x.v); err != nil {
			return fmt.Errorf("%s: %v", x.s, err)
		}
	}

	epoch := time.Unix(0, 0)
	p.Utime = time.Unix(int64(utime/hz), int64(utime%hz)).Sub(epoch)

	t := now.Add(time.Second *
		-time.Duration(uint64(si.Uptime)-(starttime/hz)))
	if t.Before(boy) {
		p.StartTime = fmt.Sprintf("%4d", t.Year())
	} else if t.Before(bod) {
		p.StartTime = fmt.Sprintf("%s%02d",
			t.Month().String()[:3], t.Day())
	} else {
		p.StartTime = fmt.Sprintf("%2d:%02d",
			t.Hour(), t.Minute())
	}

	if p.Pid == 2 || p.Ppid == 2 {
		// replace enclosing parentheses of kernel threads
		p.Comm = "[" + p.Comm[1:len(p.Comm)-1] + "]"
	} else {
		// trim enclosing parentheses of non-kernel threads
		p.Comm = p.Comm[1 : len(p.Comm)-1]
	}
	if p.Comm == "goes" || p.Comm == "exe" {
		fn := fmt.Sprintf("/proc/%d/cmdline", p.Pid)
		buf, err := ioutil.ReadFile(fn)
		if err == nil {
			for i, c := range buf {
				if c == 0 {
					p.Comm = string(buf[:i])
					break
				}
			}
		}
	}

	return nil
}
