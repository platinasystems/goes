// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || linux

package utmpx

/*
#include <utmpx.h>
*/
import "C"
import (
	"strings"
	"time"
)

type UserProcess struct {
	utx *C.struct_utmpx
}

func NewUserProcess(user, line, host string, pid int) UserProcess {
	now := time.Now()
	utx := &C.struct_utmpx{
		ut_pid:  C.int(pid),
		ut_type: C.USER_PROCESS,
		ut_tv: C.struct_timeval{
			tv_sec:  C.long(now.Truncate(time.Second).Unix()),
			tv_usec: C.int(now.Nanosecond() / 1000),
		},
	}
	n := len(utx.ut_user[:])
	for i, b := range []byte(user) {
		if i >= n {
			break
		}
		utx.ut_user[i] = C.char(b)
	}
	line = strings.TrimLeft(line, "/dev/")
	n = len(utx.ut_line[:])
	for i, b := range []byte(line) {
		if i >= n {
			break
		}
		utx.ut_line[i] = C.char(b)
	}
	n = len(utx.ut_id[:])
	if len(line) > n {
		line = line[len(line)-n:]
	}
	for i, b := range []byte(line) {
		utx.ut_id[i] = C.char(b)
	}
	n = len(utx.ut_host[:])
	for i, b := range []byte(host) {
		if i >= n {
			break
		}
		utx.ut_host[i] = C.char(b)
	}
	return UserProcess{C.pututxline(utx)}
}

func (up UserProcess) Died() {
	if up.utx != nil {
		up.utx.ut_type = C.DEAD_PROCESS
		C.pututxline(up.utx)
	}
}
