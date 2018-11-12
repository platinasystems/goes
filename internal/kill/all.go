// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package kill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Signal all processes with /proc/self/exe -> this program.
func All(sig syscall.Signal) (err error) {
	thisprog, err := os.Readlink("/proc/self/exe")
	if err != nil {
		return
	}
	thispid := os.Getpid()
	exes, err := filepath.Glob("/proc/[0-9]*/exe")
	if err != nil {
		return
	}
	var g_pids []string
	for _, exe := range exes {
		var pid int
		prog, e := os.Readlink(exe)
		if e != nil {
			continue
		}
		if prog != thisprog && filepath.Dir(prog) != "/usr/lib/goes" {
			continue
		}
		spid := strings.TrimPrefix(strings.TrimSuffix(exe, "/exe"),
			"/proc/")
		n, e := fmt.Sscan(spid, &pid)
		if n != 1 || e != nil || pid == thispid || pid == 0 ||
			pid == 1 {
			continue
		}
		_, e = os.Stat(fmt.Sprint("/proc/", spid, "/stat"))
		if e == nil {
			g_pids = append(g_pids, exe)
			e = syscall.Kill(pid, sig)
			if e != nil && err == nil {
				err = fmt.Errorf("%s %d: %v", sig, pid, e)
			}
		}

	}

	// verify that goes pids are gone
	for _, pid := range g_pids {
		err = func(proc string) error {
			for j := 0; j < 60; j++ {
				_, e := os.Stat(proc)
				if e != nil {
					return nil
				}
				time.Sleep(time.Second)
			}
			return fmt.Errorf("failed to kill %s", pid)
		}(pid)
	}
	return
}
