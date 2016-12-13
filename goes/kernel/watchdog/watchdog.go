// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Watchdog is only run by an embedded machine's /init, not by
// /usr/bin/goes start
package watchdog

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/internal/parms"
)

const Name = "watchdog"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) Kind() goes.Kind {
	k := goes.Disabled
	if os.Getpid() == 1 {
		k = goes.Daemon
	}
	return k
}

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " [OPTION]... [DEVICE]" }

func (cmd) Main(args ...string) error {
	parm, args := parms.New(args, "-T", "-t")
	for k, v := range map[string]string{
		"-T": "60",
		"-t": "30",
	} {
		if len(parm[k]) == 0 {
			parm[k] = v
		}
	}

	period, err := strconv.ParseInt(parm["-t"], 0, 0)
	if err != nil {
		return fmt.Errorf("%s: invalid Period: ", err)
	} else if period <= 0 {
		return fmt.Errorf("%v: invalid period", period)
	}
	freq := time.Duration(period) * time.Second

	fn := "/dev/watchdog"
	if n := len(args); n > 0 {
		fn = args[0]
	} else if n > 1 {
		return fmt.Errorf("%v: unexpected", args[1:])
	}
	f, err := os.OpenFile(fn, os.O_WRONLY, 0)
	if err != nil {
		// Only an error if system has watchdog.
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	defer f.Close()

	ticker := time.NewTicker(time.Duration(freq))
	defer ticker.Stop()

	for _ = range ticker.C {
		n, err := f.Write([]byte{0})
		if err != nil {
			return err
		}
		if n != 1 {
			return io.ErrShortWrite
		}
	}
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "periodic write to device",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	watchdog - periodic write to device

SYNOPSIS
	watchdog [OPTION]... [DEVICE]

DESCRIPTION
	Periodically write to the watchdog device (default /dev/watchdog).

OPTIONS
	-T TIMEOUT	Reboot after TIMEOUT seconds without a watchdog write
			(default 60)
	-t FREQUENCY	Write frequency in seconds
			(default 30)`,
	}
}
