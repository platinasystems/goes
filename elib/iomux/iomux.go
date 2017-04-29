// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iomux

import (
	"sync"
)

type Mux struct {
	// Poll/epoll file descriptor.
	fd            int
	once          sync.Once
	poolLock      sync.Mutex // protects following
	errorLog      [1024]error
	errorLogIndex int
	filePool
}

type File struct {
	Fd           int
	disableWrite bool
	disableRead  bool
	poolIndex    uint
}

func (f *File) GetFile() *File { return f }
func (f *File) SetWriteOnly()  { f.disableRead = true }
func (f *File) SetReadOnly()   { f.disableWrite = true }
func (f *File) Index() uint    { return f.poolIndex }

type Filer interface {
	GetFile() *File
	// OS indicates that file is ready to read and/or write.
	ReadReady() error
	WriteReady() error
	ErrorReady() error
	// User has data available to write to file.
	WriteAvailable() bool
	// Stringer for logging.
	String() string
}

//go:generate gentemplate -d Package=iomux -id file -d Data=files -d PoolType=filePool -d Type=Filer github.com/platinasystems/go/elib/pool.tmpl

var Default = &Mux{}

func Add(f Filer)    { Default.Add(f) }
func Del(f Filer)    { Default.Del(f) }
func Update(f Filer) { Default.Update(f) }
func Wait(once bool) { Default.Wait(once) }

func (m *Mux) Wait(once bool) {
	for {
		m.EventPoll()
		if once {
			break
		}
	}
}

func (m *Mux) logError(err error) {
	m.errorLog[m.errorLogIndex%len(m.errorLog)] = err
	m.errorLogIndex++
}
