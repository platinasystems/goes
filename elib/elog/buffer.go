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
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	log2EventBytes = 7
	EventDataBytes = 1<<log2EventBytes - (1*8 + 2*4)
)

type eventHeader struct {
	// Cpu time when event was logged.
	timestamp uint64

	// Caller index (implies PC) which uniquely identifies log caller.
	callerIndex uint32

	// Track which can be used to separate events onto different
	// "tracks" which can be separately viewed graphically.
	trackIndex uint32
}

type bufferEvent struct {
	eventHeader

	// Opaque event data.  Type dependent.
	// 1 or more cache lines could follow depending on callerIndex.
	data [EventDataBytes]byte
}

//go:generate gentemplate -d Package=elog -id Event -d VecType=bufferEventVec -d Type=bufferEvent github.com/platinasystems/go/elib/vec.tmpl

func (e *eventHeader) CallerIndex() uint { return uint(e.callerIndex) }

type StringRef uint32

const (
	StringRefNil    = StringRef(0)
	SizeofStringRef = unsafe.Sizeof(StringRef(0))
)

func (v *shared) GetString(si StringRef) string { return v.stringTable.Get(si) }
func (v *shared) SetString(s string) StringRef  { return v.stringTable.Set(s) }
func (v *shared) SetStringf(format string, args ...interface{}) StringRef {
	return v.stringTable.Setf(format, args...)
}

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

func (dst *stringTable) copyFrom(src *stringTable) { dst.init(string(src.t)) }

type sharedHeader struct {
	// CPU timestamp when log was created.
	cpuStartTime uint64

	// CPU timer tick in nanosecond units.
	cpuTimeUnitNsec float64

	// Starting time of view.
	StartTime time.Time
}

// Shared between Buffer and View.
type shared struct {
	sharedHeader
	eventFilterShared
	stringTable
	// Protects fmtBuffer and string table.
	fmtMu sync.Mutex
	// Saved buffer for formatting.
	fmtBuffer elib.ByteVec
}

const lockBit = 1 << 63

func (b *Buffer) Cap() int     { return (1 << b.log2Len) }
func (b *Buffer) capMask() int { return b.Cap() - 1 }

func (b *Buffer) getEvent() *bufferEvent {
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
	events []bufferEvent

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
		b.cpuTimeUnitNsec = 1e9 / cyclesPerSec
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
	f                 eventFilter
	pc                uint64
	fmtIndex          StringRef
	dataType          reflect.Type
	fastDataType      reflect.Type
	formatMethodIndex int
	callerIndex       uint32
	callerInfo        CallerInfo
	fe                fmtEvent
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

func (m *Buffer) getCaller(d Logger, caller Caller) (c *callerCache, disable bool) {
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
	c = &callerCache{pc: pc, callerIndex: uint32(len(m.callers)), formatMethodIndex: -1}
	if d == nil {
		d = &c.fe
	}
	c.initDataType(d)
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
		b.events = make([]bufferEvent, 1<<b.log2Len)
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

const eventDataFormatMethod = "Elog"

type Format func(format string, args ...interface{})
type Log struct {
	s *shared
	f Format
	l []string
}

func (l *Log) Logf(format string, args ...interface{}) { l.f(format, args...) }
func (l *Log) GetString(si StringRef) string           { return l.s.GetString(si) }

type Data interface {
	ElogData() Logger
}

type Logger interface {
	Elog(l *Log)
}

type iface struct{ tab, data Pointer }

// Avoid reflect.TypeOf() so that d does not escape.
func (r *callerCache) initDataType(d Logger) {
	tint0 := reflect.TypeOf(int(0))
	i0 := (*iface)(Pointer(&tint0))
	di := (*iface)(Pointer(&d))
	dtab := (*iface)(di.tab)
	var x iface
	x.tab = i0.tab // <rtype,reflect.Type>
	x.data = dtab.data
	r.dataType = nil
	r.fastDataType = *(*reflect.Type)(Pointer(&x))
}

func (r *callerCache) getDataType() reflect.Type {
	if r.dataType == nil {
		t := r.fastDataType
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		r.dataType = t
	}
	return r.dataType
}

func (e *bufferEvent) setData(d Logger) {
	type u [EventDataBytes / 8]uint64
	var src, dst *u
	// Equivalent to:
	//    src = (*u)(unsafe.Pointer(reflect.ValueOf(d).Pointer()))
	// but avoids d escaping to heap.
	src = (*u)((*iface)(Pointer(&d)).data)
	dst = (*u)(unsafe.Pointer(&e.data[0]))
	i, n_left := 0, EventDataBytes
	for n_left >= 6*8 {
		dst[i+0] = src[i+0]
		dst[i+1] = src[i+1]
		dst[i+2] = src[i+2]
		dst[i+3] = src[i+3]
		dst[i+4] = src[i+4]
		dst[i+5] = src[i+5]
		n_left -= 8 * 6
		i += 6
	}
	for n_left > 0 {
		dst[i+0] = src[i+0]
		n_left -= 8 * 1
		i += 1
	}
}

func (b *Buffer) add1(d Logger, c Caller, t uint, r *callerCache) {
	e := b.getEvent()
	e.timestamp = c.time
	e.callerIndex = r.callerIndex
	e.trackIndex = uint32(t)
	e.setData(d)
	return
}

func (b *Buffer) AddTrack(d Logger, c Caller, t uint) {
	if !b.Enabled() {
		return
	}
	if r, disabled := b.getCaller(d, c); !disabled {
		b.add1(d, c, t, r)
	}
}
func (b *Buffer) Add(d Logger, c Caller)   { b.AddTrack(d, c, 0) }
func (b *Buffer) AddData(d Data, c Caller) { b.AddTrack(d.ElogData(), c, 0) }

var DefaultBuffer = New(0)

func Addc(d Logger, c Caller)   { DefaultBuffer.Add(d, c) }
func Add(d Logger)              { DefaultBuffer.Add(d, DefaultBuffer.GetCaller(PointerToFirstArg(&d))) }
func AddDatac(d Data, c Caller) { DefaultBuffer.AddData(d, c) }
func AddData(d Data)            { DefaultBuffer.AddData(d, DefaultBuffer.GetCaller(PointerToFirstArg(&d))) }

func Print(w io.Writer, detail bool) { DefaultBuffer.Print(w, detail) }
func Len() (n int)                   { return DefaultBuffer.Len() }
func Enable(v bool)                  { DefaultBuffer.Enable(v) }

const (
	minLog2Len = 12
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
	b.events = make([]bufferEvent, 1<<log2Len)
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

// Time event happened in seconds relative to start of buffer.
func (e *eventHeader) elapsedTime(s *shared) float64 {
	return 1e-9 * float64(e.timestamp-s.cpuStartTime) * s.cpuTimeUnitNsec
}

// Time elapsed from start of buffer.
func (s *shared) elapsedTime(e *eventHeader) float64 { return e.elapsedTime(s) }

// Go time.Time that event happened.
func (e *eventHeader) goTime(s *shared) time.Time {
	nsec := float64(e.timestamp-s.cpuStartTime) * s.cpuTimeUnitNsec
	return s.StartTime.Add(time.Duration(nsec))
}

func (s *shared) unixNano(e *eventHeader) float64 { return float64(e.goTime(s).UnixNano()) * 1e-9 }
func (s *shared) absTime(e *eventHeader) float64  { return s.unixNano(e) }

func (v *View) goTime(e *eventHeader) time.Time { return e.goTime(&v.shared) }
func (v *View) absTime(e *eventHeader) float64  { return v.shared.unixNano(e) }

// Elapsed time since view start time.
func (v *View) ElapsedTime(e *eventHeader) float64 {
	return e.goTime(&v.shared).Sub(v.Times.StartTime).Seconds()
}
func (e *eventHeader) ElapsedTime(v *View) float64 {
	return e.goTime(&v.shared).Sub(v.Times.StartTime).Seconds()
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

func (v *shared) addCallerInfo(c CallerInfo) {
	pc := c.PC
	cc := &callerCache{pc: pc, callerIndex: uint32(len(v.callers)), callerInfo: c}
	var d fmtEvent
	cc.initDataType(&d)
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
	n_fs := uint(len(fs))
	i, n := n_fs-1, uint(0)
	for {
		if f != "" {
			f = "/" + f
			n++
		}
		l := uint(len(fs[i]))
		if overflow = i+1 < n_fs && n+l > max; overflow {
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

func (e *bufferEvent) format(r *callerCache, l *Log) {
	t := r.getDataType()
	v := reflect.NewAt(t, unsafe.Pointer(&e.data[0]))
	if r.formatMethodIndex < 0 {
		m, ok := v.Type().MethodByName(eventDataFormatMethod)
		if !ok {
			panic("elog: no method " + eventDataFormatMethod + " for type " + t.Name())
		}
		r.formatMethodIndex = m.Index
	}
	f := v.Method(r.formatMethodIndex)
	in := []reflect.Value{reflect.ValueOf(l)}
	f.Call(in)
}

func (l *Log) sprintf(format string, args ...interface{}) {
	// Resolve any StringRef to strings in args.
	for i := range args {
		a := args[i]
		switch v := a.(type) {
		case StringRef:
			args[i] = l.s.GetString(v)
		}
	}

	line := fmt.Sprintf(format, args...)
	lines := strings.Split(line, "\n")
	for i := range lines {
		l.l = append(l.l, strings.TrimSpace(lines[i]))
	}
}

func (e *bufferEvent) lines(l *Log) []string {
	r := l.s.callers[e.callerIndex]
	l.f = l.sprintf
	if l.l != nil {
		l.l = l.l[:0]
	}
	e.format(r, l)
	var n int
	for n = len(l.l); n > 0 && len(l.l[n-1]) == 0; n-- {
	}
	l.l = l.l[:n]
	return l.l
}
func (e *bufferEvent) string(l *Log) string {
	s := e.lines(l)
	return strings.Join(s, "\n")
}
func (e *eventHeader) timeString(s *shared) string {
	return e.goTime(s).Format("2006-01-02 15:04:05.000000000")
}
func (e *bufferEvent) eventString(l *Log) (s string) {
	s = fmt.Sprintf("%s: %s", e.timeString(l.s), e.string(l))
	return
}

func (s *shared) eventString(e *bufferEvent) string {
	var l Log
	l.s = s
	return e.string(&l)
}

func (s *shared) eventCaller(e *eventHeader) (c *CallerInfo) {
	_, c = s.getCallerInfo(e.callerIndex)
	return
}
func (b *Buffer) eventCaller(e *eventHeader) *CallerInfo { return b.eventCaller(e) }

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
