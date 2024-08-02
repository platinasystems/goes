// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build windows || plan9 || ((darwin || ios) && !cgo)

package xos

import (
	"io"
	"os"
)

type LogFile string

const PathSeparatorString = string(os.PathSeparator)

var (
	ErrorLogFile  LogFile = "goes_error.log"
	InfoLogFile   LogFile = "goes_info.log"
	NoticeLogFile LogFile = "goes.log"
)

var (
	OpenErrorLog  = ErrorLogFile.Create
	OpenInfoLog   = InfoLogFile.Create
	OpenNoticeLog = NoticeLogFile.Create
)

// If not super user and directory exists, create ~/Library/Logs/FN;
// otherwise, TMPDIR/FN.
func (lf LogFile) Create() (io.WriteCloser, error) {
	dn := os.TempDir()
	if os.Geteuid() > 0 {
		if h, err := os.UserHomeDir(); err == nil {
			h += PathSeparatorString + "Library"
			h += PathSeparatorString + "Logs"
			fi, err := os.Stat(h)
			if err == nil && fi.IsDir() {
				dn = h
			}
		}
	}
	return os.Create(dn + PathSeparatorString + string(lf))
}
