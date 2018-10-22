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
	"github.com/platinasystems/atsock"
	"github.com/platinasystems/redis/publisher"
)

const (
	LOGA          = "dmesg.txt"
	LOGB          = "dmesg2.txt"
	ENABLE        = "/tmp/mmclog_enable"
	MAXLEN        = 4096
	MAXMSG        = 50000
	MAXSIZE int64 = 512 * 1024 * 1024
	MMCDIR        = "/mnt"
)

type FileInfo struct {
	Name string
	Exst bool
	Size int64
	SeqN int64
}

type Command struct {
	Info
	Init func()
	init sync.Once
}

type Info struct {
	mutex   sync.Mutex
	rpc     *atsock.RpcServer
	pub     *publisher.Publisher
	stop    chan struct{}
	logA    string
	logB    string
	seq_end uint64
}

func (*Command) String() string { return "mmclogd" }

func (*Command) Usage() string { return "mmclogd" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "mmclogd - updater for MMC dmesg logging",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	mmclog daemon`,
	}
}

func (*Command) Kind() cmd.Kind { return cmd.Daemon }

func (c *Command) Main(...string) error {
	if c.Init != nil {
		c.init.Do(c.Init)
	}

	if err := initLogging(&c.Info); err != nil {
		return err
	}

	t := time.NewTicker(15 * time.Second)
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
	if _, err := os.Stat(ENABLE); os.IsNotExist(err) {
		return nil
	}

	if err := updateLogs(&c.Info); err != nil {
		return err
	}
	return nil
}
