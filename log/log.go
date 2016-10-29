// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package log prints messages to a given writer, /dev/log, /dev/kmsg, or a
// byte buffer until one of these are available.
package log

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log/syslog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/platinasystems/go/emptych"
)

const DevKmsg = "/dev/kmsg"
const DevLog = "/dev/log"
const PriorityMask = syslog.Priority(7)
const FacilityMask = ^PriorityMask

type Seq uint64
type Stamp uint64
type Delta uint64

type Kmsg struct {
	Pri   syslog.Priority
	Seq   Seq
	Stamp Stamp
	Msg   string
}

// Limited provides a logger with Print and Printf restricted to the given
// iterations.
type Limited struct {
	sync.Mutex
	i, N uint32
}

// RateLimited provides a logger with Print and Printf restricted to the given
// iterations per unit time. This must be created with NewRateLimited and
// destroyed with (*RateLimited).Close().
type RateLimited struct {
	*Limited
	emptych.In
}

var pid int64
var Writer io.Writer
var mutex sync.Mutex
var earlyBufs []*bytes.Buffer

var prog string

var PriorityByName = map[string]syslog.Priority{
	"emerg": syslog.LOG_EMERG,
	"alert": syslog.LOG_ALERT,
	"crit":  syslog.LOG_CRIT,
	"err":   syslog.LOG_ERR,
	"warn":  syslog.LOG_WARNING,
	"note":  syslog.LOG_NOTICE,
	"info":  syslog.LOG_INFO,
	"debug": syslog.LOG_DEBUG,
}

var LogPriorityByValue = map[syslog.Priority]string{
	syslog.LOG_EMERG:   "emerg",
	syslog.LOG_ALERT:   "alert",
	syslog.LOG_CRIT:    "crit",
	syslog.LOG_ERR:     "err",
	syslog.LOG_WARNING: "warn",
	syslog.LOG_NOTICE:  "note",
	syslog.LOG_INFO:    "info",
	syslog.LOG_DEBUG:   "debug",
}

var FacilityByName = map[string]syslog.Priority{
	"kern":   syslog.LOG_KERN,
	"user":   syslog.LOG_USER,
	"mail":   syslog.LOG_MAIL,
	"daemon": syslog.LOG_DAEMON,
	"auth":   syslog.LOG_AUTH,
	"syslog": syslog.LOG_SYSLOG,
	"lpr":    syslog.LOG_LPR,
	"news":   syslog.LOG_NEWS,
	"uucp":   syslog.LOG_UUCP,
	"cron":   syslog.LOG_CRON,
	"priv":   syslog.LOG_AUTHPRIV,
	"ftp":    syslog.LOG_FTP,
	"local0": syslog.LOG_LOCAL0,
	"local1": syslog.LOG_LOCAL1,
	"local2": syslog.LOG_LOCAL2,
	"local3": syslog.LOG_LOCAL3,
	"local4": syslog.LOG_LOCAL4,
	"local5": syslog.LOG_LOCAL5,
	"local6": syslog.LOG_LOCAL6,
	"local7": syslog.LOG_LOCAL7,
}

var LogFacilityByValue = map[syslog.Priority]string{
	syslog.LOG_KERN:     "kern",
	syslog.LOG_USER:     "user",
	syslog.LOG_MAIL:     "mail",
	syslog.LOG_DAEMON:   "daemon",
	syslog.LOG_AUTH:     "auth",
	syslog.LOG_SYSLOG:   "syslog",
	syslog.LOG_LPR:      "lpr",
	syslog.LOG_NEWS:     "news",
	syslog.LOG_UUCP:     "uucp",
	syslog.LOG_CRON:     "cron",
	syslog.LOG_AUTHPRIV: "priv",
	syslog.LOG_FTP:      "ftp",
	syslog.LOG_LOCAL0:   "local0",
	syslog.LOG_LOCAL1:   "local1",
	syslog.LOG_LOCAL2:   "local2",
	syslog.LOG_LOCAL3:   "local3",
	syslog.LOG_LOCAL4:   "local4",
	syslog.LOG_LOCAL5:   "local5",
	syslog.LOG_LOCAL6:   "local6",
	syslog.LOG_LOCAL7:   "local7",
}

func DaemonErr(args ...interface{}) {
	log(syslog.LOG_DAEMON|syslog.LOG_ERR, fmt.Sprint(args...))
}

func DaemonInfo(args ...interface{}) {
	log(syslog.LOG_DAEMON|syslog.LOG_INFO, fmt.Sprint(args...))
}

// NewLimited returns a logger with the given iteration restriction.
func NewLimited(n uint32) *Limited { return &Limited{N: n} }

// NewRateLimited returns a logger restricted to the given iterations per
// unit time that should be destroyed with `defer (*RateLimited).Close()`
func NewRateLimited(n uint32, d time.Duration) *RateLimited {
	done := emptych.Make()
	rl := &RateLimited{NewLimited(n), emptych.In(done)}
	go func(done emptych.Out) {
		t := time.NewTicker(d)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				rl.reset()
			case <-done:
				return
			}
		}
	}(emptych.Out(done))
	return rl
}

// Pipe returns a *File after starting a go routine that loops logging the
// other end of the pipe until EOF; then signals complete by closing the
// returned channel..
func Pipe(priority string) (w *os.File, done <-chan struct{}, err error) {
	pri, found := PriorityByName[priority]
	if !found {
		pri = syslog.LOG_ERR
	}
	r, w, err := os.Pipe()
	if err == nil {
		ch := make(chan struct{})
		done = ch
		go func(pri syslog.Priority, r io.Reader,
			done chan<- struct{}) {
			scan := bufio.NewScanner(r)
			for scan.Scan() {
				s := scan.Text()
				s = strings.Replace(s, "\t", "        ", -1)
				log(pri|syslog.LOG_DAEMON, s)
			}
			close(done)
			log(pri|syslog.LOG_DAEMON, "closed pipe")
		}(pri, r, ch)
	}
	return
}

// The default level is: Debug, User. Upto the first two arguments may change
// this by name; for example:
//
//	Print("daemon", ...)
//	Print("daemon", "err", ...)
//	Print("err", ...)
func Print(args ...interface{}) {
	pri, fac, a := logArgs(args...)
	log(pri|fac, fmt.Sprint(a...))
}

// The default level is: Debug, User. Upto the first two arguments may preceed
// the log format string to change the priority and facility like this:
//
//	Printf("daemon", format, ...)
//	Printf("daemon", "err", format, ...)
//	Printf("err", format, ...)
func Printf(args ...interface{}) {
	pri, fac, a := logArgs(args...)
	if len(a) <= 0 {
		// missing format
		return
	}
	format, ok := a[0].(string)
	if !ok {
		// a[0]: isn't string
		return
	}
	a = a[1:]
	log(pri|fac, fmt.Sprintf(format, a...))
}

func logArgs(args ...interface{}) (pri, fac syslog.Priority, a []interface{}) {
	pri = syslog.LOG_DEBUG
	fac = syslog.LOG_USER
	a = args
	for i := 0; len(a) > 0 && i < 2; i++ {
		s, ok := a[0].(string)
		if !ok {
			break
		}
		if v, found := PriorityByName[s]; found {
			pri = v
			a = a[1:]
			continue
		}
		if v, found := FacilityByName[s]; found {
			fac = v
			a = a[1:]
		}
	}
	return
}

func log(pri syslog.Priority, args ...interface{}) {
	mutex.Lock()
	defer mutex.Unlock()

	if pid == 0 {
		pid = int64(os.Getpid())
	}

	if len(prog) == 0 {
		s, err := os.Readlink("/proc/self/exe")
		if err == nil {
			prog = filepath.Base(s)
			if s != os.Args[0] {
				prog += "." + os.Args[0]
			}
		} else {
			prog = filepath.Base(os.Args[0])
		}
	}

	msg := strings.Split(fmt.Sprint(args...), "\n")

	if Writer != nil {
		for _, s := range msg {
			fmt.Fprintf(Writer, "<%d>%s[%d]: %s\n", pri, prog,
				pid, s)
		}
	} else if _, err := os.Stat(DevLog); err == nil {
		conn, err := net.Dial("unixgram", DevLog)
		if err != nil {
			// FIXME how to log a log error?
			return
		}
		defer conn.Close()
		for _, s := range msg {
			fmt.Fprintf(conn, "<%d>%s %s[%d]: %s\n",
				pri, time.Now().Format(time.Stamp),
				prog, pid, s)
		}
	} else if kmsg, err := os.OpenFile(DevKmsg, os.O_RDWR, 0644); err == nil {
		defer kmsg.Close()
		if len(earlyBufs) > 0 {
			for _, buf := range earlyBufs {
				kmsg.Write(buf.Bytes())
				buf.Reset()
			}
			earlyBufs = earlyBufs[:0]
		}
		for _, s := range msg {
			fmt.Fprintf(kmsg, "<%d>%s[%d]: %s\n", pri, prog,
				pid, s)
		}
	} else if os.IsNotExist(err) {
		buf := new(bytes.Buffer)
		for _, s := range msg {
			fmt.Fprintf(buf, "<%d>%s[%d]: %s\n", pri, prog,
				pid, s)
		}
		earlyBufs = append(earlyBufs, buf)
	}
}

func (p *Kmsg) IsKern() bool {
	return p.Pri&FacilityMask == syslog.LOG_KERN
}

func (p *Kmsg) Parse(b []byte) {
	sc := bytes.Index(b, []byte{';'})
	if sc < 0 {
		return
	}
	end := bytes.Index(b, []byte{'\n'})
	if end < 0 {
		end = len(b)
	}
	p.Msg = string(b[sc+1 : end])
	field := bytes.FieldsFunc(b[:sc],
		func(r rune) bool { return r == ',' })
	if u, err := strconv.ParseUint(string(field[0]), 0, 0); err == nil {
		p.Pri = syslog.Priority(u)
	} else {
		p.Pri = syslog.Priority(0)
	}
	if u, err := strconv.ParseUint(string(field[1]), 0, 0); err == nil {
		p.Seq = Seq(u)
	} else {
		p.Seq = Seq(0)
	}
	if u, err := strconv.ParseUint(string(field[2]), 0, 0); err == nil {
		p.Stamp = Stamp(u)
	} else {
		p.Stamp = Stamp(0)
	}
}

func (delta Delta) Sec() uint64 {
	return uint64(delta / 1000000)
}

func (delta Delta) Usec() uint64 {
	return uint64(delta % 1000000)
}

func (delta Delta) String() string {
	return fmt.Sprintf("%5d.%06d", delta.Sec(), delta.Usec())
}

func (stamp Stamp) Delta(last Stamp) Delta {
	return Delta(stamp - last)
}
func (stamp Stamp) Sec() uint64 {
	return uint64(stamp / 1000000)
}

func (stamp Stamp) Usec() uint64 {
	return uint64(stamp % 1000000)
}

func (stamp Stamp) String() string {
	return fmt.Sprintf("%07d.%06d", stamp.Sec(), stamp.Usec())
}

func (stamp Stamp) Time(now time.Time, uptime int64) time.Time {
	dur := time.Duration(uint64(uptime) - stamp.Sec())
	return now.Add(time.Second * -dur)
}

func (l *Limited) limited(f func(...interface{}), args ...interface{}) {
	if atomic.LoadUint32(&l.i) == l.N {
		return
	}
	l.Lock()
	defer l.Unlock()
	if l.i < l.N {
		defer atomic.AddUint32(&l.i, 1)
		f(args...)
	}
}

func (l *Limited) reset() {
	l.Lock()
	defer l.Unlock()
	defer atomic.StoreUint32(&l.i, 0)
}

func (l *Limited) Print(args ...interface{}) {
	l.limited(Print, args...)
}

func (l *Limited) Printf(args ...interface{}) {
	l.limited(Printf, args...)
}
