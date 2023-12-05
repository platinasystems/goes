// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build unix && !darwin && !ios

package oslog

import (
	"io"
	"log/syslog"
)

func OpenError() (io.WriteCloser, error) {
	return openDaemonPrority(syslog.LOG_ERR)
}

func OpenNotice() (io.WriteCloser, error) {
	return openDaemonPrority(syslog.LOG_NOTICE)
}

func OpenInfo() (io.WriteCloser, error) {
	return openDaemonPrority(syslog.LOG_INFO)
}

func openDaemonPrority(pri syslog.Priority) (io.WriteCloser, error) {
	return syslog.New(syslog.LOG_DAEMON|pri, "")
}
