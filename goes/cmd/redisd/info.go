// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package redisd

import (
	"fmt"
	"io"
	"math"
	"os"
	"runtime"
	"time"

	. "github.com/platinasystems/go"
	"github.com/platinasystems/go/internal/proc"
)

const SizeofInt = (32 << (^uint(0) >> 63)) >> 3

func (redisd *Redisd) infoServer(w io.Writer) error {
	var stat proc.Stat
	if f, err := os.Open("/proc/self/stat"); err != nil {
		return err
	} else {
		err = stat.ReadFrom(f)
		f.Close()
		if err != nil {
			return err
		}
	}
	fmt.Fprintln(w, "redis_git_sha1:", Package["version"])
	fmt.Fprintln(w, "redis_git_dirty:", len(Package["diff"]) > 0)
	fmt.Fprintln(w, "os:", runtime.GOOS)
	fmt.Fprintln(w, "arch_bits:", SizeofInt*8)
	fmt.Fprintln(w, "process_id:", stat.Pid)
	seconds := time.Now().Sub(stat.StartTime).Seconds()
	fmt.Fprintln(w, "uptime_in_seconds:", math.Floor(seconds+.5))
	fmt.Fprintln(w, "uptime_in_days:", math.Floor((seconds/(60*60*24))+.5))
	return nil
}

func (redisd *Redisd) infoClients(w io.Writer) error {
	_, err := fmt.Fprintln(w, "FIXME")
	return err
}

func (redisd *Redisd) infoMemory(w io.Writer) error {
	_, err := fmt.Fprintln(w, "FIXME")
	return err
}

func (redisd *Redisd) infoStats(w io.Writer) error {
	_, err := fmt.Fprintln(w, "FIXME")
	return err
}

func (redisd *Redisd) infoCpu(w io.Writer) error {
	_, err := fmt.Fprintln(w, "FIXME")
	return err
}
