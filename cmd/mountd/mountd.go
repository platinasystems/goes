// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mountd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/external/log"
	"github.com/platinasystems/goes/internal/partitions"
	"github.com/platinasystems/goes/lang"
)

var ErrUnknownPartition = errors.New("Unable to determine partition type")

type Command struct {
	partitions map[string]error
}

func (*Command) String() string { return "mountd" }

func (*Command) Usage() string { return "mountd" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "mount daemon",
	}
}

func (*Command) Kind() cmd.Kind { return cmd.Daemon }

func (*Command) mountone(dev, dir string) (err error) {
	sb, err := partitions.ReadSuperBlock(dev)
	if err != nil {
		return err
	}
	t := ""
	if sb != nil {
		t = sb.Kind()
	}
	if t == "" {
		return ErrUnknownPartition
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

func (c *Command) mountall(mp string) {
	pp, err := os.Open("/proc/partitions")
	if err != nil {
		log.Printf("opening /proc/partitions: %s", err)
		return
	}
	defer pp.Close()
	partMap := make(map[string]bool)
	partScanner := bufio.NewScanner(pp)

	for partScanner.Scan() {
		fields := strings.Fields(partScanner.Text())
		if len(fields) < 4 || fields[0] == "major" {
			continue
		}
		partMap[fields[3]] = true
	}

	pm, err := os.Open("/proc/mounts")
	if err != nil {
		fmt.Printf("opening /proc/mounts: %s\n", err)
		return
	}
	defer pm.Close()

	mountMap := make(map[string]string)
	mountScanner := bufio.NewScanner(pm)
	for mountScanner.Scan() {
		fields := strings.Fields(mountScanner.Text())
		if len(fields) < 6 {
			continue
		}
		mountMap[fields[0]] = fields[1]
	}

	for part := range partMap {
		mpd := mp + "/" + part
		if _, found := c.partitions[mpd]; !found {
			if _, err := os.Stat(mpd); os.IsNotExist(err) {
				err := os.Mkdir(mpd, os.FileMode(0555))
				if err != nil {
					fmt.Println("mkdir", mpd, "err:", err)
					continue
				}
			}
			err := c.mountone("/dev/"+part, mpd)
			if err == nil {
				fmt.Println("Mounted", "/dev/"+part,
					mpd)
			} else {
				if err != ErrUnknownPartition {
					fmt.Println("mount", mpd, "err:", err)
				}
			}
			c.partitions[mpd] = err
		} else {
			if mountMap["/dev/"+part] != mpd &&
				c.partitions[mpd] == nil {
				delete(c.partitions, mpd)
			}
		}
	}
}

func (c *Command) unmountall(mp string) {
	for mpd, e := range c.partitions {
		if e == nil {
			err := syscall.Unmount(mpd, syscall.MNT_DETACH)
			if err != nil {
				fmt.Printf("Error unmounting %s: %s\n", mpd, err)
			}
		}
		delete(c.partitions, mpd)
	}
}

func (c *Command) Main(args ...string) (err error) {
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
	c.partitions = make(map[string]error)
	c.mountall(mp)
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		c.mountall(mp)
		select {
		case <-goes.Stop:
			c.unmountall(mp)
			return nil
		case <-t.C:
		}
	}
	return nil
}
