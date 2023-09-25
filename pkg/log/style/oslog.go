// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build cgo && (darwin || ios)

package style

/*
#include <stdlib.h>
#include <os/log.h>

#define os_log_format "%{public}s"

const os_log_t os_log_default = OS_LOG_DEFAULT;

void asl_fault(const char* const s) {
	os_log_fault(os_log_default, os_log_format, s);
}

void asl_error(const char* const s) {
	os_log_error(os_log_default, os_log_format, s);
}

void asl_notice(const char* const s) {
	os_log(os_log_default, os_log_format, s);
}

void asl_info(const char* const s) {
	os_log_info(os_log_default, os_log_format, s);
}

void asl_debug(const char* const s) {
	os_log_debug(os_log_default, os_log_format, s);
}
*/
import "C"
import (
	"bytes"
	"unsafe"
)

type ASL struct{ Level }

func (asl ASL) Write(b []byte) (int, error) {
	cs := C.CString(string(bytes.TrimSpace(b)))
	defer C.free(unsafe.Pointer(cs))
	switch asl.Level {
	case Emergency:
		fallthrough
	case Alert:
		fallthrough
	case Critical:
		C.asl_fault(cs)
	case Errata:
		fallthrough
	case Warning:
		C.asl_error(cs)
	case Notice:
		C.asl_notice(cs)
	case Info:
		C.asl_info(cs)
	case Debug:
		C.asl_debug(cs)
	}
	return len(b), nil
}

// Log to Apple System Log instead of Std{out|err}.
func (style Style) System() {
	style.SetOutput(ASL{style.Level})
	style.SetPrefix("")
}
