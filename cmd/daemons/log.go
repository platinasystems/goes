// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package daemons

import (
	"bytes"
	"fmt"
	"sync"
	"time"
)

const (
	logEntries = 128
	logCap     = 160
)

type logEntry struct {
	sync.Mutex
	t time.Time
	b []byte
}

type daemonLog struct {
	r []logEntry
	i int
}

func (dl *daemonLog) init() {
	dl.r = make([]logEntry, logEntries)
	for i := range dl.r {
		dl.r[i].b = make([]byte, 0, logCap)
	}
}

func (dl *daemonLog) String() string {
	buf := new(bytes.Buffer)
	cat := func(l *logEntry) {
		fmt.Fprint(buf, l.t.Format(time.Stamp), " ")
		buf.Write(l.b)
	}
	for i := dl.i; i < len(dl.r); i++ {
		l := &dl.r[i]
		if l.t.IsZero() || len(l.b) == 0 {
			break
		}
		cat(l)
	}
	for i := 0; i < dl.i; i++ {
		cat(&dl.r[i])
	}
	return buf.String()
}

func (dl *daemonLog) Write(b []byte) (int, error) {
	const ellipsis = "...\n"
	l := &dl.r[dl.i]
	l.t = time.Now()
	if len(b) > 4 && b[0] == '<' && b[3] == '>' {
		// skip log priority prefix
		b = b[4:]
	}
	if len(b) > cap(l.b) {
		l.b = l.b[:cap(l.b)]
		n := cap(l.b) - len(ellipsis)
		copy(l.b, b[:n])
		copy(l.b[n:], ellipsis)
	} else {
		l.b = l.b[:len(b)]
		copy(l.b, b)
	}
	if dl.i++; dl.i >= logEntries {
		dl.i = 0
	}
	return len(b), nil
}
