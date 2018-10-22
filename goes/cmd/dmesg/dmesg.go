// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package dmesg

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/log"
	"github.com/platinasystems/go/internal/parms"
)

const (
	SYSLOG_ACTION_CLOSE = iota
	SYSLOG_ACTION_OPEN
	SYSLOG_ACTION_READ
	SYSLOG_ACTION_READ_ALL
	SYSLOG_ACTION_READ_CLEAR
	SYSLOG_ACTION_CLEAR
	SYSLOG_ACTION_CONSOLE_OFF
	SYSLOG_ACTION_CONSOLE_ON
	SYSLOG_ACTION_CONSOLE_LEVEL
	SYSLOG_ACTION_SIZE_UNREAD
	SYSLOG_ACTION_SIZE_BUFFER
)

type Command struct{}

func (Command) String() string { return "dmesg" }

func (Command) Usage() string { return "dmesg [OPTION]..." }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "print or control the kernel ring buffer",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	The default action is to print new kernel ring buffer messages since
	the last command invocation.

OPTIONS
	-C	Clear the ring buffer.
	-c	Clear the ring buffer after first printing its contents.
	-D	Disable the printing of messages to the console.
	-d	Display the delta time between messages.
	-E	Enable printing messages to the console.
	-F	Read the messages from the given file instead of /dev/kmsg.
	-H	Enable human-readable output.
	-k	Print kernel messages.
	-n level
		Set console to the given numbered or named log level.
	-r	Print the raw message, i.e. do not strip the priority prefix.
	-T	Print human-readable timestamps.
	-t	Do not print timestamps.
	-u	Print userspace messages.
	-w	Wait for new messages.
	-x	Decode facility and level (priority) numbers.
	-z	Reprint entire ring buffer.`,
	}
}

func (Command) Main(args ...string) error {
	const (
		nl = "\n"
		sp = " "
		lt = "<"
		gt = ">"
		lb = "["
		rb = "]"

		MaxEpollEvents = 32
	)
	var last, kmsg log.Kmsg
	var event syscall.EpollEvent
	var events [MaxEpollEvents]syscall.EpollEvent

	flag, args := flags.New(args, "-C", "-c", "-D", "-d", "-E", "-H",
		"-k", "-r", "-T", "-t", "-u", "-x", "-z")
	parm, args := parms.New(args, "-F", "-n")
	if len(parm.ByName["-F"]) == 0 {
		parm.ByName["-F"] = "/dev/kmsg"
	}

	if len(args) > 0 {
		fmt.Errorf("%v: unexpected", args)
	}

	f, err := os.Open(parm.ByName["-F"])
	if err != nil {
		return err
	}
	defer f.Close()

	fd := int(f.Fd())
	if err = syscall.SetNonblock(fd, true); err != nil {
		return err
	}
	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		return err
	}
	defer syscall.Close(epfd)

	buf := make([]byte, 4096)
	defer func() { buf = buf[:0] }()

	if flag.ByName["-C"] {
		_, err = syscall.Klogctl(SYSLOG_ACTION_CLEAR, buf)
		return err
	}
	if flag.ByName["-D"] {
		_, err = syscall.Klogctl(SYSLOG_ACTION_CONSOLE_OFF, buf)
		return err
	}
	if flag.ByName["-E"] {
		_, err = syscall.Klogctl(SYSLOG_ACTION_CONSOLE_ON, buf)
		return err
	}
	if s := parm.ByName["-n"]; len(s) > 0 {
		pri, found := log.PriorityByName[s]
		if !found {
			fmt.Errorf("%s: unknown", s)
		}
		_, err = syscall.Klogctl(SYSLOG_ACTION_CONSOLE_LEVEL,
			buf[:pri])
		return err
	}
	if flag.ByName["-z"] {
		last.Seq = 0
	}

	now := time.Now()
	var si syscall.Sysinfo_t
	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	event.Events = syscall.EPOLLIN
	event.Fd = int32(fd)
	err = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd, &event)
	if err != nil {
		return err
	}
	nevents, err := syscall.EpollWait(epfd, events[:], -1)
	if err != nil {
		return err
	}
	for ev := 0; ev < nevents; ev++ {
		for {
			n, err := syscall.Read(int(events[ev].Fd), buf[:])
			if err != nil {
				break
			}
			if n > 0 {
				kmsg.Parse(buf[:n])

				if kmsg.Stamp == log.Stamp(0) ||
					(flag.ByName["-k"] && !kmsg.IsKern()) ||
					(flag.ByName["-u"] && kmsg.IsKern()) {
					continue
				}

				if last.Stamp == log.Stamp(0) {
					last.Stamp = kmsg.Stamp
				}
				delta := kmsg.Stamp.Delta(last.Stamp)
				t := timeT(kmsg.Stamp.Time(now, int64(si.Uptime)))

				if last.Seq == 0 || last.Seq < kmsg.Seq {
					fac := kmsg.Pri & log.FacilityMask
					pri := kmsg.Pri & log.PriorityMask
					xs := fmt.Sprintf("%-8s%-8s",
						log.LogFacilityByValue[fac]+":",
						log.LogPriorityByValue[pri]+":")
					switch {
					case flag.ByName["-H"] &&
						flag.ByName["-d"] &&
						flag.ByName["-x"]:
						fmt.Print(xs,
							lb, t.H(), sp, lt, delta, gt, rb,
							sp, kmsg.Msg, nl)
					case flag.ByName["-H"] && flag.ByName["-d"]:
						fmt.Print(lb, t.H(), sp, lt, delta, gt, rb,
							sp, kmsg.Msg, nl)
					case flag.ByName["-H"] && flag.ByName["-x"]:
						fmt.Print(xs, sp,
							lb, t.H(), rb,
							sp, kmsg.Msg, nl)
					case flag.ByName["-H"]:
						fmt.Print(lb, t.H(), rb,
							sp, kmsg.Msg, nl)
					case flag.ByName["-T"] &&
						flag.ByName["-d"] &&
						flag.ByName["-x"]:
						fmt.Print(xs,
							lb, t.T(), sp, lt, delta, gt, rb,
							sp, kmsg.Msg, nl)
					case flag.ByName["-T"] && flag.ByName["-d"]:
						fmt.Print(lb, t.T(), sp, lt, delta, gt, rb,
							sp, kmsg.Msg, nl)
					case flag.ByName["-T"] && flag.ByName["-x"]:
						fmt.Print(xs, sp,
							lb, t.T(), rb,
							sp, kmsg.Msg, nl)
					case flag.ByName["-T"]:
						fmt.Print(lb, t.T(), rb,
							sp, kmsg.Msg, nl)
					case flag.ByName["-t"] &&
						flag.ByName["-d"] &&
						flag.ByName["-x"]:
						fmt.Print(xs, sp,
							lb, lt, delta, gt, rb,
							sp, kmsg.Msg, nl)
					case flag.ByName["-t"] && flag.ByName["-d"]:
						fmt.Print(lb, lt, delta, gt, rb,
							sp, kmsg.Msg, nl)
					case flag.ByName["-t"] && flag.ByName["-x"]:
						fmt.Print(xs, sp, kmsg.Msg, nl)
					case flag.ByName["-t"]:
						fmt.Print(kmsg.Msg, nl)
					case flag.ByName["-r"]:
						fmt.Print(lt, kmsg.Pri, gt,
							lb, kmsg.Stamp, rb,
							sp, kmsg.Msg, nl)
					case flag.ByName["-d"] && flag.ByName["-x"]:
						fmt.Print(xs,
							lb, kmsg.Stamp, sp, lt, delta, gt, rb,
							sp, kmsg.Msg, nl)
					case flag.ByName["-d"]:
						fmt.Print(lb, kmsg.Stamp, sp, lt, delta, gt, rb,
							sp, kmsg.Msg, nl)
					case flag.ByName["-x"]:
						fmt.Print(xs, sp,
							lb, kmsg.Stamp, rb,
							sp, kmsg.Msg, nl)
					default:
						fmt.Print(lb, kmsg.Stamp, rb,
							sp, kmsg.Msg, nl)
					}
					last.Seq = kmsg.Seq
				}
			}
			last.Stamp = kmsg.Stamp
		}
	}

	if flag.ByName["-c"] {
		_, err = syscall.Klogctl(SYSLOG_ACTION_CLEAR, buf)
		if err != nil {
			return err
		}
	}
	return nil
}

type timeT time.Time

func (t timeT) H() string {
	return time.Time(t).Format(time.Stamp)
}

func (t timeT) T() string {
	return time.Time(t).Format("Mon " + time.Stamp + " 2006")
}
