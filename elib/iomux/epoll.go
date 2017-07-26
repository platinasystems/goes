// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux

package iomux

import (
	"fmt"
	"syscall"
	"unsafe"
)

type eventMask uint32
type epollCtlOp int

const (
	eventRead  eventMask = 0x1
	eventWrite eventMask = 0x4
	eventError eventMask = 0x8

	opAdd epollCtlOp = 1 /* Add a file descriptor to the interface.  */
	opDel epollCtlOp = 2 /* Remove a file descriptor from the interface.  */
	opMod epollCtlOp = 3 /* Change file descriptor epoll_event structure.  */
)

type epollEvent struct {
	mask eventMask
	data [2]uint32
}

func epoll_ctl(epfd int, op epollCtlOp, fd int, event *epollEvent) (err error) {
	_, _, e := syscall.RawSyscall6(syscall.SYS_EPOLL_CTL, uintptr(epfd), uintptr(op), uintptr(fd), uintptr(unsafe.Pointer(event)), 0, 0)
	if e != 0 {
		err = e
	}
	return
}

func epoll_pwait(epfd int, events []epollEvent, secs float64) (n int, err error) {
	var zero [8]byte // Zero signal mask so any signal will stop wait.
	msec := -1
	if secs > 0 {
		msec = int(1e-3 * secs)
	}
	r0, _, e := syscall.Syscall6(syscall.SYS_EPOLL_PWAIT, uintptr(epfd), uintptr(unsafe.Pointer(&events[0])), uintptr(len(events)), uintptr(msec),
		uintptr(unsafe.Pointer(&zero[0])), uintptr(unsafe.Sizeof(zero)))
	n = int(r0)
	if e != 0 && e != syscall.EINTR {
		err = e
	}
	return
}

func epoll_create1(flag int) (fd int, err error) {
	r0, _, e := syscall.RawSyscall(syscall.SYS_EPOLL_CREATE1, uintptr(flag), 0, 0)
	fd = int(r0)
	if e != 0 {
		err = e
	}
	return
}

func (m *Mux) maybe_epoll_create() {
	m.once.Do(func() {
		var err error
		m.fd, err = epoll_create1(0)
		if err != nil {
			panic(fmt.Errorf("epoll_create %s", err))
		}
	})
}

func (l *File) event(f Filer) (e epollEvent) {
	if !l.disableRead {
		e.mask = eventRead
	}
	if !l.disableWrite && f.WriteAvailable() {
		e.mask |= eventWrite
	}
	e.data[0] = uint32(l.poolIndex)
	return
}

// Add adds a file to the file poller, certainly for read and possibly for write depending on f.WriteReady()
func (m *Mux) Add(f Filer) {
	m.poolLock.Lock()
	defer m.poolLock.Unlock()
	m.maybe_epoll_create()
	l := f.GetFile()
	fd := l.Fd
	if err := syscall.SetNonblock(fd, true); err != nil {
		panic(fmt.Errorf("setnonblock: %s", err))
	}

	fi := m.filePool.GetIndex()
	m.files[fi] = f
	l.poolIndex = fi

	e := l.event(f)
	if err := epoll_ctl(m.fd, opAdd, fd, &e); err != nil {
		panic(fmt.Errorf("epoll_ctl: add %s", err))
	}
}

// Del removes the file (descriptor) from polling and frees file pool entry.
func (m *Mux) Del(f Filer) {
	m.poolLock.Lock()
	defer m.poolLock.Unlock()
	l := f.GetFile()
	if err := epoll_ctl(m.fd, opDel, l.Fd, nil); err != nil {
		panic(fmt.Errorf("epoll_ctl: del %s", err))
	}
	fi := l.poolIndex
	// Poison index.
	l.poolIndex = ^uint(0)
	m.filePool.PutIndex(fi)
	m.files[fi] = nil
}

// Update is needed when f.WriteReady() changes value.
func (m *Mux) Update(f Filer) {
	m.poolLock.Lock()
	defer m.poolLock.Unlock()
	l := f.GetFile()
	e := l.event(f)
	if err := epoll_ctl(m.fd, opMod, l.Fd, &e); err != nil {
		panic(fmt.Errorf("epoll_ctl: mod %s", err))
	}
}

func (m *Mux) do(e *epollEvent) {
	m.poolLock.Lock()
	fi := e.data[0]
	f := m.files[fi]
	m.poolLock.Unlock()

	em := e.mask

	// Deleted file?
	if f == nil {
		return
	}

	if em&eventWrite != 0 {
		if err := f.WriteReady(); err != nil {
			m.logError(err)
		}
	}
	if em&eventRead != 0 {
		if err := f.ReadReady(); err != nil {
			m.logError(err)
		}
	}
	if em&eventError != 0 {
		if err := f.ErrorReady(); err != nil {
			m.logError(err)
		}
	}
}

func (m *Mux) EventPoll() {
	var events [256]epollEvent
	m.maybe_epoll_create()
	es := events[:]
	n, err := epoll_pwait(m.fd, es, float64(-1))
	if err != nil {
		panic(fmt.Errorf("epoll_pwait %s", err))
	}
	for i := 0; i < n; i++ {
		m.do(&es[i])
	}
}
