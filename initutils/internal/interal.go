// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func AssertRoot() (err error) {
	if os.Geteuid() != 0 {
		err = errors.New("you aren't root")
	}
	return
}

func KillAll(sig syscall.Signal) (err error) {
	thisprog, err := os.Readlink("/proc/self/exe")
	if err != nil {
		return
	}
	thispid := os.Getpid()
	exes, err := filepath.Glob("/proc/*/exe")
	if err != nil {
		return
	}
	for _, exe := range exes {
		var pid int
		prog, e := os.Readlink(exe)
		if e != nil || prog != thisprog {
			continue
		}
		spid := strings.TrimPrefix(strings.TrimSuffix(exe, "/exe"),
			"/proc/")
		fmt.Sscan(spid, &pid)
		if pid == thispid {
			continue
		}
		_, e = os.Stat(fmt.Sprint("/proc/", spid, "/stat"))
		if e == nil {
			e = syscall.Kill(pid, syscall.SIGTERM)
			if err == nil {
				err = e
			}
		}
	}
	return
}
