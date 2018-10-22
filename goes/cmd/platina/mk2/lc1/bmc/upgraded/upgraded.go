// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package upgraded

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/log"
)

const (
	nl = "\n"
	sp = " "
	lt = "<"
	gt = ">"
	lb = "["
	rb = "]"
)

var BootedQSPI int = 3

type Command struct {
	Info
	Init func()
	init sync.Once
}

type Info struct {
	stop  chan struct{}
	last  map[string]uint16
	lasts map[string]string
}

func (*Command) String() string { return "upgraded" }

func (*Command) Usage() string { return "upgraded" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "upgraded - software updater",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	upgraded daemon`,
	}
}

func (*Command) Kind() cmd.Kind { return cmd.Daemon }

func (c *Command) Main(...string) error {
	if c.Init != nil {
		c.init.Do(c.Init)
	}

	getBootedQSPI()

	t := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-c.stop:
			return nil
		case <-t.C:
			if err := c.update(); err != nil {
			}
		}
	}
	return nil
}

func (c *Command) Close() error {
	close(c.stop)
	return nil
}

func (c *Command) update() error {
	return nil
}

func getBootedQSPI() {
	var kmsg log.Kmsg
	f, err := os.Open("/dev/kmsg")
	if err != nil {
		return
	}
	defer f.Close()
	if err = syscall.SetNonblock(int(f.Fd()), true); err != nil {
		return
	}
	buf := make([]byte, 4096)
	defer func() { buf = buf[:0] }()
	var si syscall.Sysinfo_t
	if err = syscall.Sysinfo(&si); err != nil {
		return
	}
	fo, err := os.Create("/tmp/qspi")
	if err != nil {
		return
	}
	defer f.Close()
	for i := 0; i < 400; i++ {
		n, err := f.Read(buf)
		if err != nil {
			break
		}
		kmsg.Parse(buf[:n])
		ksq := strconv.Itoa(int(kmsg.Seq))
		kst := strconv.Itoa(int(kmsg.Stamp))
		fs := ksq + sp + lb + kst + rb + sp + kmsg.Msg + nl
		if strings.Contains(fs, "Booted from QSPI") {
			_, err = fo.Write([]byte(fs))
			fo.Sync()
			return
		}
	}
	fo.Sync()
	return
}
