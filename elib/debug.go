package elib

import (
	"runtime"
)

// Name of current function as string.
func FuncName() (n string) {
	if pc, _, _, ok := runtime.Caller(1); ok {
		n = runtime.FuncForPC(pc).Name()
	} else {
		n = "unknown"
	}
	return
}
