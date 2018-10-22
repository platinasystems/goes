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

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/redis"
	"github.com/platinasystems/redis/publisher"
)

type Command chan struct{}

func (Command) String() string { return "uptimed" }

func (Command) Usage() string { return "uptimed" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "record system uptime in redis",
	}
}

func (c Command) Close() error {
	close(c)
	return nil
}

func (Command) Kind() cmd.Kind { return cmd.Daemon }

func (c Command) Main(...string) error {
	err := redis.IsReady()
	if err != nil {
		return err
	}
	if err = update(); err != nil {
		return err
	}
	t := time.NewTicker(60 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-c:
			return nil
		case <-t.C:
			if err = update(); err != nil {
				return err
			}
		}
	}
	return nil
}

func update_uptime(si *syscall.Sysinfo_t, pub *publisher.Publisher) {
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
	pub.Print("sys.uptime: ", buf.String())
}

func update_cpu(si *syscall.Sysinfo_t, pub *publisher.Publisher) {
	const scale float64 = 65536.0 // magic SI_LOAD_SHIFT

	for _, v := range []struct {
		key  string
		load uint64
	}{
		{"sys.cpu.load1", uint64(si.Loads[0])},
		{"sys.cpu.load10", uint64(si.Loads[1])},
		{"sys.cpu.load15", uint64(si.Loads[2])},
	} {
		fload := float64(v.load) / scale
		pub.Print(fmt.Sprintf("%s: %2.2f", v.key, fload))
	}
}

func update_mem(si *syscall.Sysinfo_t, pub *publisher.Publisher) {
	for _, v := range []struct {
		key   string
		value uint64 // 32 bit on bmc
	}{
		{"sys.mem.total", uint64(si.Totalram)},
		{"sys.mem.free", uint64(si.Freeram)},
		{"sys.mem.shared", uint64(si.Sharedram)},
		{"sys.mem.buffer", uint64(si.Bufferram)},
	} {
		pub.Print(fmt.Sprintf("%s: %v", v.key, v.value))
	}
}

func update() error {
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

	update_uptime(&si, pub)
	update_cpu(&si, pub)
	update_mem(&si, pub)

	return nil
}
