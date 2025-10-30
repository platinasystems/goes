// Copyright © 2021-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xlog

import (
	"io"
	"log"
	"os"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/xflag"
)

var (
	Errata = NewUnmutedLogger(os.Stderr)
	Info   = NewMutedLogger(os.Stderr)
	Trace  = NewMutedLogger(os.Stderr)
)

var (
	QuietFlag = xflag.Label{"quiet", "Log errata only.", func() error {
		Errata.Mute()
		return nil
	}}
	TraceFlag = xflag.Label{"trace", "Very verbose logging.", func() error {
		Trace.Unmute()
		return nil
	}}
	VerboseFlag = xflag.Label{"verbose", "Log info.", func() error {
		Info.Unmute()
		return nil
	}}
	Flags = xflag.Labels{QuietFlag, TraceFlag, VerboseFlag}
)

// SetFlags of [Errata], [Info], and [Trace].
// (default [log.Lshortfile])
func SetFlags(i int) {
	Errata.SetFlags(i)
	Info.SetFlags(i)
	Trace.SetFlags(i)
}

// SetPrefix of [Errata], [Info], and [Trace]
func SetPrefixes(s string) {
	Errata.SetPrefix(s)
	Info.SetPrefix(s)
	Trace.SetPrefix(s)
}

type Mutable struct {
	*log.Logger
	w     io.Writer
	mutex sync.Mutex
}

func NewMutedLogger(w io.Writer) *Mutable {
	return &Mutable{
		Logger: log.New(io.Discard, "", log.Lshortfile),
		w:      w,
	}
}

func NewUnmutedLogger(w io.Writer) *Mutable {
	return &Mutable{
		Logger: log.New(w, "", log.Lshortfile),
		w:      w,
	}
}

func (m *Mutable) Mute() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.SetOutput(io.Discard)
}

func (m *Mutable) Toggle() {
	var w io.Writer = io.Discard
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if w == m.Writer() {
		w = m.w
	}
	m.SetOutput(w)
}

func (m *Mutable) Unmute() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.SetOutput(m.w)
}

func (m *Mutable) Write(b []byte) (int, error) {
	return m.Writer().Write(b)
}
