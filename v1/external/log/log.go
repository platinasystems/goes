// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
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
	stop chan<- struct{}
}

func (p *RateLimited) Close() error {
	close(p.stop)
	return nil
}

type teeT struct {
	sync.Mutex
	w io.Writer

	exclusive bool
}

type earlyT struct {
	sync.Mutex
	buf *bytes.Buffer
}

var (
	tee   teeT
	early = earlyT{buf: &bytes.Buffer{}}
)

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

// NewLimited returns a logger with the given iteration restriction.
func NewLimited(n uint32) *Limited { return &Limited{N: n} }

// NewRateLimited returns a logger restricted to the given iterations per
// unit time that should be destroyed with `defer (*RateLimited).Close()`
func NewRateLimited(n uint32, d time.Duration) *RateLimited {
	stop := make(chan struct{})
	rl := &RateLimited{NewLimited(n), stop}
	go func(wait <-chan struct{}) {
		t := time.NewTicker(d)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				rl.reset()
			case <-wait:
				return
			}
		}
	}(stop)
	return rl
}

// Tee logged lines to Writer
func Tee(w io.Writer) { tee.w = w }

// log lines from the given reader until EOF or error.
func LinesFrom(rc io.ReadCloser, id, priority string) {
	defer rc.Close()
	pri, found := PriorityByName[priority]
	if !found {
		pri = syslog.LOG_ERR
	}
	scan := bufio.NewScanner(rc)
	for scan.Scan() {
		log(pri|syslog.LOG_DAEMON, id, scan.Text())
	}
}

// The default level is: Debug, User. Upto the first two arguments may change
// this by name; e.g.
//
//	Print("daemon", ...)
//	Print("daemon", "err", ...)
//	Print("err", ...)
func Print(args ...interface{}) {
	pri, fac, a := logArgs(args...)
	log(pri|fac, id(), fmt.Sprint(a...))
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
	log(pri|fac, id(), fmt.Sprintf(format, a...))
}

var cache struct {
	once sync.Once
	id   string
	pid  int
}

func id() string {
	cache.once.Do(func() {
		var prog string
		s, err := os.Readlink("/proc/self/exe")
		if err == nil {
			prog = filepath.Base(s)
			if s != os.Args[0] {
				if strings.HasPrefix(os.Args[0], prog) {
					prog = os.Args[0]
				} else {
					prog += "." + os.Args[0]
				}
			}
		} else {
			prog = filepath.Base(os.Args[0])
		}
		if cache.pid == 0 {
			cache.pid = os.Getpid()
		}
		cache.id = fmt.Sprintf("%s[%d]", prog, cache.pid)
	})
	return cache.id
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

func log(pri syslog.Priority, id string, args ...interface{}) {
	lines := strings.Split(fmt.Sprint(args...), "\n")
	if tee.w != nil {
		tee.log(pri, id, lines)
		if tee.exclusive {
			return
		}
	}
	if _, err := os.Stat(DevLog); err == nil {
		conn, err := net.Dial("unixgram", DevLog)
		if err != nil {
			// FIXME how to log a log error?
			return
		}
		defer conn.Close()
		for _, s := range lines {
			fmt.Fprintf(conn, "<%d>%s %s: %s\n",
				pri, time.Now().Format(time.Stamp),
				id, s)
		}
	} else if k, err := os.OpenFile(DevKmsg, os.O_RDWR, 0644); err == nil {
		defer k.Close()
		early.flush(k)
		for _, s := range lines {
			fmt.Fprintf(k, "<%d>%s: %s\n", pri, id, s)
		}
	} else if os.IsNotExist(err) {
		early.log(pri, id, lines)
	}
}

func (p *teeT) log(pri syslog.Priority, id string, lines []string) {
	p.Lock()
	defer p.Unlock()
	for _, s := range lines {
		fmt.Fprintf(p.w, "<%d>%s: %s\n", pri, id, s)
	}
}

func (p *earlyT) log(pri syslog.Priority, id string, lines []string) {
	p.Lock()
	defer p.Unlock()
	for _, s := range lines {
		fmt.Fprintf(p.buf, "<%d>%s: %s\n", pri, id, s)
	}
}

func (p *earlyT) flush(w io.Writer) {
	p.Lock()
	defer p.Unlock()
	if p.buf.Len() == 0 {
		return
	}
	w.Write(p.buf.Bytes())
	p.buf.Reset()
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
	if u, err := strconv.ParseUint(string(field[2]), 10, 64); err == nil {
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
