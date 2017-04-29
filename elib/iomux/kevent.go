// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin

package iomux

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	EVFILT_READ     = -1
	EVFILT_WRITE    = -2
	EVFILT_AIO      = -3  // attached to aio requests
	EVFILT_VNODE    = -4  // attached to vnodes
	EVFILT_PROC     = -5  // attached to struct proc
	EVFILT_SIGNAL   = -6  // attached to struct proc
	EVFILT_TIMER    = -7  // timers
	EVFILT_MACHPORT = -8  // Mach portsets
	EVFILT_FS       = -9  // Filesystem events
	EVFILT_USER     = -10 // User events
	EVFILT_VM       = -12 // Virtual memory events

	/* actions */
	EV_ADD     = 1 << 0 // add event to kq (implies enable)
	EV_DELETE  = 1 << 1 // delete event from kq
	EV_ENABLE  = 1 << 2 // enable event
	EV_DISABLE = 1 << 3 // disable event (not reported)

	/* flags */
	EV_ONESHOT  = 1 << 4 // only report one occurrence
	EV_CLEAR    = 1 << 5 // clear event state after reporting
	EV_RECEIPT  = 1 << 6 // force EV_ERROR on success, data == 0
	EV_DISPATCH = 1 << 7 // disable event after reporting

	/* returned values */
	EV_EOF   = 1 << 15 // EOF detected
	EV_ERROR = 1 << 14 // error, data contains errno
)

type event struct {
	ident  uint64    // identifier for this event (e.g. file descriptor)
	filter int16     // filter for event (see EVFILT_*)
	flags  uint16    // general flags
	fflags uint32    // filter-specific flags
	data   int64     // filter-specific data
	udata  uint64    // opaque user data identifier
	ext    [2]uint64 // filter-specific extensions
}

func kevent64(kq int, cs []event, es []event, timeout *syscall.Timespec) (n int, err error) {
	var c0, e0 *event
	nc, ne := len(cs), len(es)
	if nc > 0 {
		c0 = &cs[0]
	}
	if ne > 0 {
		e0 = &es[0]
	}
	r, _, e := syscall.Syscall9(syscall.SYS_KEVENT64, uintptr(kq),
		uintptr(unsafe.Pointer(c0)), uintptr(nc),
		uintptr(unsafe.Pointer(e0)), uintptr(ne),
		uintptr(0),
		uintptr(unsafe.Pointer(timeout)),
		0, 0)
	n = int(r)
	if e != 0 {
		err = e
	}
	return
}

func kqueue() (fd int, err error) {
	r, _, e := syscall.RawSyscall(syscall.SYS_KQUEUE, 0, 0, 0)
	fd = int(r)
	if e != 0 {
		err = e
	}
	return
}

func (m *Mux) maybe_create() {
	m.once.Do(func() {
		var err error
		m.fd, err = kqueue()
		if err != nil {
			panic(fmt.Errorf("kqueue %s", err))
		}
	})
}

// Add adds a file to the file poller, certainly for read and possibly for write depending on f.WriteReady()
func (m *Mux) Add(f Filer) {
	m.poolLock.Lock()
	defer m.poolLock.Unlock()
	m.maybe_create()
	l := f.GetFile()
	fd := l.Fd
	if err := syscall.SetNonblock(fd, true); err != nil {
		panic(fmt.Errorf("setnonblock: %s", err))
	}

	fi := m.filePool.GetIndex()
	m.files[fi] = f
	l.poolIndex = fi

	var changes [2]event
	n_changes := 0
	if !l.disableRead {
		changes[n_changes] = event{
			ident:  uint64(l.Fd),
			filter: EVFILT_READ,
			flags:  EV_ADD,
			udata:  uint64(l.poolIndex),
		}
		n_changes++
	}

	changes[n_changes] = event{
		ident:  uint64(l.Fd),
		filter: EVFILT_WRITE,
		flags:  EV_ADD,
		udata:  uint64(l.poolIndex),
	}
	flags := EV_DISABLE
	if !l.disableWrite && f.WriteAvailable() {
		flags = EV_ENABLE
	}
	changes[n_changes].flags |= uint16(flags)
	n_changes++

	if _, err := kevent64(m.fd, changes[:n_changes], nil, nil); err != nil {
		panic(fmt.Errorf("kevent64: add %s", err))
	}
}

// Del removes the file (descriptor) from polling and frees file pool entry.
func (m *Mux) Del(f Filer) {
	m.poolLock.Lock()
	defer m.poolLock.Unlock()
	l := f.GetFile()

	changes := [1]event{
		event{
			ident:  uint64(l.Fd),
			filter: EVFILT_READ,
			flags:  EV_DELETE,
		},
	}
	if _, err := kevent64(m.fd, changes[:], nil, nil); err != nil {
		panic(fmt.Errorf("kevent64: del %s", err))
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

	changes := [1]event{
		event{
			ident:  uint64(l.Fd),
			filter: EVFILT_WRITE,
			flags:  EV_DISABLE,
		},
	}
	if f.WriteAvailable() {
		changes[0].flags = EV_ENABLE
	}
	if _, err := kevent64(m.fd, changes[:], nil, nil); err != nil {
		panic(fmt.Errorf("kevent64: update %s", err))
	}
}

func (m *Mux) do(e *event) {
	fi := uint(e.udata)

	// Deleted file?
	if m.files[fi] == nil {
		return
	}

	switch e.filter {
	case EVFILT_WRITE:
		err := m.files[fi].WriteReady()
		if err != nil {
			panic(err)
		}
	case EVFILT_READ:
		err := m.files[fi].ReadReady()
		if err != nil {
			panic(err)
		}
	}
}

func (m *Mux) EventPoll() {
	var events [256]event
	m.maybe_create()
	n, err := kevent64(m.fd, nil, events[:], nil)
	if err != nil {
		panic(fmt.Errorf("kevent64 wait %s", err))
	}
	for i := 0; i < n; i++ {
		m.do(&events[i])
	}
}
