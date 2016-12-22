// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package gopanic

import "strings"

const Name = "go-panic"

type cmd struct{}

func New() cmd { return cmd{} }

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
		"en_US.UTF-8": "test error output from go-routine",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	go-panic - test error outpu from go-routine

SYNOPSIS
	go-panic [MESSAGE]...

DESCRIPTION
	Print the given or default message to standard error followed by
	go-routine trace and exit 2.`,
	}
}
