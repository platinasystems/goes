// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package lsmod

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const Name = "lsmod"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name }

func (cmd) Main(_ ...string) error {
	f, err := os.Open("/proc/modules")
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	fmt.Printf("%-19s %s\n",
		"Module", "    Size  Used by")
	for scanner.Scan() {
		line := scanner.Text()
		x := strings.Fields(line)
		if len(x) < 4 {
			x = append(x, "")
		}
		if x[3] == "-" {
			x[3] = ""
		}
		if strings.HasSuffix(x[3], ",") {
			x[3] = x[3][:len(x[3])-1]
		}
		fmt.Printf("%-19s %8s %2s %s\n",
			x[0], x[1], x[2], x[3])
	}
	if err = scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print status of Linux Kernel modules",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	lsmod - print Linux Kernel module status

SYNOPSIS
	lsmod

DESCRIPTION
	Formatted print of /proc/modules.`,
	}
}
