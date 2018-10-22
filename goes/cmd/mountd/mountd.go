// Copyright © 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mountd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/log"
	"github.com/platinasystems/go/internal/partitions"
)

type Command chan struct{}

func (Command) String() string { return "mountd" }

func (Command) Usage() string { return "mountd" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "mount daemon",
	}
}

func (c Command) Close() error {
	close(c)
	return nil
}

func (Command) Kind() cmd.Kind { return cmd.Daemon }

func mountone(dev, dir string) (err error) {
	sb, err := partitions.ReadSuperBlock(dev)
	if err != nil {
		return err
	}
	t := ""
	if sb != nil {
		t = sb.Kind()
	}
	if t == "" {
		return fmt.Errorf("Unable to determine type of block device on %s", dev)
	}
	err = syscall.Mount(dev, dir, t, 0, "")
	if err == nil {
		return nil
	}
	if err == syscall.EACCES {
		err = syscall.Mount(dev, dir, t, syscall.MS_RDONLY, "")
	}
	return err
}

func (Command) mountall(mp string) {
	pp, err := os.Open("/proc/partitions")
	if err != nil {
		log.Print("opening /proc/partitions: %s\n", err)
		return
	}
	defer pp.Close()
	scanner := bufio.NewScanner(pp)
scan:
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 4 || fields[0] == "major" {
			continue scan
		}
		fileName := fields[3]
		mpd := mp + "/" + fileName
		if _, err := os.Stat(mpd); os.IsNotExist(err) {
			err := os.Mkdir(mpd, os.FileMode(0555))
			if err != nil {
				fmt.Println("mkdir", mpd, "err:", err)
			} else {
				_ = mountone("/dev/"+fileName, mpd)
			}
		}
	}
}

func (c Command) Main(args ...string) (err error) {
	mp := "/mountd"
	if len(args) > 0 {
		mp = args[0]
	}
	if _, err = os.Stat(mp); os.IsNotExist(err) {
		err = os.Mkdir(mp, 0755)
		if err != nil {
			log.Print("err", mp, ": ", err)
		}
	}
	c.mountall(mp)
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		c.mountall(mp)
		select {
		case <-c:
			return nil
		case <-t.C:
		}
	}
	return nil
}
