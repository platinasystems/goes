// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package umount

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

type umount struct {
	flags *flags.Flags
}

func (Command) String() string { return "umount" }

func (Command) Usage() string {
	return "umount [OPTION]... FILESYSTEM|DIR"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "deactivate filesystems",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
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

func (Command) Main(args ...string) error {
	var err error
	var umount umount
	umount.flags, args = flags.New(args, "--fake", "-v", "-a", "-r", "-l",
		"-f", "-donot-free-loop-device")
	if umount.flags.ByName["-a"] {
		err = umount.all()
	} else {
		if n := len(args); n == 0 {
			err = fmt.Errorf("FILESYSTEM or DEVICE: missing")
		} else if n > 1 {
			err = fmt.Errorf("%v: unexpected", args[1:])
		} else {
			err = umount.one(args[0])
		}
	}
	return err
}

// Unmount all filesystems in reverse order of /proc/mounts
func (umount *umount) all() error {
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
	systemDir := map[string]struct{}{
		"/":        {},
		"/tmp":     {},
		"/proc":    {},
		"/dev":     {},
		"/dev/pts": {},
		"/sys":     {},
		"/run":     {},
	}

	for i := len(targets) - 1; i >= 0; i-- {
		if _, x := systemDir[targets[i]]; !x {
			terr := umount.one(targets[i])
			if terr != nil {
				fmt.Printf("Error unmounting %s: %s\n", targets[i],
					terr)
				if err == nil {
					err = terr
				}
			}
		}
	}
	return err
}

func (umount *umount) one(target string) error {
	var flags int
	if umount.flags.ByName["-l"] {
		flags |= syscall.MNT_DETACH
	}
	if umount.flags.ByName["-f"] {
		flags |= syscall.MNT_FORCE
	}
	if umount.flags.ByName["--fake"] {
		fmt.Println("Would umount", target)
		return nil
	}
	if err := syscall.Unmount(target, flags); err != nil {
		return err
	}
	if umount.flags.ByName["-v"] {
		fmt.Println("Unmounted", target)
	}
	return nil
}
