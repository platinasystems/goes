// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package gopanicd

import "strings"

const Name = "go-panicd"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) Daemon() int    { return -1 }
func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " [MESSAGE]..." }

func (cmd) Main(args ...string) error {
	msg := "---"
	if len(args) > 0 {
		msg = strings.Join(args, " ")
	}
	stop := make(chan error)
	go func() {
		defer func() { stop <- nil }()
		panic(msg)
	}()
	return <-stop
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "test daemon error log from go-routine",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	go-panicd - test daemon error log from go-routine

SYNOPSIS
	go-panicd [MESSAGE]...

DESCRIPTION
	Print the given or default message to klog or syslog followed by
	go-routine trace.`,
	}
}
