// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package uptimed publishes the system uptime every 60 seconds to the local
// redis server.
package uptimed

import (
	"bytes"
	"fmt"
	"syscall"
	"time"

	"github.com/platinasystems/go/redis"
)

const Name = "uptimed"

type cmd chan struct{}

func New() cmd { return cmd(make(chan struct{})) }

func (cmd) Daemon() int    { return 1 }
func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name }

func (cmd cmd) Main(...string) error {
	var si syscall.Sysinfo_t
	err := syscall.Sysinfo(&si)
	if err != nil {
		return err
	}
	pub, err := redis.Publish(redis.Machine)
	if err != nil {
		return err
	}
	pub <- fmt.Sprint("uptime: ", update())
	t := time.NewTicker(60 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-cmd:
			return nil
		case <-t.C:
			pub <- fmt.Sprint("uptime: ", update())
		}
	}
	return nil
}

func (cmd cmd) Close() error {
	close(cmd)
	return nil
}

func update() string {
	var si syscall.Sysinfo_t
	if err := syscall.Sysinfo(&si); err != nil {
		return err.Error()
	}
	buf := new(bytes.Buffer)
	updecades := si.Uptime / (60 * 60 * 24 * 365 * 10)
	upyears := (si.Uptime / (60 * 60 * 24 * 365)) % 10
	upweeks := (si.Uptime / (60 * 60 * 24 * 7)) % 52
	updays := (si.Uptime / (60 * 60 * 24)) % 7
	upminutes := si.Uptime / 60
	uphours := upminutes / 60
	uphours = uphours % 24
	upminutes = upminutes % 60
	if si.Uptime < 60 {
		fmt.Fprint(buf, si.Uptime, " seconds")
	}
	if updecades > 0 {
		fmt.Fprint(buf, updecades, " decade")
		if updecades > 1 {
			fmt.Fprint(buf, "s")
		}
		fmt.Fprint(buf, ", ")
	}
	if upyears > 0 {
		fmt.Fprint(buf, upyears, " year")
		if upyears > 1 {
			fmt.Fprint(buf, "s")
		}
		fmt.Fprint(buf, ", ")
	}
	if upweeks > 0 {
		fmt.Fprint(buf, upweeks, " week")
		if upweeks > 1 {
			fmt.Fprint(buf, "s")
		}
		fmt.Fprint(buf, ", ")
	}
	if updays > 0 {
		fmt.Fprint(buf, updays, " day")
		if updays > 1 {
			fmt.Fprint(buf, "s")
		}
		fmt.Fprint(buf, ", ")
	}
	if uphours > 0 {
		fmt.Fprint(buf, uphours, " hour")
		if uphours > 1 {
			fmt.Fprint(buf, "s")
		}
		fmt.Fprint(buf, ", ")
	}
	if upminutes > 0 {
		fmt.Fprint(buf, upminutes, " minute")
		if upminutes > 1 {
			fmt.Fprint(buf, "s")
		}
	}
	return buf.String()
}
