// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mknod

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

const Name = "mknod"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }

func (cmd) Usage() string {
	return Name + " [OPTION]... NAME TYPE [MAJOR MINOR]"
}

func (cmd) Main(args ...string) error {
	var filetype uint32 = 0
	if len(args) == 0 {
		return fmt.Errorf("FILE: missing")
	}
	aa := 0
	for _, ar := range args {
		if strings.Contains(ar, "-m=") {
			aa++
		}
	}
	l := len(args) - aa
	a0 := 0 + aa
	a1 := 1 + aa
	a2 := 2 + aa
	a3 := 3 + aa
	switch args[a1] {
	case "b":
		filetype = syscall.S_IFBLK
	case "c", "u":
		filetype = syscall.S_IFCHR
	case "p":
		filetype = syscall.S_IFIFO
	case "d":
		filetype = syscall.S_IFDIR
	case "r":
		filetype = syscall.S_IFREG
	}
	filetype |= uint32(os.FileMode(flagValue(args, "m")))
	nmaj := 0
	nmin := 0
	var err error
	if l > 2 {
		nmaj, err = strconv.Atoi(args[a2])
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	}
	if l > 3 {
		nmin, err = strconv.Atoi(args[a3])
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	}
	n := (nmaj * 256) + nmin
	err = syscall.Mknod(args[a0], filetype, n)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "make block or character special files",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
        mknod - make block or character special files

SYNOPSIS
        mknod [OPTION]... NAME TYPE [MAJOR MINOR]

OPTIONS
        -m`,
	}
}

func flagValue(a []string, f string) uint32 {
	for _, arg := range a {
		if strings.Contains(arg, "-"+f+"=") {
			result := strings.SplitAfter(arg, "=")
			if len(result) > 1 {
				i, err := strconv.ParseInt("0"+strings.TrimSpace(result[1]), 8, 32)
				if err != nil {
					return 0
				}
				return uint32(i)
			}
		}
	}
	return 0
}
