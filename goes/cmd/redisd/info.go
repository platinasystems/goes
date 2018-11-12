// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package redisd

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"runtime"
	"strings"
	"time"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/internal/proc"
)

const SizeofInt = (32 << (^uint(0) >> 63)) >> 3

func (redisd *Redisd) Info(secs ...string) ([]byte, error) {
	stat := new(proc.Stat)
	err := proc.Load(stat).FromFile("/proc/self/stat")
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if len(secs) == 0 || secs[0] == "default" || secs[0] == "all" {
		secs = []string{
			"server",
			"memory",
			"cpu",
		}
	}

	funcs := map[string]func(io.Writer){
		"server": func(w io.Writer) {
			fmt.Fprint(w, "version: ", goes.Version, "\r\n")
			fmt.Fprint(w, "os: ", runtime.GOOS, "\r\n")
			fmt.Fprint(w, "arch_bits: ", SizeofInt*8, "\r\n")
			fmt.Fprint(w, "process_id: ", stat.Pid, "\r\n")
			s := time.Now().Sub(stat.StartTime).Seconds()
			fmt.Fprint(w, "uptime_in_seconds: ", math.Floor(s+.5), "\r\n")
			fmt.Fprint(w, "uptime_in_days: ",
				math.Floor((s/(60*60*24))+.5),
				"\r\n")
		},
		"memory": func(w io.Writer) {
			fmt.Fprint(w, "used_memory: ",
				stat.Vsize, "\r\n")
			fmt.Fprint(w, "used_memory_rss: ",
				stat.Rss, "\r\n")
		},
		"cpu": func(w io.Writer) {
			fmt.Fprint(w, "used_cpu_sys: ",
				stat.Stime, "\r\n")
			fmt.Fprint(w, "used_cpu_user: ",
				stat.Utime, "\r\n")
			fmt.Fprint(w, "used_cpu_sys_children: ",
				stat.Cstime, "\r\n")
			fmt.Fprint(w, "used_cpu_user_children: ",
				stat.Cutime, "\r\n")
		},
	}
	for i, sec := range secs {
		f, found := funcs[sec]
		if !found {
			err = fmt.Errorf("%s: unavailable", sec)
			break
		}
		if len(secs) > 1 {
			if i > 0 {
				fmt.Fprint(buf, "\r\n")
			}
			fmt.Fprint(buf, "# ", strings.Title(sec), "\r\n")
		}
		f(buf)
	}
	return buf.Bytes(), err
}
