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

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/redis/publisher"
)

const (
	Name    = "uptimed"
	Apropos = "record system uptime in redis"
	Usage   = "uptimed"
)

type Interface interface {
	Apropos() lang.Alt
	Close() error
	Kind() goes.Kind
	Main(...string) error
	String() string
	Usage() string
}

func New() Interface { return cmd(make(chan struct{})) }

type cmd chan struct{}

func (cmd cmd) Apropos() lang.Alt { return apropos }

func (cmd cmd) Close() error {
	close(cmd)
	return nil
}

func (cmd) Kind() goes.Kind { return goes.Daemon }

func (cmd cmd) Main(...string) error {
	var si syscall.Sysinfo_t
	err := syscall.Sysinfo(&si)
	if err != nil {
		return err
	}

	pub, err := publisher.New()
	if err != nil {
		return err
	}
	defer pub.Close()

	pub.Print("uptime: ", update())
	t := time.NewTicker(60 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-cmd:
			return nil
		case <-t.C:
			pub.Print("uptime: ", update())
		}
	}
	return nil
}

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

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

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
