// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cmd

const (
	DontFork Kind = 1 << iota
	Daemon
	Hidden
	CantPipe
)

func WhatKind(v Cmd) Kind {
	if m, found := v.(kinder); found {
		return m.Kind()
	}
	return 0
}

type kinder interface {
	Kind() Kind
}

type Kind uint16

func (k Kind) IsDontFork() bool    { return (k & DontFork) == DontFork }
func (k Kind) IsDaemon() bool      { return (k & Daemon) == Daemon }
func (k Kind) IsHidden() bool      { return (k & Hidden) == Hidden }
func (k Kind) IsInteractive() bool { return (k & (Daemon | Hidden)) == 0 }
func (k Kind) IsCantPipe() bool    { return (k & CantPipe) == CantPipe }

func (k Kind) String() string {
	s := "unknown"
	switch k {
	case DontFork:
		s = "don't fork"
	case Daemon:
		s = "daemon"
	case Hidden:
		s = "hidden"
	}
	return s
}
