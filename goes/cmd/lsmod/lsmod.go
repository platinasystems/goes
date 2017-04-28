// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package lsmod

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "lsmod"
	Apropos = "print status of Linux Kernel modules"
	Usage   = "lsmod"
	Man     = `
DESCRIPTION
	Formatted print of /proc/modules.`
)

type Interface interface {
	Apropos() lang.Alt
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }

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

func (cmd) Man() lang.Alt  { return man }
func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)
