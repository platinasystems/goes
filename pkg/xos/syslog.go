// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build unix && !darwin && !ios

package xos

import (
	"io"
	"log/syslog"
)

func OpenErrorLog() (io.WriteCloser, error) {
	return openDaemonPrority(syslog.LOG_ERR)
}

func OpenNoticeLog() (io.WriteCloser, error) {
	return openDaemonPrority(syslog.LOG_NOTICE)
}

func OpenInfoLog() (io.WriteCloser, error) {
	return openDaemonPrority(syslog.LOG_INFO)
}

func openDaemonPrority(pri syslog.Priority) (io.WriteCloser, error) {
	return syslog.New(syslog.LOG_DAEMON|pri, "")
}
