// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// High speed event logging
package elog

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cpu"

	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"reflect"
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
	log2EventBytes = 6
	EventDataBytes = 1<<log2EventBytes - (1*8 + 1*4)
)

type Event struct {
	// Cpu time when event was logged.
	timestamp uint64

	// Caller index (implies PC) which uniquely identifies log caller.
	callerIndex uint32

	// Track which can be used to separate events onto different
	// "tracks" which can be separately viewed graphically.
	track uint32

	// Opaque event data.  Type dependent.
	// 1 or more cache lines could follow depending on callerIndex.
	data [EventDataBytes]byte
}

type Format func(format string, args ...interface{}) string

type eventData interface {
	Format(c *Context, f Format) string
	SetData(c *Context, p Pointer)
}

type EventTrack struct {
	Name  string
	index uint32
}

// Context for stringers.
type Context shared

func (v *shared) GetContext() *Context { return (*Context)(v) }

type StringRef uint32

const (
	StringRefNil    = StringRef(0)
	SizeofStringRef = unsafe.Sizeof(StringRef(0))
)

func (v *Context) GetString(si StringRef) string { return v.stringTable.Get(si) }
func (v *Context) SetString(s string) StringRef  { return v.stringTable.Set(s) }
func (v *Context) SetStringf(format string, args ...interface{}) StringRef {
	return v.stringTable.Setf(format, args...)
}

func GetContext() *Context          { return DefaultBuffer.GetContext() }
func GetString(si StringRef) string { return DefaultBuffer.stringTable.Get(si) }
func SetString(s string) StringRef  { return DefaultBuffer.stringTable.Set(s) }
func SetStringf(format string, args ...interface{}) StringRef {
	return DefaultBuffer.stringTable.Setf(format, args...)
}

type stringTable struct {
	t []byte
	m map[string]StringRef
}

func (t *stringTable) Get(si StringRef) (s string) {
	s, _ = t.get(si)
	return
}
func (t *stringTable) get(si StringRef) (s string, l int) {
	b := t.t[si:]
	l = strings.IndexByte(string(b), 0)
	s = string(b[:l])
	return
}

func (t *stringTable) Set(s string) (si StringRef) {
	var ok bool
	if si, ok = t.m[s]; ok {
		return
	}
	si = StringRef(len(t.t))
	if si == StringRefNil {
		t.t = append(t.t, 0)
		si = 1
	}

	if t.m == nil {
		t.m = make(map[string]StringRef)
	}
	t.m[s] = si
	s += "\x00" // null terminate
	t.t = append(t.t, s...)
	return
}

func (t *stringTable) Setf(format string, args ...interface{}) StringRef {
	return t.Set(fmt.Sprintf(format, args...))
}

func (t *stringTable) init(s string) {
	t.t = []byte(s)
	t.m = make(map[string]StringRef)
	i := 0
	for i < len(s) {
		si := StringRef(i)
		x, l := t.get(si)
		t.m[x] = si
		i += 1 + l
	}
}

type sharedOkToCopy struct {
	// Timestamp when log was created.
	cpuStartTime uint64

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

	pcHashSeed uint64

	eventFilterMain
	shared
}

func (b *Buffer) Enable(v bool) {
	{
		cyclesPerSec := cpu.TimeInit()
		b.timeUnitNsec = 1e9 / cyclesPerSec
	}
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
	pc          uint64
	fmtIndex    StringRef
	t           reflect.Type
	callerIndex uint32
	callerInfo  CallerInfo
	de          dataEvent
	fe          fmtEvent
}

// Event filter info shared between Buffer and View.
type eventFilterShared struct {
	mu         sync.RWMutex
	callerByPC map[uint64]*callerCache
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
	pc          uint64
}

type eventFilterMain struct {
	m map[string]*eventFilter
	h [1 << log2HLen]pcHashEntry
}

const log2HLen = 9

type Caller struct {
	time   uint64
	pc     uint64
	pcHash uint64
}
type Pointer unsafe.Pointer
type PointerToFirstArg unsafe.Pointer

func (b *Buffer) GetCaller(a PointerToFirstArg) (c Caller) {
	if Enabled() {
		t, p, h := getPC(unsafe.Pointer(a), b.pcHashSeed)
		c = Caller{time: t, pc: p, pcHash: h}
	}
	return
}
func GetCaller(a PointerToFirstArg) (c Caller) { return DefaultBuffer.GetCaller(a) }

func (m *Buffer) getCaller(caller Caller) (c *callerCache, disable bool) {
	// Check 1st level hash.  No lock required.
	pc := caller.pc
	pch := &m.h[caller.pcHash&(1<<log2HLen-1)]
	disable = pch.disable
	if pch.pc == pc {
		c = m.callers[pch.callerIndex]
		return
	}

	// Check 2nd level cache.
	m.mu.RLock()
	c, ok := m.callerByPC[pc]
	if ok {
		disable = c.f.disable
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
	path := runtime.FuncForPC(uintptr(pc)).Name()
	for _, f := range m.m {
		if ok := f.re.MatchString(path); ok {
			found = f
			disable = f.disable
			break
		}
	}
	if m.callerByPC == nil {
		m.callerByPC = make(map[uint64]*callerCache)
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
	// Filter change clears buffer.  genEvents may have pcs cached.
	b.Clear()
}
func ResetFilters() { DefaultBuffer.ResetFilters() }

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

func typeOf(d eventData) (t reflect.Type) {
	t = reflect.TypeOf(d)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return
}

func (e *Event) setData(c *Context, d eventData) { d.SetData(c, Pointer(&e.data[0])) }
func (b *Buffer) add1(d eventData, c Caller, r *callerCache) {
	e := b.getEvent()
	e.timestamp = c.time
	e.callerIndex = r.callerIndex
	r.t = typeOf(d)
	e.setData(b.GetContext(), d)
	return
}

func (b *Buffer) Add(d eventData, c Caller) {
	if !b.Enabled() {
		return
	}
	if r, disabled := b.getCaller(c); !disabled {
		b.add1(d, c, r)
	}
}

var DefaultBuffer = New(0)

func Addc(d eventData, c Caller)     { DefaultBuffer.Add(d, c) }
func Add(d eventData)                { DefaultBuffer.Add(d, DefaultBuffer.GetCaller(PointerToFirstArg(&d))) }
func Print(w io.Writer, detail bool) { DefaultBuffer.Print(w, detail) }
func Len() (n int)                   { return DefaultBuffer.Len() }
func Enable(v bool)                  { DefaultBuffer.Enable(v) }

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
	caller := b.GetCaller(PointerToFirstArg(&log2Len))
	b.cpuStartTime = caller.time
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

// Time event happened in seconds relative to start of log.
func (e *Event) elapsedTime(s *shared) float64 {
	return 1e-9 * float64(e.timestamp-s.cpuStartTime) * s.timeUnitNsec
}

// Time elapsed from start of buffer.
func (b *Buffer) ElapsedTime(e *Event) float64 { return e.elapsedTime(&b.shared) }

// Go time.Time that event happened.
func (e *Event) time(s *shared) time.Time {
	nsec := float64(e.timestamp-s.cpuStartTime) * s.timeUnitNsec
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

type CallerInfo struct {
	PC    uint64
	Entry uint64
	Name  string
	File  string
	Line  int
}

func (v *shared) getCallerInfo(ci uint32) (r *callerCache, c *CallerInfo) {
	r = v.callers[ci]
	c = &r.callerInfo
	if c.PC == 0 {
		pc := uintptr(r.pc)
		fi := runtime.FuncForPC(pc)
		c.PC = r.pc
		c.Entry = uint64(fi.Entry())
		c.Name = fi.Name()
		c.File, c.Line = fi.FileLine(pc)
	}
	return
}

func (v *shared) addCallerInfo(c CallerInfo, isFmtEvent bool) {
	pc := c.PC
	cc := &callerCache{pc: pc, callerIndex: uint32(len(v.callers)), callerInfo: c}
	var d eventData
	if isFmtEvent {
		d = &fmtEvent{}
	} else {
		d = &dataEvent{}
	}
	cc.t = typeOf(d)
	v.callers = append(v.callers, cc)
	if v.callerByPC == nil {
		v.callerByPC = make(map[uint64]*callerCache)
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

func (e *Event) format(t reflect.Type, x *Context, c Format) string {
	v := reflect.NewAt(t, unsafe.Pointer(&e.data[0]))
	in := []reflect.Value{reflect.ValueOf(x), reflect.ValueOf(fmt.Sprintf)}
	out := v.MethodByName("Format").Call(in)
	return out[0].Interface().(string)
}

func (e *Event) String(c *Context) string {
	r := c.callers[e.callerIndex]
	return e.format(r.t, c, fmt.Sprintf)
}
func (e *Event) Strings(c *Context) []string { return strings.Split(e.String(c), "\n") }
func (e *Event) timeString(sh *shared) string {
	return e.time(sh).Format("2006-01-02 15:04:05.000000000")
}

func (e *Event) eventString(sh *shared, detail bool) (s string) {
	s = fmt.Sprintf("%s: %s", e.timeString(sh), e.String(sh.GetContext()))
	return
}

func (v *View) EventString(e *Event) string            { return e.eventString(&v.shared, false) }
func (b *Buffer) EventString(e *Event) string          { return e.eventString(&b.shared, false) }
func (s *shared) EventCaller(e *Event) (c *CallerInfo) { _, c = s.getCallerInfo(e.callerIndex); return }
func (b *Buffer) EventCaller(e *Event) *CallerInfo     { return b.EventCaller(e) }

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
		lines := e.Strings(v.GetContext())
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

// Generic events using fmt.Printf formatting.

type fmtKind uint8

const (
	fmtEnd = iota
	fmtBoolTrue
	fmtBoolFalse
	fmtUint
	fmtInt
	fmtFloat
	fmtStringRef
	fmtString
)

type fmtEvent struct {
	format StringRef
	args   [EventDataBytes - SizeofStringRef]byte
}

func (e *fmtEvent) SetData(c *Context, p Pointer) { *(*fmtEvent)(p) = *e }
func (e *fmtEvent) Format(c *Context, f Format) string {
	g := c.GetString(e.format)
	args := e.decode(c)
	return f(g, args...)
}
func (e *fmtEvent) setFmt(c *Context, r StringRef, format string) {
	if r == StringRefNil {
		r = c.SetString(format)
	}
	e.format = r
}
func (e *fmtEvent) set(c *Context, r StringRef, format string, args []interface{}) {
	e.setFmt(c, r, format)
	e.encode(c, args)
}

func encodeInt(b []byte, i0 int, v int64) (i int) {
	i = i0
	b[i] = fmtInt
	i += 1 + binary.PutVarint(b[i+1:], v)
	return
}

func encodeUint(b []byte, i0 int, v uint64, kind int) (i int) {
	i = i0
	b[i] = byte(kind)
	i += 1 + binary.PutUvarint(b[i+1:], v)
	return
}

func (e *fmtEvent) encode(c *Context, args []interface{}) (i int) {
	b := e.args[:]
	i = binary.PutUvarint(b[i:], uint64(e.format))
	for _, a := range args {
		switch v := a.(type) {
		case bool:
			b[i] = fmtBoolFalse
			if v {
				b[i] = fmtBoolTrue
			}
			i++
		case string:
			l, left := len(v), len(b)
			if l > left-1 {
				l = left - 1
			}
			b[i] = byte(l)
			copy(b[i+1:], v)
			i += 1 + l
		case StringRef:
			i = encodeUint(b, i, uint64(v), fmtStringRef)
		case int8:
			i = encodeInt(b, i, int64(v))
		case int16:
			i = encodeInt(b, i, int64(v))
		case int32:
			i = encodeInt(b, i, int64(v))
		case int64:
			i = encodeInt(b, i, int64(v))
		case int:
			i = encodeInt(b, i, int64(v))
		case uint8:
			i = encodeUint(b, i, uint64(v), fmtUint)
		case uint16:
			i = encodeUint(b, i, uint64(v), fmtUint)
		case uint32:
			i = encodeUint(b, i, uint64(v), fmtUint)
		case uint64:
			i = encodeUint(b, i, uint64(v), fmtUint)
		case uint:
			i = encodeUint(b, i, uint64(v), fmtUint)
		case float64:
			i = encodeUint(b, i, uint64(math.Float64bits(v)), fmtFloat)
		case float32:
			i = encodeUint(b, i, uint64(float64(math.Float32bits(v))), fmtFloat)
		default:
			panic(fmt.Errorf("elog fmtEvent encode value with unknown type: %v", a))
		}
	}
	b[i] = fmtEnd
	i++
	return
}

func (e *fmtEvent) decode(c *Context) (args []interface{}) {
	b := e.args[:]
	i := 0

	{
		x, n := binary.Uvarint(b[i:])
		i += n
		e.format = StringRef(x)
	}

	for {
		kind := b[i]
		i++
		if kind == fmtEnd {
			break
		}
		switch kind {
		case fmtBoolTrue, fmtBoolFalse:
			args = append(args, b[i] == fmtBoolTrue)
			i++
		case fmtInt:
			x, n := binary.Varint(b[i:])
			i += n
			args = append(args, x)
		case fmtUint:
			x, n := binary.Uvarint(b[i:])
			i += n
			args = append(args, x)
		case fmtFloat:
			x, n := binary.Uvarint(b[i:])
			i += n
			f := math.Float64frombits(x)
			args = append(args, f)
		case fmtStringRef:
			x, n := binary.Uvarint(b[i:])
			i += n
			s := c.GetString(StringRef(x))
			args = append(args, s)
		case fmtString:
			l := int(b[i])
			args = append(args, string(b[i+1:i+1+l]))
			i += 1 + l
		default:
			panic(fmt.Errorf("elog fmtEvent decode unknown kind: 0x%x", kind))
		}
	}
	return
}

type dataEvent struct {
	b [EventDataBytes]byte
}

func (e *dataEvent) SetData(c *Context, p Pointer)      { *(*dataEvent)(p) = *e }
func (e *dataEvent) Format(c *Context, f Format) string { return f(String(e.b[:])) }
func (e *dataEvent) set(s string)                       { copy(e.b[:], s) }

func (b *Buffer) fmt(c Caller, format string, args []interface{}) {
	if !Enabled() {
		return
	}
	if !b.Enabled() {
		return
	}
	if r, disabled := b.getCaller(c); !disabled {
		if isFmt := len(args) == 0; isFmt {
			d := &r.de
			d.set(format)
			b.add1(d, c, r)
		} else {
			x := b.GetContext()
			f := &r.fe
			f.set(x, r.fmtIndex, format, args)
			r.fmtIndex = f.format
			b.add1(f, c, r)
		}
	}
}

func (b *Buffer) F(format string, args ...interface{}) {
	c := b.GetCaller(PointerToFirstArg(&format))
	b.fmt(c, format, args)
}
func (b *Buffer) Fc(format string, c Caller, args ...interface{}) {
	b.fmt(c, format, args)
}
func F(format string, args ...interface{}) {
	b := DefaultBuffer
	c := b.GetCaller(PointerToFirstArg(&format))
	b.fmt(c, format, args)
}
func Fc(format string, c Caller, args ...interface{}) {
	DefaultBuffer.fmt(c, format, args)
}

func (b *Buffer) S(s string) {
	c := b.GetCaller(PointerToFirstArg(&b))
	b.fmt(c, s, nil)
}
func (b *Buffer) Sc(s string, c Caller) {
	b.fmt(c, s, nil)
}
func S(s string) {
	b := DefaultBuffer
	c := b.GetCaller(PointerToFirstArg(&s))
	b.Sc(s, c)
}
func Sc(s string, c Caller) {
	DefaultBuffer.Sc(s, c)
}

func (b *Buffer) FBool(format string, v bool) {
	c := b.GetCaller(PointerToFirstArg(&format))
	b.FcBool(format, c, v)
}
func (b *Buffer) FcBool(format string, c Caller, v bool) {
	if r, disabled := b.getCaller(c); !disabled {
		f := &r.fe
		x := b.GetContext()
		f.setFmt(x, r.fmtIndex, format)
		r.fmtIndex = f.format
		i := binary.PutUvarint(f.args[:], uint64(f.format))
		bv := byte(fmtBoolFalse)
		if v {
			bv = fmtBoolTrue
		}
		f.args[i] = bv
		b.add1(f, c, r)
	}
}
func FBool(f string, v bool) {
	b := DefaultBuffer
	c := b.GetCaller(PointerToFirstArg(&f))
	b.FcBool(f, c, v)
}
func FcBool(f string, c Caller, v bool) { DefaultBuffer.FcBool(f, c, v) }

func (b *Buffer) FUint(format string, v uint64) {
	c := b.GetCaller(PointerToFirstArg(&format))
	b.FcUint(format, c, v)
}
func (b *Buffer) FcUint(format string, c Caller, v uint64) {
	if r, disabled := b.getCaller(c); !disabled {
		f := &r.fe
		x := b.GetContext()
		f.setFmt(x, r.fmtIndex, format)
		r.fmtIndex = f.format
		i := binary.PutUvarint(f.args[:], uint64(f.format))
		encodeUint(f.args[:], i, v, fmtUint)
		b.add1(f, c, r)
	}
}
func FUint(f string, v uint64) {
	b := DefaultBuffer
	c := b.GetCaller(PointerToFirstArg(&f))
	b.FcUint(f, c, v)
}
func FcUint(f string, c Caller, v uint64) { DefaultBuffer.FcUint(f, c, v) }

func (b *Buffer) F2Uint(format string, v0, v1 uint64) {
	c := b.GetCaller(PointerToFirstArg(&format))
	b.Fc2Uint(format, c, v0, v1)
}
func (b *Buffer) Fc2Uint(format string, c Caller, v0, v1 uint64) {
	if r, disabled := b.getCaller(c); !disabled {
		f := &r.fe
		x := b.GetContext()
		f.setFmt(x, r.fmtIndex, format)
		r.fmtIndex = f.format
		i := binary.PutUvarint(f.args[:], uint64(f.format))
		i = encodeUint(f.args[:], i, v0, fmtUint)
		i = encodeUint(f.args[:], i, v1, fmtUint)
		b.add1(f, c, r)
	}
}
func F2Uint(f string, v0, v1 uint64) {
	b := DefaultBuffer
	c := b.GetCaller(PointerToFirstArg(&f))
	b.Fc2Uint(f, c, v0, v1)
}
func Fc2Uint(f string, c Caller, v0, v1 uint64) { DefaultBuffer.Fc2Uint(f, c, v0, v1) }

// Make it so all events are either dataEvent or fmtEvent.
// We can't save other event types since decoder might not know about these types.
func (e *Event) normalize(c *Context) {
	r := c.callers[e.callerIndex]
	switch r.t {
	case reflect.TypeOf(fmtEvent{}), reflect.TypeOf(dataEvent{}):
		// nothing to do: already normal.
	default:
		copy := *e
		copyValid := false
		e.format(r.t, c, func(format string, args ...interface{}) (s string) {
			if copyValid {
				return
			}
			copyValid = true
			if len(args) == 0 {
				var d dataEvent
				d.set(format)
				copy.setData(c, &d)
			} else {
				var f fmtEvent
				f.set(c, StringRefNil, format, args)
				copy.setData(c, &f)
			}
			return
		})
	}
}

func (v *View) normalizeEvents() {
	c := v.GetContext()
	for i := range v.Events {
		v.Events[i].normalize(c)
	}
}
