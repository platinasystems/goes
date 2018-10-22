// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package log

import (
	"errors"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/log"
)

type Command struct{}

func (Command) String() string { return "log" }

func (Command) Usage() string {
	return "log [PRIORITY [FACILITY]] TEXT..."
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "print text to /dev/kmsg",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
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

func (Command) Main(args ...string) error {
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
