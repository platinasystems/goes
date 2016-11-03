// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package nsid

import (
	"io/ioutil"
	"strings"
)

func (*nsid) Complete(args ...string) (c []string) {
	var common = []string{
		"apropos",
		"-apropos",
		"--apropos",
		"help",
		"-help",
		"--help",
		"man",
		"-man",
		"--man",
		"usage",
		"-usage",
		"--usage",
	}
	var cmds = []string{
		"list",
		"set",
		"unset",
	}
	if len(args) > 0 && strings.HasSuffix(args[0], "nsid") {
		args = args[1:]
	}
	switch len(args) {
	case 0:
		c = cmds
	case 1:
		for _, slice := range [][]string{common, cmds} {
			for _, cmd := range slice {
				if strings.HasPrefix(cmd, args[0]) {
					c = append(c, cmd)
				}
			}
		}
	case 2:
		if args[0] != "set" && args[0] != "unset" {
			return
		}
		dir, err := ioutil.ReadDir(VarRunNetns)
		if err != nil {
			return
		}
		for _, info := range dir {
			name := info.Name()
			if strings.HasPrefix(name, args[1]) {
				c = append(c, name)
			}
		}
	}
	return
}
