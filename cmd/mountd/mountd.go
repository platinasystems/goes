// Copyright © 2018 Platina Systems, Inc. All rights reserved.
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

	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/internal/partitions"
	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/log"
)

var ErrUnknownPartition = errors.New("Unable to determine partition type")

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

func (Command) mountMap(mountFile string) (mm map[string]string, err error) {
	f, err := os.Open(mountFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if strings.HasPrefix(fields[0], "/dev/") {
			if mm == nil {
				mm = make(map[string]string)
			}
			if mm[fields[0]] == "" {
				mm[fields[0]] = fields[1]
			}
		}
	}
	return mm, scanner.Err()
}

func (Command) mountone(mm map[string]string, dev, dir string) (err error) {
	mp := mm[dev]
	if mp != "" {
		return syscall.Mount(mp, dir, "", uintptr(syscall.MS_BIND), "")
	}
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

func (c Command) mountall(mp string) {
	mm, err := c.mountMap("/proc/mounts")
	if err != nil {
		log.Printf("reading /proc/mounts: %s\n", err)
		return
	}
	pp, err := os.Open("/proc/partitions")
	if err != nil {
		log.Printf("opening /proc/partitions: %s", err)
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
				err := c.mountone(mm, "/dev/"+fileName, mpd)
				if err != nil && err != ErrUnknownPartition {
					fmt.Println("mount", mpd, "err:", err)
				}
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
