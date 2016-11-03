// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package uptime

import (
	"bytes"
	"fmt"
	"syscall"
	"time"

	"github.com/platinasystems/go/info"
)

const Name = "uptime"

type Info chan struct{}

func New() Info { return Info(make(chan struct{})) }

func (Info) String() string { return Name }

func (uptime Info) Main(...string) error {
	var si syscall.Sysinfo_t
	err := syscall.Sysinfo(&si)
	if err != nil {
		return err
	}
	info.Publish(Name, update())
	go uptime.ticker()
	return nil
}

func (uptime Info) Close() error {
	close(uptime)
	return nil
}

func (Info) Del(key string) error { return info.CantDel(key) }

func (Info) Prefixes(...string) []string { return []string{Name} }

func (Info) Set(key, value string) error { return info.CantSet(key) }

func (uptime Info) ticker() {
	t := time.NewTicker(60 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-uptime:
			return
		case <-t.C:
			info.Publish(Name, update())
		}
	}
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
