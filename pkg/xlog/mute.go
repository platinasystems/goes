// Copyright © 2021-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xlog

import "io"

func IsMuted(printer Printer) bool {
	_, ok := printer.(Muted)
	return ok
}

type Mutable interface {
	Writer() io.Writer
	Printer
}

// Muted [Logger] encapsulation with no-op Print and Write methods.
type Muted struct{ printer Printer }

func (Muted) Print(v ...any)                 {}
func (Muted) Printf(format string, v ...any) {}
func (Muted) Println(v ...any)               {}
func (Muted) Write(b []byte) (int, error)    { return len(b), nil }

type Printer interface {
	Print(v ...any)
	Printf(format string, v ...any)
	Println(v ...any)
}

// Unmuted [Logger] encapsulation with an additional Write method.
type Unmuted struct{ Mutable }

func (um Unmuted) Write(b []byte) (int, error) { return um.Writer().Write(b) }

// either Muted or Unmuted
type WritePrinter interface {
	io.Writer
	Printer
}

// Mute encapsulates an [Unmuted] [Printer] or [Logger] with no-op methods.
func Mute(printer Printer) WritePrinter {
	if m, ok := printer.(Muted); ok {
		return m
	}
	if um, ok := printer.(Unmuted); ok {
		return Mute(um.Mutable)
	}
	return Muted{printer}
}

// Unmute either de-encapsulates a [Muted] [Printer] or encapsulates a [Logger]
// with a Write method.
func Unmute(printer Printer) WritePrinter {
	if um, ok := printer.(Unmuted); ok {
		return um
	}
	if m, ok := printer.(Muted); ok {
		return Unmute(m.printer)
	}
	return Unmuted{printer.(Mutable)}
}
