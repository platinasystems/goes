// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// High speed event logging
package elog

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cpu"

	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

const (
	log2EventBytes = 7
	EventDataBytes = 1<<log2EventBytes - (1*8 + 1*4 + 2*2)
)

type Event struct {
	// Cpu time when event was logged.
	timestamp cpu.Time

	// Caller index (implies PC) which uniquely identifies log caller.
	callerIndex uint32

	// Event type.
	typeIndex uint16

	// Track which can be used to separate events onto different
	// "tracks" which can be separately viewed graphically.
	track uint16

	// Opaque event data.  Type dependent.
	Data [EventDataBytes]byte
}

type EventTrack struct {
	Name  string
	index uint32
}

type EventType struct {
	Name     string
	Strings  func(c *Context, e *Event) []string
	Decode   func(c *Context, e *Event, b []byte) int
	Encode   func(c *Context, e *Event, b []byte) int
	index    uint32
	mu       sync.Mutex // protects following
	disabled bool
}

// Context for event type functions.
type Context shared

func (v *shared) GetContext() *Context        { return (*Context)(v) }
func (v *Context) GetString(si uint32) string { return v.stringTable.Get(si) }
func (v *Context) SetString(s string) uint32  { return v.stringTable.Set(s) }
func (v *Context) SetStringf(format string, args ...interface{}) uint32 {
	return v.stringTable.Setf(format, args...)
}

func GetContext() *Context       { return DefaultBuffer.GetContext() }
func GetString(si uint32) string { return DefaultBuffer.stringTable.Get(si) }
func SetString(s string) uint32  { return DefaultBuffer.stringTable.Set(s) }
func SetStringf(format string, args ...interface{}) uint32 {
	return DefaultBuffer.stringTable.Setf(format, args...)
}

type stringTable struct {
	t []byte
	m map[string]uint32
}

func (t *stringTable) Get(si uint32) (s string) {
	s, _ = t.get(si)
	return
}
func (t *stringTable) get(si uint32) (s string, l int) {
	b := t.t[si:]
	l = strings.IndexByte(string(b), 0)
	s = string(b[:l])
	return
}

func (t *stringTable) Set(s string) (si uint32) {
	var ok bool
	if si, ok = t.m[s]; ok {
		return
	}
	si = uint32(len(t.t))
	if t.m == nil {
		t.m = make(map[string]uint32)
	}
	t.m[s] = si
	s += "\x00" // null terminate
	t.t = append(t.t, s...)
	return
}

func (t *stringTable) Setf(format string, args ...interface{}) uint32 {
	return t.Set(fmt.Sprintf(format, args...))
}

func (t *stringTable) init(s string) {
	t.t = []byte(s)
	t.m = make(map[string]uint32)
	i := 0
	for i < len(s) {
		si := uint32(i)
		x, l := t.get(si)
		t.m[x] = si
		i += 1 + l
	}
}

type sharedOkToCopy struct {
	// Timestamp when log was created.
	cpuStartTime cpu.Time

	// Starting time of view.
	StartTime time.Time

	// Timer tick in nanosecond units.
	timeUnitNsec float64

	stringTable
}

// Shared between Buffer and View.
type shared struct {
	sharedOkToCopy
	eventFilterShared
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

	pcHashSeed uint64

	eventFilterMain
	shared
}

func (b *Buffer) Enable(v bool) {
	cpu.TimeInit()
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

type eventFilter struct {
	re      *regexp.Regexp
	disable bool
}

type callerCache struct {
	f           eventFilter
	pc          uintptr
	callerIndex uint32
	callerInfo  CallerInfo
}

// Event filter info shared between Buffer and View.
type eventFilterShared struct {
	mu         sync.RWMutex
	callerByPC map[uintptr]*callerCache
	callers    []*callerCache
}

// Copy everything except mutex.
func (dst *eventFilterShared) copyFrom(src *eventFilterShared) {
	dst.callerByPC = src.callerByPC
	dst.callers = src.callers
}

type pcHashEntry struct {
	disable     bool
	callerIndex uint32
	pc          uintptr
}

type eventFilterMain struct {
	m map[string]*eventFilter
	h [1 << log2HLen]pcHashEntry
}

const log2HLen = 9

type Caller uintptr
type PointerToFirstArg unsafe.Pointer

func GetCaller(a PointerToFirstArg) (c Caller) {
	if Enabled() {
		c = Caller(cpu.GetCallerPC(unsafe.Pointer(a)))
	}
	return
}

func rotl_31(x uint64) uint64 { return (x << 31) | (x >> (64 - 31)) }
func (m *Buffer) pcHash(pc uintptr) uint {
	const (
		// Constants for multiplication: four random odd 64-bit numbers.
		m1 = 16877499708836156737
		m2 = 2820277070424839065
		m3 = 9497967016996688599
	)
	h := uint64(pc) ^ m.pcHashSeed
	h = rotl_31(h*m1) * m2
	h ^= h >> 29
	h *= m3
	h ^= h >> 32
	return uint(h)
}

func (m *Buffer) eventDisabled(caller Caller) (callerIndex uint32, disable bool) {
	// Check 1st level hash.  No lock required.
	pc := uintptr(caller)
	pch := &m.h[m.pcHash(pc)&(1<<log2HLen-1)]
	callerIndex = uint32(pch.callerIndex)
	disable = pch.disable
	if pch.pc == pc {
		return
	}

	// Check 2nd level cache.
	m.mu.RLock()
	c, ok := m.callerByPC[pc]
	if ok {
		disable = c.f.disable
		callerIndex = c.callerIndex
		m.mu.RUnlock()
		// Update 1st level cache. No lock required.
		pch.pc = pc
		pch.callerIndex = c.callerIndex
		pch.disable = disable
		return
	}
	m.mu.RUnlock()

	// Now grab write lock.
	m.mu.Lock()
	// Miss? Scan regexps.
	var found *eventFilter
	path := runtime.FuncForPC(pc).Name()
	for _, f := range m.m {
		if ok := f.re.MatchString(path); ok {
			found = f
			disable = f.disable
			break
		}
	}
	if m.callerByPC == nil {
		m.callerByPC = make(map[uintptr]*callerCache)
	}
	c = &callerCache{pc: pc, callerIndex: uint32(len(m.callers))}
	m.callers = append(m.callers, c)
	if found != nil {
		c.f = *found
	}
	m.callerByPC[pc] = c
	pch.pc = pc
	pch.callerIndex = c.callerIndex
	pch.disable = disable
	m.mu.Unlock()
	return
}

var ErrFilterNotFound = errors.New("event filter not found")

func (m *Buffer) AddDelEventFilter(matching string, enable, isDel bool) (err error) {
	var f eventFilter
	if !isDel {
		if f.re, err = regexp.Compile(matching); err != nil {
			return
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if isDel {
		if _, ok := m.m[matching]; !ok {
			err = ErrFilterNotFound
			return
		}
		delete(m.m, matching)
		return
	}
	f.disable = !enable
	if m.m == nil {
		m.m = make(map[string]*eventFilter)
	}
	m.m[matching] = &f
	m.invalidateCache()
	m.applyFilters()
	return
}

func AddDelEventFilter(matching string, enable, isDel bool) (err error) {
	return DefaultBuffer.AddDelEventFilter(matching, enable, isDel)
}

// Reset caches to initial zero state.
func (m *eventFilterMain) invalidateCache() { *m = eventFilterMain{} }

func (b *Buffer) ResetFilters() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.m = nil
	for _, c := range b.callerByPC {
		c.f.disable = false
	}
	b.invalidateCache()
	b.applyFilters()
	// Filter change clears buffer.  genEvents may have pcs cached.
	b.Clear()
}
func ResetFilters() { DefaultBuffer.ResetFilters() }

func (m *eventFilterMain) applyFilters() {
	// Invalidate cache
	eventTypesLock.Lock()
	defer eventTypesLock.Unlock()
	for _, t := range eventTypes {
		disable := false
		for _, f := range m.m {
			if ok := f.re.MatchString(t.Name); ok {
				disable = f.disable
				break
			}
		}
		t.disabled = disable
	}
}

func (b *Buffer) clear(resize uint) {
	b.lockIndex(true)
	if resize != 0 {
		b.log2Len = elib.MaxLog2(elib.Word(resize))
		if b.log2Len < minLog2Len {
			b.log2Len = minLog2Len
		}
		if b.log2Len > maxLog2Len {
			b.log2Len = maxLog2Len
		}
		b.events = make([]Event, 1<<b.log2Len)
	}
	b.index = lockBit
	b.lockIndex(false)
}

func (b *Buffer) Clear() { b.clear(0) }
func Clear()             { DefaultBuffer.Clear() }

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

func (b *Buffer) Add(t *EventType, c Caller) (e *Event) {
	e = &b.disabledEvent
	if !b.Enabled() || t.disabled {
		return
	}
	if ci, disabled := b.eventDisabled(c); !disabled {
		e = b.getEvent()
		e.timestamp = cpu.TimeNow()
		e.callerIndex = ci
		e.typeIndex = uint16(t.index)
	}
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

func Add(t *EventType, c Caller) *Event { return DefaultBuffer.Add(t, c) }
func Print(w io.Writer, detail bool)    { DefaultBuffer.Print(w, detail) }
func Len() (n int)                      { return DefaultBuffer.Len() }
func Enable(v bool)                     { DefaultBuffer.Enable(v) }

const (
	minLog2Len = 10
	maxLog2Len = 24 // no need to allow buffer to be too large.
)

func New(log2Len uint) (b *Buffer) {
	b = &Buffer{}
	switch {
	case log2Len < minLog2Len:
		log2Len = minLog2Len
	case log2Len > maxLog2Len:
		log2Len = maxLog2Len
	}
	b.events = make([]Event, 1<<log2Len)
	b.log2Len = log2Len
	b.cpuStartTime = cpu.TimeNow()
	b.StartTime = time.Now()
	b.pcHashSeed = uint64(b.cpuStartTime)

	// Seed hash with random bytes.
	{
		if f, err := os.Open("/dev/urandom"); err == nil {
			var d [8]byte
			if _, err = f.Read(d[:]); err == nil {
				for i := range d {
					b.pcHashSeed ^= uint64(d[i]) << (8 * (uint(i) % 8))
				}
			}
			f.Close()
		}
	}

	return
}
func (b *Buffer) Resize(n uint) { b.clear(n) }
func Resize(n uint)             { DefaultBuffer.Resize(n) }

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

// Time elapsed from start of buffer.
func (b *Buffer) ElapsedTime(e *Event) float64 { return e.elapsedTime(&b.shared) }

// Go time.Time that event happened.
func (e *Event) time(s *shared) time.Time {
	nsec := float64(e.timestamp-s.cpuStartTime) * s.timeUnitNsecs()
	return s.StartTime.Add(time.Duration(nsec))
}

func (b *Buffer) Time(e *Event) time.Time   { return e.time(&b.shared) }
func (e *Event) unixNano(s *shared) float64 { return float64(e.time(s).UnixNano()) * 1e-9 }
func (b *Buffer) AbsTime(e *Event) float64  { return e.unixNano(&b.shared) }

func (v *View) Time(e *Event) time.Time  { return e.time(&v.shared) }
func (v *View) AbsTime(e *Event) float64 { return e.unixNano(&v.shared) }

// Elapsed time since view start time.  (As computed in roundViewTimes.)
func (v *View) ElapsedTime(e *Event) float64 { return e.time(&v.shared).Sub(v.Times.Start).Seconds() }

func (v *View) getViewTimes() {
	if l := len(v.e); l != 0 {
		t0 := v.e[0].elapsedTime(&v.shared)
		t1 := v.e[l-1].elapsedTime(&v.shared)
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

func (e *Event) Type() *EventType { return e.getType() }

type CallerInfo struct {
	PC    uintptr
	Entry uintptr
	Name  string
	File  string
	Line  int
}

func (v *shared) getCallerInfo(ci uint32) (c *CallerInfo) {
	r := v.callers[ci]
	c = &r.callerInfo
	if c.PC == 0 {
		pc := r.pc
		fi := runtime.FuncForPC(pc)
		c.PC = pc
		c.Entry = fi.Entry()
		c.Name = fi.Name()
		c.File, c.Line = fi.FileLine(pc)
	}
	return
}

func (v *shared) addCallerInfo(c CallerInfo) {
	pc := c.PC
	cc := &callerCache{pc: pc, callerIndex: uint32(len(v.callers)), callerInfo: c}
	v.callers = append(v.callers, cc)
	if v.callerByPC == nil {
		v.callerByPC = make(map[uintptr]*callerCache)
	}
	v.callerByPC[pc] = cc
}

func (c *CallerInfo) ShortPath(p string, max uint) (f string, overflow bool) {
	if strings.Index(p, "/") < 0 {
		f = p
		return
	}

	fs := strings.Split(p, "/")
	i, n := uint(len(fs))-1, uint(0)
	for {
		if f != "" {
			f = "/" + f
			n++
		}
		l := uint(len(fs[i]))
		if overflow = n+l > max; overflow {
			break
		}
		f = fs[i] + f
		n += l
		if i == 0 {
			break
		}
		i--
	}
	if overflow {
		f = f[1:] // skip /
	}
	return
}

func (s *shared) Strings(e *Event) []string { t := e.getType(); return t.Strings(s.GetContext(), e) }
func (e *Event) timeString(sh *shared) string {
	return e.time(sh).Format("2006-01-02 15:04:05.000000000")
}

func (e *Event) eventString(sh *shared, detail bool) (s string) {
	s = fmt.Sprintf("%s: %s", e.timeString(sh), strings.Join(sh.Strings(e), " "))
	if detail {
		s += "(" + e.getType().Name + ")"
	}
	return
}

func (v *View) EventString(e *Event) string        { return e.eventString(&v.shared, false) }
func (b *Buffer) EventString(e *Event) string      { return e.eventString(&b.shared, false) }
func (v *View) EventCaller(e *Event) *CallerInfo   { return v.getCallerInfo(e.callerIndex) }
func (b *Buffer) EventCaller(e *Event) *CallerInfo { return b.getCallerInfo(e.callerIndex) }

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

type timeBounds struct {
	// Starting time truncated to nearest second.
	Start        time.Time
	Min, Max, Dt float64
	Unit         float64
	UnitName     string
}

type View struct {
	Events EventVec
	e      EventVec
	Times  timeBounds
	shared
}

//go:generate gentemplate -d Package=elog -id Event -d VecType=EventVec -d Type=Event github.com/platinasystems/go/elib/vec.tmpl

func (b *Buffer) NewView() (v *View) {
	v = &View{}

	v.shared.sharedOkToCopy = b.shared.sharedOkToCopy
	v.shared.eventFilterShared.copyFrom(&b.shared.eventFilterShared)

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
	v.e = v.Events // save a copy for sub views
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
		et := v.ElapsedTime(&v.e[i])
		return et >= t0
	})
	i1 := sort.Search(l, func(i int) bool {
		et := v.ElapsedTime(&v.e[i])
		return et > t1
	})
	v.Events = v.e[i0:i1]
	v.doViewTimes(t0, t1, true)
	return uint(len(v.Events))
}
func (v *View) Reset() {
	v.Events = v.e
	v.getViewTimes()
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
	rows := make([]row, 0, len(v.Events))
	lastTime := 0.
	for i := range v.Events {
		e := &v.Events[i]
		t, delta := v.ElapsedTime(e), 0.
		if i > 0 {
			delta = t - lastTime
		}
		lastTime = t
		lines := v.Strings(e)
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
				c := v.getCallerInfo(e.callerIndex)
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

// Generic events
type genEvent struct {
	s string
}

func (e *genEvent) Strings(c *Context) []string { return strings.Split(e.s, "\n") }
func (e *genEvent) Encode(c *Context, b []byte) int {
	i := 0
	l := len(e.s)
	if i+l < len(b) {
		b[i+l] = 0 // null terminate
	}
	i += copy(b[i:], e.s)
	return i
}
func (e *genEvent) Decode(c *Context, b []byte) int {
	i := 0
	l := StringLen(b[i:])
	e.s = String(b[i:])
	return i + l
}

func (x *genEvent) log(b *Buffer, c Caller) {
	e := b.Add(genEventType, c)
	x.s = strings.TrimSpace(x.s)
	x.Encode(b.GetContext(), e.Data[:])
}

func GenEvent(s string) {
	if !Enabled() {
		return
	}
	c := GetCaller(PointerToFirstArg(&s))
	e := genEvent{s: s}
	e.log(DefaultBuffer, c)
}

func GenEventc(s string, c Caller) {
	if !Enabled() {
		return
	}
	e := genEvent{s: s}
	e.log(DefaultBuffer, c)
}

func GenEventf(format string, args ...interface{}) {
	if !Enabled() {
		return
	}
	c := GetCaller(PointerToFirstArg(&format))
	e := genEvent{s: fmt.Sprintf(format, args...)}
	e.log(DefaultBuffer, c)
}

func GenEventfc(format string, c Caller, args ...interface{}) {
	if !Enabled() {
		return
	}
	e := genEvent{s: fmt.Sprintf(format, args...)}
	e.log(DefaultBuffer, c)
}

//go:generate gentemplate -d Package=elog -id genEvent -d Type=genEvent github.com/platinasystems/go/elib/elog/event.tmpl
