// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package log

import (
	"errors"

	"github.com/platinasystems/go/log"
)

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return "log" }
func (cmd) Usage() string  { return "log [PRIORITY [FACILITY]] TEXT..." }

func (cmd) Main(args ...string) error {
	if len(args) == 0 {
		return errors.New("TEXT: missing")
	}
	argv := make([]interface{}, 0, len(args))
	defer func() { argv = argv[:0] }()
	for _, s := range args {
		argv = append(argv, s)
	}
	log.Print(argv...)
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print text to /dev/kmsg",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	log - print text to /dev/kmsg

SYNOPSIS
	log [PRIORITY [FACILITY]] TEXT...

DESCRIPTION
	Logged text may be viewed with 'dmesg' command.

PRIORITIES
	emerg, alert, crit, err, warn, note, info, debug

	The default priority is: info.

FACILITIES
	kern, user, mail, daemon, auth, syslog, lpr, news, uucp, cron, priv,
	ftp, local0, local1, local2, local3, local4, local5, local6, local7

	The default priority is: user.`,
	}
}
