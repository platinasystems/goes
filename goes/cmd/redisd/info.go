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

	. "github.com/platinasystems/go"
	grs "github.com/platinasystems/go-redis-server"
	"github.com/platinasystems/go/internal/proc"
)

const SizeofInt = (32 << (^uint(0) >> 63)) >> 3

func (redisd *Redisd) Info(secs ...string) (*grs.StatusReply, error) {
	stat := new(proc.Stat)
	err := proc.Load(stat).FromFile("/proc/self/stat")
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if len(secs) == 0 || secs[0] == "default" || secs[0] == "all" {
		secs = []string{
			"server",
			"clients",
			"memory",
			"stats",
			"cpu",
		}
	}

	for _, sec := range secs {
		f, found := map[string]func(io.Writer){
			"server": func(w io.Writer) {
				fmt.Fprintln(w, "redis_git_sha1:",
					Package["version"])
				fmt.Fprintln(w, "redis_git_dirty:",
					len(Package["diff"]) > 0)
				fmt.Fprintln(w, "os:", runtime.GOOS)
				fmt.Fprintln(w, "arch_bits:", SizeofInt*8)
				fmt.Fprintln(w, "process_id:", stat.Pid)
				s := time.Now().Sub(stat.StartTime).Seconds()
				fmt.Fprintln(w, "uptime_in_seconds:",
					math.Floor(s+.5))
				fmt.Fprintln(w, "uptime_in_days:",
					math.Floor((s/(60*60*24))+.5))
			},
			"clients": func(w io.Writer) {
				fmt.Fprintln(w, "FIXME")
			},
			"memory": func(w io.Writer) {
				fmt.Fprintln(w, "used_memory:", stat.Vsize)
				fmt.Fprintln(w, "used_memory_rss:", stat.Rss)
			},
			"stats": func(w io.Writer) {
				fmt.Fprintln(w, "FIXME")
			},
			"cpu": func(w io.Writer) {
				fmt.Fprintln(w, "used_cpu_sys:", stat.Stime)
				fmt.Fprintln(w, "used_cpu_user:", stat.Utime)
				fmt.Fprintln(w, "used_cpu_sys_children:",
					stat.Cstime)
				fmt.Fprintln(w, "used_cpu_user_children:",
					stat.Cutime)
			},
		}[sec]
		if !found {
			err = fmt.Errorf("%s: unavailable", sec)
			break
		}
		if len(secs) > 1 {
			fmt.Fprintln(buf, "#", strings.Title(sec))
		}
		f(buf)
	}
	return grs.NewStatusReply(buf.String()), err
}
