// Copyright © 2021-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xlog

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/platinasystems/goes/v2/pkg/xcontext"
	"github.com/platinasystems/goes/v2/pkg/xsignal"
)

var (
	ErrLog = log.New(os.Stderr, "", log.Lshortfile)
	OutLog = log.New(os.Stdout, "", log.Lshortfile)
	Errata = Unmute(ErrLog)
	Info   = Mute(OutLog)
	Trace  = Mute(OutLog)
)

func SetPrefixes(s string) {
	ErrLog.SetPrefix(s)
	OutLog.SetPrefix(s)
}

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

// Toggle a [Printer]'s [Mute]/[Unmute].
func ToggleMute(printer Printer) WritePrinter {
	if m, ok := printer.(Muted); ok {
		return Unmute(m.printer)
	}
	if um, ok := printer.(Unmuted); ok {
		return Mute(um.Mutable)
	}
	return Unmuted{printer.(Mutable)}
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

// Toggle [Info] [Mute] on recept of [xsignal.Alarm].
func AlarmHandler(ctx context.Context) {
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, xsignal.Alarm)
	xcontext.Range(ctx, ch, func(sig os.Signal) bool {
		if sig != xsignal.Alarm {
			Errata.Println("unexpected", sig)
			return false
		}
		if m, ok := Info.(Muted); ok {
			Info = Unmute(m)
			Info.Println("enable info")
		} else if um, ok := Info.(Unmuted); ok {
			Info.Println("disable info")
			Info = Mute(um.Mutable)
		}
		if um, ok := Trace.(Unmuted); ok {
			Trace.Println("disable trace")
			Trace = Mute(um.Mutable)
		}
		return true
	})
}

func MuteErrata()  { Errata = Mute(Errata) }
func UnmuteInfo()  { Info = Unmute(Info) }
func UnmuteTrace() { Trace = Unmute(Trace) }
