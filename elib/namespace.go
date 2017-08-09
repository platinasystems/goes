// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux

package elib

import (
	"os"
	"sync"
	"syscall"
)

func setns(fd, nstype int) (e syscall.Errno) {
	const SYS_SETNS = 308 // fixme AMD64 linux specific
	_, _, e = syscall.Syscall(uintptr(SYS_SETNS), uintptr(fd), uintptr(nstype), uintptr(0))
	return
}

var withNamespaceMutex sync.Mutex

// WithNamespace performs given function in given namespace.
// {new,old}_ns_fd are open file descriptors to namespace "special" files.
func WithNamespace(new_ns_fd, old_ns_fd, namespace_type int, f func() (err error)) (err error, first_setns_errno syscall.Errno) {
	withNamespaceMutex.Lock()
	defer withNamespaceMutex.Unlock()

	change_ns := new_ns_fd != old_ns_fd

	// Move to new namespace.
	if change_ns {
		if e := setns(new_ns_fd, namespace_type); e != 0 {
			err = os.NewSyscallError("setns", error(e))
			first_setns_errno = e
			return
		}
	}

	err = f()

	// If we changed name space, return to old namespace.
	if change_ns {
		if e := setns(old_ns_fd, namespace_type); e != 0 {
			err = os.NewSyscallError("setns", error(e))
		}
	}

	return
}

// Performs given callback in default namespace (e.g. without setns).
func WithDefaultNamespace(f func() (err error)) (err error) {
	withNamespaceMutex.Lock()
	err = f()
	withNamespaceMutex.Unlock()
	return
}
