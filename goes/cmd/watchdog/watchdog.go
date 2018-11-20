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
	"sync"
	"time"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/gpio"
)

type Command struct {
	GpioPin string
	Init    func()
	init    sync.Once
}

func (*Command) String() string { return "watchdog" }

func (*Command) Usage() string { return "watchdog [OPTION]... [DEVICE]" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "periodic write to device",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Periodically write to the watchdog device (default /dev/watchdog).

OPTIONS
	-T TIMEOUT	Reboot after TIMEOUT seconds without a watchdog write
			(default 60)
	-t FREQUENCY	Write frequency in seconds
			(default 30)`,
	}
}

func (*Command) Kind() cmd.Kind { return cmd.Daemon }

func (c *Command) Main(args ...string) error {
	if c.Init != nil {
		c.init.Do(c.Init)
	}
	parm, args := parms.New(args, "-T", "-t")
	for k, v := range map[string]string{
		"-T": "60",
		"-t": "30",
	} {
		if len(parm.ByName[k]) == 0 {
			parm.ByName[k] = v
		}
	}

	period, err := strconv.ParseInt(parm.ByName["-t"], 0, 0)
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
		if len(c.GpioPin) > 0 {
			pin, found := gpio.Pins[c.GpioPin]
			t, err := pin.Value()
			if found && err == nil {
				pin.SetValue(!t)
			}
		}

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
