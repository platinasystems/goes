// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xeth

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

type NetNs uint64
type NetNses []NetNs

type NetNsAdd struct {
	NetNs
}

type NetNsDel struct {
	NetNs
}

const DefaultNetNs NetNs = 1

type netnsAttrs struct {
	path     string // or pid
	xids     sync.Map
	neigbors sync.Map
	localRT  sync.Map
	mainRT   sync.Map
	otherRTs sync.Map
}

var netnsAttrsMap sync.Map

func NetNsRange(f func(ns NetNs) bool) {
	netnsAttrsMap.Range(func(k, v interface{}) bool {
		return f(k.(NetNs))
	})
}

func NewNetNses() (l NetNses) {
	netnsAttrsMap.Range(func(k, v interface{}) bool {
		l = append(l, k.(NetNs))
		return true
	})
	return
}

func (l NetNses) Cut(i int) NetNses {
	copy(l[i:], l[i+1:])
	return l[:len(l)-1]
}

func (l NetNses) FilterName(re *regexp.Regexp) NetNses {
	for i := 0; i < len(l); {
		ns := l[i]
		if re.MatchString(ns.String()) {
			i += 1
		} else {
			l = l.Cut(i)
		}
	}
	return l
}

func (l NetNses) InodeFilter(inodes []uint64) NetNses {
	for i := 0; i < len(l); {
		inode := l[i].Inode()
		match := false
		for _, c := range inodes {
			if inode == c {
				match = true
			}
		}
		if match {
			i += 1
		} else {
			l = l.Cut(i)
		}
	}
	return l
}

func (l NetNses) Inodes() []uint64 {
	inodes := make([]uint64, len(l))
	for i, ns := range l {
		inodes[i] = ns.Inode()
	}
	return inodes
}

func (ns NetNs) Base() string {
	return filepath.Base(ns.Path())
}

func (ns NetNs) FibEntry(rt RtTable, ipnet string) (fe *FibEntry) {
	if v, ok := ns.rtm(rt).Load(ipnet); ok {
		fe = v.(*FibEntry)
	}
	return
}

func (ns NetNs) FibEntries(rt RtTable, f func(fe *FibEntry) bool) {
	var rtm *sync.Map
	attrs := ns.attrs()
	switch rt {
	case MainRtTable:
		rtm = &attrs.mainRT
	case LocalRtTable:
		rtm = &attrs.localRT
	default:
		if v, ok := attrs.otherRTs.Load(rt); ok {
			rtm = v.(*sync.Map)
		} else {
			return
		}
	}
	rtm.Range(func(k, v interface{}) bool {
		return f(v.(*FibEntry))
	})
}

func (ns NetNs) Inode() uint64 {
	return uint64(ns)
}

func (ns NetNs) Neighbor(ip string) (neigh *Neighbor) {
	if v, ok := ns.attrs().neigbors.Load(ip); ok {
		neigh = v.(*Neighbor)
	}
	return
}

func (ns NetNs) Neighbors(f func(neigh *Neighbor) bool) {
	ns.attrs().neigbors.Range(func(k, v interface{}) bool {
		return f(v.(*Neighbor))
	})
}

func (ns NetNs) Path() string {
	attrs := ns.attrs()
	if len(attrs.path) > 0 {
		if attrs.path == "default" {
			return attrs.path
		}
		// check if path is still valid
		if attrs.path != "unknown" {
			if strings.Contains(attrs.path, "/run/netns") {
				if info, err := os.Stat(attrs.path); err == nil && info.Sys() != nil {
					stat := info.Sys().(*syscall.Stat_t)
					if stat.Ino == uint64(ns) {
						return attrs.path
					}
				}
			} else {
				path := filepath.Join("/proc/", attrs.path, "ns/net")
				if _, err := os.Stat(path); err == nil {
					if ln, err := os.Readlink(path); err == nil {
						if strings.Contains(ln, strconv.Itoa(int(ns))) {
							return attrs.path
						}
					}
				}
			}
		}
		// path no longer valid; reset
		attrs.path = ""
	}
	if ns == DefaultNetNs {
		attrs.path = "default"
		return attrs.path
	}
	filepath.Walk("/run/netns",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if len(attrs.path) > 0 {
				return filepath.SkipDir
			}
			stat := info.Sys().(*syscall.Stat_t)
			if stat.Ino == uint64(ns) {
				attrs.path = path
				return filepath.SkipDir
			}
			return nil
		})
	if len(attrs.path) > 0 {
		return attrs.path
	}
	attrs.path = "unknown"
	globs, err := filepath.Glob("/proc/[0-9]*/ns")
	if err != nil {
		return attrs.path
	}
	for _, procns := range globs {
		pid := strings.Split(procns, "/")[2]
		procnsnet := filepath.Join(procns, "net")
		ln, err := os.Readlink(procnsnet)
		if err != nil {
			continue
		}
		if !strings.HasPrefix(ln, "net:[") {
			continue
		}
		is := strings.TrimSuffix(strings.TrimPrefix(ln, "net:["), "]")
		var inode uint64
		_, err = fmt.Sscan(is, &inode)
		if err == nil && inode == uint64(ns) {
			attrs.path = pid
			return pid
		}
	}
	return attrs.path
}

func (ns NetNs) RtTables() []RtTable {
	rts := []RtTable{MainRtTable, LocalRtTable}
	ns.attrs().otherRTs.Range(func(k, v interface{}) bool {
		rts = append(rts, k.(RtTable))
		return true
	})
	return rts
}

func (ns NetNs) String() string {
	return ns.Base()
}

// if set[0] == 0, delete ifindex entry
// if set[0] != 0, map by ifindex
// otherwise, return Xid mapped by ifindex
func (ns NetNs) Xid(ifindex int32, set ...Xid) (xid Xid) {
	attrs := ns.attrs()
	if len(set) > 0 {
		xid = set[0]
		if xid == 0 {
			attrs.xids.Delete(ifindex)
		} else {
			attrs.xids.Store(ifindex, xid)
		}
	} else if v, ok := attrs.xids.Load(ifindex); ok {
		xid = v.(Xid)
	}
	return
}

func (ns NetNs) Xids(f func(xid Xid) bool) {
	ns.attrs().xids.Range(func(k, v interface{}) bool {
		return f(v.(Xid))
	})
}

func (ns NetNs) attrs() (attrs *netnsAttrs) {
	if v, ok := netnsAttrsMap.Load(ns); ok {
		attrs = v.(*netnsAttrs)
	} else {
		attrs = new(netnsAttrs)
		netnsAttrsMap.Store(ns, attrs)
	}
	return
}

func (ns NetNs) fibentry(fe *FibEntry) {
	rtm := ns.rtm(fe.RtTable)
	sipnet := fe.IPNet.String()
	switch fe.FibEntryEvent {
	case FIB_EVENT_ENTRY_DEL:
		if v, ok := rtm.Load(sipnet); ok {
			defer v.(*FibEntry).Pool()
			rtm.Delete(sipnet)
		}
	case FIB_EVENT_ENTRY_ADD, FIB_EVENT_ENTRY_REPLACE:
		if v, ok := rtm.Load(sipnet); ok {
			defer v.(*FibEntry).Pool()
		}
		fe.Hold()
		rtm.Store(sipnet, fe)
	case FIB_EVENT_ENTRY_APPEND:
	case FIB_EVENT_RULE_ADD:
	case FIB_EVENT_RULE_DEL:
	case FIB_EVENT_NH_ADD:
	case FIB_EVENT_NH_DEL:
	}
}

func (ns NetNs) neighbor(neigh *Neighbor) {
	attrs := ns.attrs()
	sip := neigh.IP.String()
	for _, b := range neigh.HardwareAddr {
		if b == 0 {
			if v, ok := attrs.neigbors.Load(sip); ok {
				attrs.neigbors.Delete(sip)
				v.(*Neighbor).Pool()
			}
			return
		}
	}
	neigh.Hold()
	if v, ok := attrs.neigbors.Load(sip); ok {
		v.(*Neighbor).Pool()
	}
	attrs.neigbors.Store(sip, neigh)
}

func (ns NetNs) rtm(rt RtTable) (rtm *sync.Map) {
	attrs := ns.attrs()
	switch rt {
	case MainRtTable:
		rtm = &attrs.mainRT
	case LocalRtTable:
		rtm = &attrs.localRT
	default:
		if v, ok := attrs.otherRTs.Load(rt); ok {
			rtm = v.(*sync.Map)
		} else {
			rtm = new(sync.Map)
			attrs.otherRTs.Store(rt, rtm)
		}
	}
	return
}
