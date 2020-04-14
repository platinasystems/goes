// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package search

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/external/parms"
	"github.com/platinasystems/goes/internal/partitions"
	"github.com/platinasystems/goes/lang"
)

type Command struct {
	g *goes.Goes
}

func (c Command) String() string { return "search" }

func (c Command) Usage() string {
	return "NOP"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "NOP",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: Man,
	}
}

const Man = "Search for a filesystem"

func (Command) Kind() cmd.Kind { return cmd.DontFork }

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (c Command) Main(args ...string) error {
	parm, args := parms.New(args, "--file", "--label", "--hint", "--uuid", "--set")
	_, args = flags.New(args, "--no-floppy", "--fs-uuid")

	v := parm.ByName["--set"]

	if v == "" {
		v = "root"
	}
	if len(args) != 1 {
		return fmt.Errorf("Unexpected %v\n", args)
	}
	f, err := os.Open("/proc/partitions")
	if err != nil {
		fmt.Printf("opening /proc/partitions: %s\n", err)
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 4 || fields[0] == "major" {
			continue
		}
		fileName := fields[3]
		sb, err := partitions.ReadSuperBlock("/dev/" + fileName)
		if err == nil {
			u, _ := sb.UUID()
			if u.String() == args[0] {
				if c.g.EnvMap == nil {
					c.g.EnvMap = make(map[string]string)
				}
				c.g.EnvMap[v] = "/dev/" + fileName
				return nil
			}
		}
	}
	return nil
}
