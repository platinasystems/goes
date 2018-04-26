// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package testcmd

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"path/filepath"

	"github.com/mattn/go-isatty"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
)

const (
	R_OK = 4
	W_OK = 2
	X_OK = 1
)

var errUnexpected = fmt.Errorf("unexpected")

type Command struct{}

func (Command) String() string { return "[" }

func (Command) Usage() string { return "[ COND ]" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "test conditions and set exit status",
	}
}

func (Command) Kind() cmd.Kind { return cmd.NoCLIFlags }

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Tests conditions and returns zero or non-zero exit status
`,
	}
}

func isGroupMember(gid uint32) bool {
	if gid == uint32(syscall.Getgid()) || gid == uint32(syscall.Getegid()) {
		return true
	}

	groups, err := syscall.Getgroups()
	if err != nil {
		return false
	}

	for _, group := range groups {
		if gid == uint32(group) {
			return true
		}
	}
	return false
}

func statWithLinks(file string) (syscall.Stat_t, error) {
	var stats syscall.Stat_t

	file, err := filepath.EvalSymlinks(file)
	if err != nil {
		return stats, err
	}

	err = syscall.Stat(file, &stats)
	return stats, err
}

func (c Command) parseBinaryOpt(args []string) (bool, error) {
	if args[1] == "=" {
		return args[0] == args[2], nil
	}
	if args[1] == "!=" {
		return args[0] != args[2], nil
	}
	if args[1] == "-ef" {
		stats1, err := statWithLinks(args[0])
		if err != nil {
			return false, err
		}
		stats2, err := statWithLinks(args[2])
		if err != nil {
			return false, err
		}
		return stats1.Dev == stats2.Dev &&
			stats1.Ino == stats2.Ino, nil
	}
	if args[1] == "-nt" {
		stats1, err := statWithLinks(args[0])
		if err != nil {
			return false, err
		}
		stats2, err := statWithLinks(args[2])
		if err != nil {
			return false, err
		}
		return stats1.Mtim.Nano() > stats2.Mtim.Nano(), nil
	}
	if args[1] == "-ot" {
		stats1, err := statWithLinks(args[0])
		if err != nil {
			return false, err
		}
		stats2, err := statWithLinks(args[2])
		if err != nil {
			return false, err
		}
		return stats1.Mtim.Nano() <= stats2.Mtim.Nano(), nil
	}

	for _, o := range []struct {
		opt string
	}{
		{"-eq"}, {"-ge"}, {"-gt"}, {"-le"}, {"-lt"}, {"-ne"},
	} {
		if args[1] == o.opt {
			int1, err := strconv.Atoi(args[0])
			if err != nil {
				return false, err
			}
			int2, err := strconv.Atoi(args[2])
			if err != nil {
				return false, err
			}
			switch args[1] {
			case "-eq":
				return int1 == int2, nil
			case "-ge:":
				return int1 >= int2, nil
			case "-gt":
				return int1 > int2, nil
			case "-le":
				return int1 <= int2, nil
			case "-lt":
				return int1 < int2, nil
			case "-ne":
				return int1 != int2, nil
			}
		}
	}
	return false, errUnexpected
}

func (c Command) parseUnaryOp(args []string) (bool, error) {
	if len(args) < 2 {
		return false, errUnexpected
	}
	for _, o := range []struct {
		opt  string
		mask uint32
		val  uint32
	}{
		{"-h", syscall.S_IFMT, syscall.S_IFLNK},
		{"-L", syscall.S_IFMT, syscall.S_IFLNK},
	} {
		if args[0] == o.opt {
			var stats syscall.Stat_t
			err := syscall.Stat(args[1], &stats)
			if err != nil {
				return false, err
			}
			if stats.Mode&o.mask == o.val {
				return true, nil
			}
			return false, nil
		}
	}

	for _, o := range []struct {
		opt  string
		mask uint32
		val  uint32
	}{
		{"-b", syscall.S_IFMT, syscall.S_IFBLK},
		{"-c", syscall.S_IFMT, syscall.S_IFCHR},
		{"-d", syscall.S_IFMT, syscall.S_IFDIR},
		{"-f", syscall.S_IFMT, syscall.S_IFREG},
		{"-g", syscall.S_ISGID | syscall.S_IXGRP, syscall.S_ISGID | syscall.S_IXGRP},
		{"-k", syscall.S_ISVTX, syscall.S_ISVTX},
		{"-p", syscall.S_IFMT, syscall.S_IFIFO},
		{"-S", syscall.S_IFMT, syscall.S_IFSOCK},
		{"-u", syscall.S_ISUID, syscall.S_ISUID},
	} {
		if args[0] == o.opt {
			stats, err := statWithLinks(args[1])
			if err != nil {
				return false, err
			}
			if stats.Mode&o.mask == o.val {
				return true, nil
			}
			return false, nil
		}
	}
	for _, o := range []struct {
		opt  string
		mode uint32
	}{
		{"-r", R_OK},
		{"-w", W_OK},
		{"-x", X_OK},
	} {
		if args[0] == o.opt {
			stats, err := statWithLinks(args[1])
			if err != nil {
				return false, err
			}
			euid := syscall.Geteuid()
			if euid == 0 {
				if o.mode != X_OK {
					return true, nil
				}
				if (stats.Mode & (syscall.S_IXUSR | syscall.S_IXGRP | syscall.S_IXOTH)) != 0 {
					return true, nil
				}
			}
			if stats.Uid == uint32(euid) {
				o.mode <<= 6
			} else {
				if isGroupMember(stats.Gid) {
					o.mode <<= 3
				}
			}
			if stats.Mode&o.mode != 0 {
				return true, nil
			}
			return false, nil
		}
	}
	if args[0] == "-e" {
		_, err := statWithLinks(args[1])
		if err != nil {
			return false, err
		}
		return true, nil
	}
	if args[0] == "-G" {
		stats, err := statWithLinks(args[1])
		if err != nil {
			return false, err
		}
		return stats.Gid == uint32(syscall.Getegid()), nil
	}
	if args[0] == "-n" {
		return len(args[1]) != 0, nil
	}
	if args[0] == "-O" {
		stats, err := statWithLinks(args[1])
		if err != nil {
			return false, err
		}
		return stats.Uid == uint32(syscall.Geteuid()), nil
	}
	if args[0] == "-s" {
		stats, err := statWithLinks(args[1])
		if err != nil {
			return false, err
		}
		return stats.Blocks > 0, nil
	}
	if args[0] == "-t" {
		fd, err := strconv.Atoi(args[1])
		if err != nil {
			return false, err
		}
		return isatty.IsTerminal(uintptr(fd)), nil
	}
	if args[0] == "-z" {
		return len(args[1]) == 0, nil
	}
	return false, errUnexpected
}

func (c Command) parse(args []string) ([]string, bool, error) {
	if len(args) >= 4 {
		if args[2] == "-l" {
			str := strconv.Itoa(len(args[3]))
			pre := append(args[:2], str)
			args = append(pre, args[4:]...)
			return c.parse(args)
		}
		if args[0] == "-l" {
			str := []string{strconv.Itoa(len(args[1]))}
			args = append(str, args[2:]...)
			return c.parse(args)
		}
		if args[0] == "(" && args[len(args)-1] == ")" {
			return c.parse(args[1 : len(args)-1])
		}
		res, val, err := c.parse(args[:3])
		if err != nil {
			return args, val, err
		}
		args = append(res, args[3:]...)
		if args[0] == "-o" {
			if val {
				return nil, true, nil
			}
			return c.parse(args[1:])
		}
		if args[0] == "-a" {
			if !val {
				return nil, false, nil
			}
			return c.parse(args[1:])
		}
		return args, false, nil
	}
	if len(args) >= 3 {
		val, err := c.parseBinaryOpt(args)
		if err == nil {
			return args[3:], val, nil
		} else {
			if err != errUnexpected {
				return args, false, err
			}
		}
	}
	if len(args) >= 2 {
		val, err := c.parseUnaryOp(args)
		if err == nil {
			return args[2:], val, nil
		} else {
			if err != errUnexpected {
				return args, false, err
			}
		}
		if args[0] == "!" {
			args, val, err = c.parse(args[1:])
			return args, !val, err
		}
	}
	if len(args[0]) != 0 {
		return args[1:], true, nil
	}
	return args[1:], false, nil
}

func (c Command) Main(args ...string) error {
	if len(args) < 1 || args[len(args)-1] != "]" {
		return fmt.Errorf("missing ]")
	}
	args = args[0 : len(args)-1]

	args, val, err := c.parse(args)

	if err != nil {
		os.Exit(1)
		return err
	}

	if len(args) == 0 {
		if val {
			return nil
		}
		os.Exit(1)
		return fmt.Errorf("false")
	}

	return fmt.Errorf("Unexpected %v", args)
}
