// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package `nsid` provides a List, Set and Unset of network namespace
// identifiers.
package nsid

import "fmt"

const (
	Usage = `
	nsid [list [NAME]...]
	nsid set NAME ID
	nsid unset NAME ID`

	VarRunNetns = "/var/run/netns"
)

func Main(args ...string) error {
	cmd := "list"
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}
	setf := Unset
	switch cmd {
	case "-h", "-help", "--help":
		fmt.Print("usage:", Usage[1:], "\n")
		return nil
	case "list":
		entries, err := List(args...)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			fmt.Print(entry.Name, ": ", entry.Nsid, "\n")
		}
	case "set":
		setf = Set
		fallthrough
	case "unset":
		if len(args) < 1 {
			return fmt.Errorf("NAME: missing")
		}
		name := args[0]
		if len(args) < 2 {
			return fmt.Errorf("ID: missing")
		}
		var id int32
		if _, err := fmt.Sscan(args[1], &id); err != nil {
			return fmt.Errorf("%s: %v", args[1], err)
		}
		return setf(name, id)
	default:
		return fmt.Errorf("%s: command not found\nusage:%s", cmd,
			Usage[1:])
	}
	return nil
}
