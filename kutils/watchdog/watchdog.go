// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package watchdog

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/platinasystems/go/parms"
)

type watchdog struct{}

func New() watchdog { return watchdog{} }

func (watchdog) String() string { return "watchdog" }
func (watchdog) Usage() string  { return "watchdog [OPTION]... [DEVICE]" }

func (watchdog) Daemon() int {
	lvl := -1 // don't run from /usr/sbin/goesd
	if os.Getpid() == 1 {
		lvl = 0
	}
	return lvl
}

func (watchdog) Main(args ...string) error {
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

func (watchdog) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "periodic write to device",
	}
}

func (watchdog) Man() map[string]string {
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
