// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Cache parse of /etc/iproute2/group
package group

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

const FileName = "/etc/iproute2/group"

var (
	byid   map[uint32]string
	byname map[string]uint32
	mutex  sync.Mutex
)

func Id(name string) uint32 {
	if byname == nil {
		parse()
	}
	id, found := byname[name]
	if !found {
		id = ^uint32(0)
		fmt.Sscan(name, &id)
	}
	return id
}

func Name(id uint32) string {
	if byid == nil {
		parse()
	}
	name, found := byid[id]
	if !found {
		name = fmt.Sprint(id)
	}
	return name
}

func parse() {
	mutex.Lock()
	defer mutex.Unlock()
	if byid != nil {
		return
	}
	byid = make(map[uint32]string)
	byname = make(map[string]uint32)
	f, err := os.Open(FileName)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var id uint32
		var name string
		n, err := fmt.Sscan(scanner.Text(), &id, &name)
		if err == nil && n == 2 {
			byid[id] = name
			byname[name] = id
		}
	}
}
