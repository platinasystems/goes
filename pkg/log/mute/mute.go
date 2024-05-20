// Copyright © 2021-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package mute provides log.Logger encapsulation with no-op Print and Write
// methods.
package mute

import "io"

// True if muted.
func IsOn(printer Printer) bool {
	_, ok := printer.(Mute)
	return ok
}

type Logger interface {
	Writer() io.Writer
	Printer
}

type Mute struct{ printer Printer }

func (Mute) Print(v ...any)                 {}
func (Mute) Printf(format string, v ...any) {}
func (Mute) Println(v ...any)               {}
func (Mute) Write(b []byte) (int, error)    { return len(b), nil }

type Printer interface {
	Print(v ...any)
	Printf(format string, v ...any)
	Println(v ...any)
}

type Unmute struct{ Logger }

func (um Unmute) Write(b []byte) (int, error) { return um.Writer().Write(b) }

// e.g. Mute or Unmute
type WritePrinter interface {
	io.Writer
	Printer
}

// Returns an unmuted logger.
func Off(printer Printer) WritePrinter {
	if um, ok := printer.(Unmute); ok {
		return um
	}
	if m, ok := printer.(Mute); ok {
		return Off(m.printer)
	}
	return Unmute{printer.(Logger)}
}

// Returns a muted logger.
func On(printer Printer) WritePrinter {
	if m, ok := printer.(Mute); ok {
		return m
	}
	if um, ok := printer.(Unmute); ok {
		return On(um.Logger)
	}
	return Mute{printer}
}
