// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package fdset

import (
	"sync"
	"syscall"
	"time"
)

const SelectDuration = 500 * time.Millisecond

var SelectTimeval = syscall.NsecToTimeval(SelectDuration.Nanoseconds())

var bits = sync.OnceValue(func() int {
	return syscall.FD_SETSIZE / len(syscall.FdSet{}.Bits)
})

type FdSet struct {
	syscall.FdSet
	nfds int
}

func (fdset *FdSet) Clear(fds ...int) {
	for _, fd := range fds {
		fdset.Bits[fd/bits()] &^= 1 << uint(fd%bits())
	}
}

func (fdset *FdSet) IsSet(fd int) bool {
	return (fdset.Bits[fd/bits()] & (1 << uint(fd%bits()))) != 0
}

func (fdset *FdSet) Set(fds ...int) {
	for _, fd := range fds {
		fdset.Bits[fd/bits()] |= 1 << uint(fd%bits())
		if fd > fdset.nfds {
			fdset.nfds = fd + 1
		}
	}
}

func (fdset *FdSet) Zero() {
	for i := range fdset.Bits {
		fdset.Bits[i] = 0
	}
}

type Selection struct {
	Read, Write, Err FdSet
}

func (sel *Selection) Select() error {
	var rfds, wfds, efds *syscall.FdSet
	var nfds int
	if sel.Read.nfds > 0 {
		rfds = &sel.Read.FdSet
		nfds = sel.Read.nfds
	}
	if sel.Write.nfds > 0 {
		wfds = &sel.Write.FdSet
		if nfds < sel.Write.nfds {
			nfds = sel.Write.nfds
		}
	}
	if sel.Err.nfds > 0 {
		efds = &sel.Err.FdSet
		if nfds < sel.Err.nfds {
			nfds = sel.Err.nfds
		}
	}
	return Select(nfds, rfds, wfds, efds, &SelectTimeval)
}
