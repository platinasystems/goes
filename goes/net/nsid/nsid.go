// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build linux

// Package `nsid` provides a List, Set and Unset of network namespace
// identifiers.
package nsid

import "fmt"

const VarRunNetns = "/var/run/netns"

type Entry struct {
	Name string
	Id   int32
	Pid  uint32
}

type nsid struct {
	seq uint32
}

func New() *nsid { return &nsid{} }

func (nsid *nsid) String() string { return "nsid" }

func (nsid *nsid) Usage() string {
	return `nsid [list]
	nsid set NAME ID
	nsid unset NAME ID`
}

func (nsid *nsid) Main(args ...string) error {
	cmd := "list"
	if len(args) > 0 {
		cmd = args[0]
	}
	setf := nsid.Unset
	switch cmd {
	case "list":
		list, err := nsid.List()
		if err != nil {
			return err
		}
		for _, entry := range list {
			fmt.Print(entry.Name, ": ", entry.Id, "\n")
		}
	case "set":
		setf = nsid.Set
		fallthrough
	case "unset":
		if len(args) < 2 {
			return fmt.Errorf("NAME: missing")
		}
		name := args[1]
		if len(args) < 3 {
			return fmt.Errorf("ID: missing")
		}
		var id int32
		if _, err := fmt.Sscan(args[2], &id); err != nil {
			return fmt.Errorf("%s: %v", args[2], err)
		}
		return setf(name, id)
	case "apropos", "-apropos", "--apropos":
		fmt.Println(nsid.Apropos())
	case "complete", "-complete", "--complete":
		for _, s := range nsid.Complete(args[1:]...) {
			fmt.Println(s)
		}
	case "help", "-h", "-help", "--help":
		if len(args) > 1 {
			fmt.Println(nsid.Help(args[1:]...))
			break
		}
		fallthrough
	case "usage", "-usage", "--usage":
		fmt.Print("usage:\t", nsid.Usage(), "\n")
	case "man", "-man", "--man":
		fmt.Println(nsid.Man())
	default:
		return fmt.Errorf("%s: command not found\nusage:\t%s", cmd,
			nsid.Usage())
	}
	return nil
}

func (*nsid) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "net namespace identifier config",
	}
}

func (*nsid) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	nsid - net namespace identifier config

SYNOPSIS
	nsid [list]
	nsid set NAME ID
	nsid unet NAME ID

DESCRIPTION
	[list]	show the identifier of each network namespace with "-1"
		indicating an unidentifeid namespace.

	set	set the namespace identifier

	unset	unset the namespace identifier`,
	}
}
