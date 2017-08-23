// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mmclogd

import (
	"os"
	"sync"
	"time"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "mmclogd"
	Apropos = "mmclogd - updater for MMC dmesg logging"
	Usage   = "mmclogd"
	Man     = `
DESCRIPTION
	mmclog daemon`
)

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}

var (
	Init = func() {}
	once sync.Once
)

type Command struct {
	Info
}

type Info struct {
	stop  chan struct{}
	last  map[string]uint16
	lasts map[string]string
}

func New() *Command { return new(Command) }

func (*Command) Apropos() lang.Alt { return apropos }
func (*Command) Kind() cmd.Kind    { return cmd.Daemon }
func (*Command) String() string    { return Name }
func (*Command) Usage() string     { return Usage }

func (c *Command) Main(...string) error {
	once.Do(Init)
	if err := InitLogging(); err != nil {
		return err
	}

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
	if _, err := os.Stat("/tmp/mmclog_enable"); os.IsNotExist(err) {
		return nil
	}

	if err := UpdateMMC(); err != nil {
		return err
	}
	return nil
}
