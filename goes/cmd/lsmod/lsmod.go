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

type Command struct{}

func (Command) String() string { return "lsmod" }

func (Command) Usage() string { return "lsmod" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "print status of Linux Kernel modules",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Formatted print of /proc/modules.`,
	}
}

func (Command) Main(_ ...string) error {
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
