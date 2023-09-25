// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build linux

// Package group provides an /etc/group parser.
package group

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Entry struct {
	pw      string
	gid     int
	members []string
}

func Parse() map[string]*Entry {
	f, err := os.Open("/etc/group")
	if err != nil {
		return map[string]*Entry{}
	}
	defer f.Close()
	m := make(map[string]*Entry)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) < 3 {
			continue
		}
		entry := new(Entry)
		entry.pw = fields[1]
		gid, err := strconv.ParseInt(fields[2], 0, 0)
		if err != nil {
			continue
		}
		entry.gid = int(gid)
		if len(fields) == 4 {
			entry.members = strings.Split(fields[3], ",")
		}
		m[fields[0]] = entry
	}
	return m
}

func (p *Entry) Gid() (gid int) {
	if p != nil {
		gid = p.gid
	}
	return
}

func (p *Entry) Members() (members []string) {
	if p != nil {
		members = p.members
	}
	return
}

func (p *Entry) Passwd() (pw string) {
	if p != nil {
		pw = p.pw
	}
	return
}
