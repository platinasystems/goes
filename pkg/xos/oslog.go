// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build cgo && (darwin || ios)

package xos

/*
#include <stdlib.h>
#include <os/log.h>

#define os_log_format "%{public}s"

const os_log_t os_log_default = OS_LOG_DEFAULT;

void oslog_fault(const char* const s) {
	os_log_fault(os_log_default, os_log_format, s);
}

void oslog_error(const char* const s) {
	os_log_error(os_log_default, os_log_format, s);
}

void oslog_notice(const char* const s) {
	os_log(os_log_default, os_log_format, s);
}

void oslog_info(const char* const s) {
	os_log_info(os_log_default, os_log_format, s);
}

void oslog_debug(const char* const s) {
	os_log_debug(os_log_default, os_log_format, s);
}
*/
import "C"
import (
	"bytes"
	"io"
	"unsafe"
)

type oslog uint8

const (
	oslogError oslog = iota
	oslogNotice
	oslogInfo
)

func OpenErrorLog() (io.WriteCloser, error) {
	return oslogError, nil
}

func OpenNoticeLog() (io.WriteCloser, error) {
	return oslogNotice, nil
}

func OpenInfoLog() (io.WriteCloser, error) {
	return oslogInfo, nil
}

func (oslog) Close() error {
	return nil
}

func (oslog oslog) Write(b []byte) (int, error) {
	cs := C.CString(string(bytes.TrimSpace(b)))
	defer C.free(unsafe.Pointer(cs))
	switch oslog {
	case oslogError:
		C.oslog_error(cs)
	case oslogNotice:
		C.oslog_notice(cs)
	case oslogInfo:
		C.oslog_info(cs)
	}
	return len(b), nil
}
