// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elog

import (
	"github.com/platinasystems/go/elib"

	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"
)

type View struct {
	viewEvents
	bufferEvents []bufferEvent
	e            bufferEventVec
	name         string
	Times        timeBounds
	shared
}

func (v *View) SetName(name string) { v.name = name }
func (v *View) Name() string        { return v.name }
func (v *View) NumEvents() (l uint) {
	if v.bufferEvents != nil {
		l = uint(len(v.bufferEvents))
	} else {
		l = uint(len(v.ve))
	}
	return
}
func (v *View) Event(i uint) (h *eventHeader) {
	if v.bufferEvents != nil {
		return &v.bufferEvents[i].eventHeader
	}
	return &v.ve[i].eventHeader
}
func (v *View) EventLines(i uint) (s []string) {
	l := &Log{s: &v.shared}
	if v.bufferEvents != nil {
		return v.bufferEvents[i].lines(l)
	}
	return v.viewEventLines(l, i)
}
func (v *View) EventCaller(i uint) (c *CallerInfo) {
	e := v.Event(i)
	_, c = v.getCallerInfo(e.callerIndex)
	return
}

func (b *Buffer) NewView() (v *View) {
	v = &View{}

	v.shared.sharedHeader = b.shared.sharedHeader
	v.shared.stringTable.copyFrom(&b.shared.stringTable)
	v.shared.eventFilterShared.copyFrom(&b.shared.eventFilterShared)

	l := len(v.bufferEvents)
	cap := b.Cap()
	mask := b.capMask()
	v.e.Resize(uint(cap))
	i := int(b.lockIndex(true))
	if i >= cap {
		l += copy(v.e[l:], b.events[i&mask:])
	}
	l += copy(v.e[l:], b.events[0:i&mask])
	b.lockIndex(false)
	v.e = v.e[:l]
	v.bufferEvents = v.e
	v.getViewTimes()
	return
}

func NewView() *View { return DefaultBuffer.NewView() }

// Make subview with only events between elapsed times t0 and t1.
func (v *View) SubView(t0, t1 float64) (n uint) {
	l := len(v.e)
	if t0 > t1 {
		t0, t1 = t1, t0
	}
	i0 := sort.Search(l, func(i int) bool {
		et := v.elapsedTime(&v.e[i].eventHeader)
		return et >= t0
	})
	i1 := sort.Search(l, func(i int) bool {
		et := v.elapsedTime(&v.e[i].eventHeader)
		return et > t1
	})
	v.bufferEvents = v.e[i0:i1]
	v.doViewTimes(t0, t1, true)
	return uint(len(v.bufferEvents))
}
func (v *View) Reset() {
	v.bufferEvents = v.e
	v.getViewTimes()
}

type timeBounds struct {
	// Starting time truncated to nearest second.
	Start        time.Time
	Min, Max, Dt float64
	Unit         float64
	UnitName     string
}

func (v *View) getViewTimes() {
	if ne := v.NumEvents(); ne > 0 {
		e0, e1 := v.Event(0), v.Event(ne-1)
		t0 := e0.elapsedTime(&v.shared)
		t1 := e1.elapsedTime(&v.shared)
		v.doViewTimes(t0, t1, false)
	}
	return
}

func (v *View) doViewTimes(t0, t1 float64, isViewTime bool) (err error) {
	tb := &v.Times
	tUnit := float64(1)
	roundUnit := tUnit
	unitName := "sec"

	if isViewTime {
		// Translate view elapsed times to buffer elapsed times.
		dt := v.StartTime.Sub(tb.Start).Seconds()
		t0 -= dt
		t1 -= dt
	}

	if t1 > t0 {
		v := math.Floor(math.Log10(t1 - t0))
		switch {
		case v < -6:
			unitName = "ns"
			tUnit = 1e-9
			switch {
			case v < -8:
				roundUnit = 1e-9
			case v < -7:
				roundUnit = 1e-8
			default:
				roundUnit = 1e-7
			}
		case v < -3:
			unitName = "μs"
			tUnit = 1e-6
			switch {
			case v < -5:
				roundUnit = 1e-6
			case v < -6:
				roundUnit = 1e-7
			default:
				roundUnit = 1e-8
			}
		case v < 0:
			unitName = "ms"
			tUnit = 1e-3
			switch {
			case v < -5:
				roundUnit = 1e-5
			case v < -4:
				roundUnit = 1e-4
			default:
				roundUnit = 1e-3
			}
		}
	}

	// Round buffer start time to seconds and add difference (nanoseconds part) to times.
	startTime := v.StartTime.Truncate(time.Duration(1e9 * tUnit))
	dt := v.StartTime.Sub(startTime).Seconds()
	t0 += dt
	t1 += dt

	t0 = roundUnit * math.Floor(t0/roundUnit)
	t1 = roundUnit * math.Ceil(t1/roundUnit)

	tb.Start = startTime
	tb.Min = t0
	tb.Max = t1
	tb.Dt = t1 - t0
	tb.Unit = tUnit
	tb.UnitName = unitName
	return
}

func (v *View) Print(w io.Writer, verbose bool) {
	type row struct {
		Time  string `format:"%-30s"`
		Data  string `format:"%s" align:"left" width:"60"`
		Delta string `format:"%s" align:"left" width:"9"`
		Path  string `format:"%s" align:"left" width:"30"`
	}
	colMap := map[string]bool{
		"Delta": verbose,
		"Path":  verbose,
	}
	ne := v.NumEvents()
	rows := make([]row, 0, ne)
	lastTime := 0.
	for ei := uint(0); ei < ne; ei++ {
		e := v.Event(ei)
		t, delta := v.ElapsedTime(e), 0.
		if ei > 0 {
			delta = t - lastTime
		}
		lastTime = t
		lines := v.EventLines(ei)
		for j := range lines {
			if lines[j] == "" {
				continue
			}
			indent := ""
			if j > 0 {
				indent = "  "
			}
			r := row{
				Data: indent + lines[j],
			}
			if j == 0 {
				r.Time = e.timeString(&v.shared)
				r.Delta = fmt.Sprintf("%8.6f", delta)
				_, c := v.getCallerInfo(e.callerIndex)
				r.Path = c.Name
			}
			rows = append(rows, r)
		}
	}
	elib.Tabulate(rows).WriteCols(w, colMap)
}
func (b *Buffer) Print(w io.Writer, detail bool) { b.NewView().Print(w, detail) }

// Dump log on SIGUP.
func (b *Buffer) PrintOnHangupSignal(w io.Writer, detail bool) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)
	for {
		<-c
		v := b.NewView()
		v.Print(w, detail)
	}
}
func PrintOnHangupSignal(w io.Writer, detail bool) { DefaultBuffer.PrintOnHangupSignal(w, detail) }

type viewEvent struct {
	eventHeader

	// View buffer b[lo:hi] contains encoded format and arguments.
	lo, hi uint32
}

type viewEvents struct {
	ve   []viewEvent
	b    elib.ByteVec
	args []interface{}
}

func (v *viewEvents) viewEventLines(l *Log, ei uint) []string {
	e := &v.ve[ei]
	b := v.b[e.lo:e.hi]
	i := 0
	if l.l != nil {
		l.l = l.l[:0]
	}
	for i < len(b) {
		if v.args != nil {
			v.args = v.args[:0]
		}
		x, n := binary.Uvarint(b[i:])
		i += n
		format := l.s.GetString(StringRef(x))

		for {
			var (
				a    interface{}
				kind byte
			)
			if a, kind, i = l.s.decodeArg(b, i); kind == fmtEnd {
				l.sprintf(format, v.args...)
				break
			} else {
				v.args = append(v.args, a)
			}
		}
	}
	return l.l
}

func (v *viewEvents) addBufferEvent(l *Log, e *bufferEvent) {
	lo := v.b.Len()
	i := lo
	l.f = func(format string, args ...interface{}) {
		_, i = fmtEncode(l.s, &v.b, i, true, nil, StringRefNil, format, args)
	}
	r := l.s.callers[e.callerIndex]
	e.format(r, l)
	var ve viewEvent
	ve.eventHeader = e.eventHeader
	ve.lo = uint32(lo)
	ve.hi = uint32(i)
	v.ve = append(v.ve, ve)
}

func (v *View) convertBufferEvents() {
	l := &Log{s: &v.shared}
	for i := range v.bufferEvents {
		v.addBufferEvent(l, &v.bufferEvents[i])
	}
	v.bufferEvents = nil
	v.e = nil
	return
}
