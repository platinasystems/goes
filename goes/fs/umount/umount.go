// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package umount

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/platinasystems/go/flags"
)

const Name = "umount"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " [OPTION]... FILESYSTEM|DIR" }

func (cmd) Main(args ...string) error {
	var err error
	flag, args := flags.New(args, "--fake", "-v", "-a", "-r", "-l", "-f",
		"-donot-free-loop-device")
	if flag["-a"] {
		err = umountall(flag)
	} else {
		if n := len(args); n == 0 {
			err = fmt.Errorf("FILESYSTEM or DEVICE: missing")
		} else if n > 1 {
			err = fmt.Errorf("%v: unexpected", args[1:])
		} else {
			err = umountone(args[0], flag)
		}
	}
	return err
}

// Unmount all filesystems in reverse order of /proc/mounts
func umountall(flag flags.Flag) error {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return err
	}
	defer f.Close()
	var targets []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		targets = append(targets, fields[1])
	}
	if err = scanner.Err(); err != nil {
		return err
	}
	for i := len(targets) - 1; i >= 0; i-- {
		if targets[i] == "/" && i != 0 {
			continue
		}
		terr := umountone(targets[i], flag)
		if terr != nil && err == nil {
			err = terr
		}
	}
	return err
}

func umountone(target string, flag flags.Flag) error {
	var flags int
	if flag["-l"] {
		flags |= syscall.MNT_DETACH
	}
	if flag["-f"] {
		flags |= syscall.MNT_FORCE
	}
	if flag["--fake"] {
		fmt.Println("Would umount", target)
		return nil
	}
	if err := syscall.Unmount(target, flags); err != nil {
		return err
	}
	if flag["-v"] {
		fmt.Println("Unmounted", target)
	}
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "deactivate filesystems",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	umount [OPTION]... FILESYSTEM|DIRECTORY

DESCRIPTION
	Deactivate file systems

OPTIONS
	--fake
	-v		verbose
	-a		all
	-r		Try to remount devices as read-only if mount is busy
	-l		Lazy umount (detach filesystem)
	-f		Force umount from unreachable NFS server
	-donot-free-loop-device`,
	}
}
