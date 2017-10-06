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
	"github.com/platinasystems/go/internal/redis/publisher"
	"github.com/platinasystems/go/internal/sockfile"
)

const (
	Name    = "mmclogd"
	Apropos = "mmclogd - updater for MMC dmesg logging"
	Usage   = "mmclogd"
	Man     = `
DESCRIPTION
	mmclog daemon`

	LOGA          = "dmesg.txt"
	LOGB          = "dmesg2.txt"
	ENABLE        = "/tmp/mmclog_enable"
	MAXLEN        = 4096
	MAXMSG        = 50000
	MAXSIZE int64 = 512 * 1024 * 1024
	MMCDIR        = "/tmp" //FIXME /mnt once mount is working
)

type FileInfo struct {
	Name string
	Exst bool
	Size int64
	SeqN int64
}

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
	mutex   sync.Mutex
	rpc     *sockfile.RpcServer
	pub     *publisher.Publisher
	stop    chan struct{}
	logA    string
	logB    string
	seq_end uint64
}

func New() *Command { return new(Command) }

func (*Command) Apropos() lang.Alt { return apropos }
func (*Command) Kind() cmd.Kind    { return cmd.Daemon }
func (*Command) String() string    { return Name }
func (*Command) Usage() string     { return Usage }

func (c *Command) Main(...string) error {
	once.Do(Init)
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
