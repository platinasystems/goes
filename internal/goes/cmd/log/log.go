// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package log

import (
	"errors"

	"github.com/platinasystems/go/internal/goes/lang"
	"github.com/platinasystems/go/internal/log"
)

const (
	Name    = "log"
	Apropos = "print text to /dev/kmsg"
	Usage   = "log [PRIORITY [FACILITY]] TEXT..."
	Man     = `
DESCRIPTION
	Logged text may be viewed with 'dmesg' command.

PRIORITIES
	emerg, alert, crit, err, warn, note, info, debug

	The default priority is: info.

FACILITIES
	kern, user, mail, daemon, auth, syslog, lpr, news, uucp, cron, priv,
	ftp, local0, local1, local2, local3, local4, local5, local6, local7

	The default priority is: user.`
)

type Interface interface {
	Apropos() lang.Alt
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }

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

func (cmd) Man() lang.Alt  { return man }
func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}

	man = lang.Alt{
		lang.EnUS: Man,
	}
)
