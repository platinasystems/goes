// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netns

import (
	"fmt"
	"path/filepath"
	"syscall"
)

func Switch(name string) error {
	fn := filepath.Join("/var/run/netns", name)
	fd, err := syscall.Open(fn, syscall.O_RDONLY|syscall.O_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("%s: %v", fn, err)
	}
	err = setns(fd, syscall.CLONE_NEWNET)
	syscall.Close(fd)
	if err != nil {
		return fmt.Errorf("setns %s: %v", fn, err)
	}
	if err = syscall.Unshare(syscall.CLONE_NEWNS); err != nil {
		return err
	}
	// Don't let subsequent mounts propagate to the parent
	err = syscall.Mount("", "/", "none",
		syscall.MS_SLAVE|syscall.MS_REC, "")
	if err != nil {
		return err
	}
	// Mount a version of /sys that describes the network namespace
	if err = syscall.Unmount("/sys", syscall.MNT_DETACH); err != nil {
		return err
	}
	if err = syscall.Mount(fn, "/sys", "sysfs", 0, ""); err != nil {
		return err
	}
	return nil
}

func setns(fd, nstype int) (err error) {
	_, _, errno := syscall.Syscall(uintptr(SYS_SETNS), uintptr(fd),
		uintptr(nstype), uintptr(0))
	if errno != 0 {
		err = errno
	}
	return
}
