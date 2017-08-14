// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// High speed event logging
package elog

import (
	"github.com/platinasystems/go/elib/cpu"

	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	log2EventBytes = 6
	EventDataBytes = 1<<log2EventBytes - (8 + 2*2)
)

type Event struct {
	timestamp cpu.Time

	typeIndex uint16
	track     uint16

	Data [EventDataBytes]byte
}

type EventTrack struct {
	Name  string
	index uint32
}

type EventType struct {
	Name     string
	Stringer func(e *Event) string
	Decode   func(b []byte, e *Event) int
	Encode   func(b []byte, e *Event) int

	index       uint32
	lock        sync.Mutex // protects following
	Tags        []string
	IndexForTag map[string]int
}

type shared struct {
	// Timestamp when log was created.
	cpuStartTime cpu.Time

	StartTime time.Time

	// Timer tick in nanosecond units.
	timeUnitNsec float64
}

const lockBit = 1 << 63

func (b *Buffer) Cap() int     { return (1 << b.log2Len) }
func (b *Buffer) capMask() int { return b.Cap() - 1 }

func (b *Buffer) getEvent() *Event {
	for {
		i := atomic.LoadUint64(&b.index)
		if i&lockBit == 0 && atomic.CompareAndSwapUint64(&b.index, i, i+1) {
			return &b.events[int(i)&b.capMask()]
		}
	}
}

func (b *Buffer) lockIndex(wantLock bool) uint64 {
	for {
		i := atomic.LoadUint64(&b.index)
		isLocked := i&lockBit != 0
		if isLocked == wantLock {
			continue
		}
		if !atomic.CompareAndSwapUint64(&b.index, i, i^lockBit) {
			continue
		}
		// Return index sans lock bit to user.
		return i &^ lockBit
	}
}

// A buffer of events being collected.
type Buffer struct {
	// Circular buffer of events.
	events []Event

	// Index into circular buffer.
	index uint64

	// Disable logging when index reaches limit.
	disableIndex uint64

	// Buffer has space for 1<<log2Len.
	log2Len uint

	// Dummy event to use when logging is disabled.
	disabledEvent Event

	shared
}

func (b *Buffer) Enable(v bool) {
	b.lockIndex(true)
	b.index &= lockBit
	b.disableIndex = 0
	if v {
		b.disableIndex = ^b.disableIndex ^ lockBit
	}
	b.lockIndex(false)
}

func (b *Buffer) Enabled() bool {
	return Enabled() && b.index < b.disableIndex
}

func (b *Buffer) Clear() {
	b.lockIndex(true)
	b.index = lockBit
	b.lockIndex(false)
}
func Clear() { DefaultBuffer.Clear() }

// Disable logging after specified number of events have been logged.
// This is used as a "debug trigger" when a certain target event has occurred.
// Events will be logged both before and after the target event.
func (b *Buffer) DisableAfter(n uint64) {
	if n > 1<<(b.log2Len-1) {
		n = 1 << (b.log2Len - 1)
	}
	b.lockIndex(true)
	b.disableIndex = (b.index &^ lockBit) + n
	b.lockIndex(false)
}

func (b *Buffer) Add(t *EventType) *Event {
	if !b.Enabled() {
		return &b.disabledEvent
	}
	e := b.getEvent()
	e.timestamp = cpu.TimeNow()
	e.typeIndex = uint16(t.index)
	return e
}

var (
	eventTypesLock sync.Mutex
	eventTypes     []*EventType
	typeByName     = make(map[string]*EventType)
)

func addTypeNoLock(t *EventType) {
	t.index = uint32(len(eventTypes))
	eventTypes = append(eventTypes, t)
}

func addType(t *EventType) {
	eventTypesLock.Lock()
	defer eventTypesLock.Unlock()
	addTypeNoLock(t)
}

func getTypeByIndex(i int) *EventType {
	eventTypesLock.Lock()
	defer eventTypesLock.Unlock()
	return eventTypes[i]
}

func (e *Event) getType() *EventType { return getTypeByIndex(int(e.typeIndex)) }

func RegisterType(t *EventType) {
	eventTypesLock.Lock()
	defer eventTypesLock.Unlock()
	if _, ok := typeByName[t.Name]; ok {
		panic("duplicate event type name: " + t.Name)
	}
	typeByName[t.Name] = t
	addTypeNoLock(t)
}

func getTypeByName(n string) (t *EventType, ok bool) {
	eventTypesLock.Lock()
	defer eventTypesLock.Unlock()
	t, ok = typeByName[n]
	return
}

var DefaultBuffer = New(0)

func Add(t *EventType) *Event { return DefaultBuffer.Add(t) }
func Print(w io.Writer)       { DefaultBuffer.Print(w) }
func Len() (n int)            { return DefaultBuffer.Len() }
func Enable(v bool)           { DefaultBuffer.Enable(v) }

func New(log2Len uint) (b *Buffer) {
	b = &Buffer{}
	switch {
	case log2Len == 0:
		log2Len = 10
	case log2Len < 8:
		log2Len = 8
	}
	b.events = make([]Event, 1<<log2Len)
	b.log2Len = log2Len
	b.cpuStartTime = cpu.TimeNow()
	b.StartTime = time.Now()
	return
}

func (s *shared) timeUnitNsecs() (u float64) {
	u = s.timeUnitNsec
	if u == 0 {
		var c cpu.Time
		c.Cycles(1 * cpu.Second)
		s.timeUnitNsec = 1e9 / float64(c)
		u = s.timeUnitNsec
	}
	return
}

// Time event happened in seconds relative to start of log.
func (e *Event) elapsedTime(s *shared) float64 {
	return 1e-9 * float64(e.timestamp-s.cpuStartTime) * s.timeUnitNsecs()
}

func (v *View) ElapsedTime(e *Event) float64   { return e.elapsedTime(&v.shared) }
func (b *Buffer) ElapsedTime(e *Event) float64 { return e.elapsedTime(&b.shared) }

// Go time.Time that event happened.
func (e *Event) time(s *shared) time.Time {
	nsec := float64(e.timestamp-s.cpuStartTime) * s.timeUnitNsecs()
	return s.StartTime.Add(time.Duration(nsec))
}

func (v *View) Time(e *Event) time.Time   { return e.time(&v.shared) }
func (b *Buffer) Time(e *Event) time.Time { return e.time(&b.shared) }

func (e *Event) unixNano(s *shared) float64 { return float64(e.time(s).UnixNano()) * 1e-9 }

func (v *View) AbsTime(e *Event) float64   { return e.unixNano(&v.shared) }
func (b *Buffer) AbsTime(e *Event) float64 { return e.unixNano(&b.shared) }

func (v *View) GetTimeBounds(tb *TimeBounds) (err error) {
	l := len(v.Events)
	if l == 0 {
		err = errors.New("no events in view")
		return
	}

	t0 := v.Events[0].elapsedTime(&v.shared)
	t1 := v.Events[l-1].elapsedTime(&v.shared)

	tUnit := float64(1)
	mult := float64(1)
	unitName := "sec"
	if t1 > t0 {
		v := math.Floor(math.Log10(t1 - t0))
		iv := float64(0)
		switch {
		case v < -6:
			iv = -9.
			tUnit = 1e-9
			unitName = "nsec"
		case v < -3:
			iv = -6.
			tUnit = 1e-6
			unitName = "μsec"
		case v < 0:
			iv = -3.
			tUnit = 1e-3
			unitName = "msec"
		}
		mult = math.Pow10(int(math.Floor(v - iv)))
	}

	// Round absolute Go start time to seconds and add difference (nanoseconds part) to times.
	startTime := v.StartTime.Truncate(time.Second)
	dt := 1e-9 * float64(v.StartTime.Sub(startTime))
	t0 += dt
	t1 += dt

	t0 = math.Floor(t0 / tUnit)
	t1 = math.Ceil(t1 / tUnit)

	t0 = tUnit * mult * math.Floor(t0/mult)
	t1 = tUnit * mult * math.Ceil(t1/mult)

	tb.Min = t0
	tb.Max = t1
	tb.Dt = t1 - t0
	tb.Round = mult
	tb.Unit = tUnit
	tb.Start = startTime
	tb.UnitName = unitName
	return
}

func (e *Event) Type() *EventType { return e.getType() }

func (e *Event) String() string { return e.getType().Stringer(e) }

func (e *Event) eventString(sh *shared) (s string) {
	s = fmt.Sprintf("%s: %s",
		e.time(sh).Format("2006-01-02 15:04:05.000000000"),
		e)
	return
}

func (v *View) EventString(e *Event) string   { return e.eventString(&v.shared) }
func (b *Buffer) EventString(e *Event) string { return e.eventString(&b.shared) }

func StringLen(b []byte) (l int) {
	l = bytes.IndexByte(b, 0)
	if l < 0 {
		l = len(b)
	}
	return
}

func String(b []byte) string {
	return string(b[:StringLen(b)])
}

func PutData(b []byte, data []byte) {
	b = PutUvarint(b, len(data))
	copy(b, data)
}

func HexData(p []byte) string {
	b, l := Uvarint(p)
	m := l
	dots := ""
	if m > len(b) {
		m = len(b)
		dots = "..."
	}
	return fmt.Sprintf("%d %x%s", l, b[:m], dots)
}

func Printf(b []byte, format string, a ...interface{}) {
	copy(b, fmt.Sprintf(format, a...))
}

func (b *Buffer) Len() (n int) {
	n = int(b.index)
	max := 1 << b.log2Len
	if n > max {
		n = max
	}
	return
}

func (b *Buffer) firstIndex() (f int) {
	f = int(b.index - 1<<b.log2Len)
	if f < 0 {
		f = 0
	}
	f &= 1<<b.log2Len - 1
	return
}

func (b *Buffer) GetEvent(index int) *Event {
	f := b.firstIndex()
	return &b.events[(f+index)&(1<<b.log2Len-1)]
}

type TimeBounds struct {
	// Starting time truncated to nearest second.
	Start        time.Time
	Min, Max, Dt float64
	Unit, Round  float64
	UnitName     string
}

type View struct {
	Events EventVec
	shared
}

//go:generate gentemplate -d Package=elog -id Event -d VecType=EventVec -d Type=Event github.com/platinasystems/go/elib/vec.tmpl

func (b *Buffer) NewView() (v *View) {
	v = &View{}
	v.shared = b.shared
	l := len(v.Events)
	cap := b.Cap()
	mask := b.capMask()
	v.Events.Resize(uint(cap))
	i := int(b.lockIndex(true))
	if i >= cap {
		l += copy(v.Events[l:], b.events[i&mask:])
	}
	l += copy(v.Events[l:], b.events[0:i&mask])
	b.lockIndex(false)
	v.Events = v.Events[:l]
	return
}

func NewView() *View { return DefaultBuffer.NewView() }

func (v *View) Print(w io.Writer) {
	for i := range v.Events {
		fmt.Fprintln(w, v.Events[i].eventString(&v.shared))
	}
}

func (b *Buffer) Print(w io.Writer) { b.NewView().Print(w) }

func (t *EventType) Tag(i int, sep string) (tag string) {
	tag = ""
	if i < len(t.Tags) {
		tag = t.Tags[i] + sep
	}
	return
}

func (t *EventType) TagIndex(s string) (i int) {
	t.lock.Lock()
	defer t.lock.Unlock()
	l := len(t.Tags)
	if t.IndexForTag == nil {
		t.IndexForTag = make(map[string]int)
	}
	i, ok := t.IndexForTag[s]
	if !ok {
		i = l
		t.IndexForTag[s] = i
		t.Tags = append(t.Tags, s)
	}
	return
}

// Dump log on SIGUP.
func (b *Buffer) PrintOnHangupSignal(w io.Writer) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)
	for {
		<-c
		v := b.NewView()
		v.Print(w)
	}
}
func PrintOnHangupSignal(w io.Writer) { DefaultBuffer.PrintOnHangupSignal(w) }

// Generic events
type genEvent struct {
	s [EventDataBytes]byte
}

func (e *genEvent) String() string      { return String(e.s[:]) }
func (e *genEvent) Encode(b []byte) int { return copy(b, e.s[:]) }
func (e *genEvent) Decode(b []byte) int { return copy(e.s[:], b) }

func GenEvent(s string) {
	if !Enabled() {
		return
	}
	e := genEvent{}
	copy(e.s[:], s)
	e.Log()
}

func GenEventf(format string, args ...interface{}) {
	if !Enabled() {
		return
	}
	e := genEvent{}
	Printf(e.s[:], format, args...)
	e.Log()
}

//go:generate gentemplate -d Package=elog -id genEvent -d Type=genEvent github.com/platinasystems/go/elib/elog/event.tmpl
