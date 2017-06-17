// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/internal/netlink"
	"github.com/platinasystems/go/vnet"

	"regexp"
)

var packageIndex uint

type interface_filter struct {
	// Map of source string regexp to indication of whether or not matching interfaces should be terminated.
	s map[string]bool
	// As above but after compilation of regexps.
	m map[*regexp.Regexp]bool
}

func (f *interface_filter) add(s string, v bool) {
	if f.s == nil {
		f.s = make(map[string]bool)
	}
	f.s[s] = v
}
func AddInterfaceFilter(v *vnet.Vnet, s string, ok bool) { GetMain(v).interface_filter.add(s, ok) }

func (f *interface_filter) compile() (err error) {
	f.m = make(map[*regexp.Regexp]bool, len(f.s))
	for s, v := range f.s {
		var e *regexp.Regexp
		if e, err = regexp.Compile(s); err != nil {
			return
		}
		f.m[e] = v
	}
	return
}

func (f *interface_filter) run(s string, kind netlink.InterfaceKind) (ok bool) {
	if len(f.m) != len(f.s) {
		err := f.compile()
		if err != nil {
			panic(err)
		}
	}
	for e, v := range f.m {
		if e.MatchString(s) {
			ok = v
			return
		}
	}
	switch kind {
	case netlink.InterfaceKindDummy, netlink.InterfaceKindTun, netlink.InterfaceKindVeth:
		ok = true
	}
	return
}

type Main struct {
	vnet.Package

	v *vnet.Vnet

	verbose_packets bool
	verbose_netlink int
	interface_filter

	netlink_main
	tuntap_main
}

func GetMain(v *vnet.Vnet) *Main { return v.GetPackage(packageIndex).(*Main) }

func Init(v *vnet.Vnet) {
	m := &Main{}
	m.v = v
	m.tuntap_main.Init(v)
	m.netlink_main.Init(m)
	packageIndex = v.AddPackage("unix", m)
}

func (m *Main) Configure(in *parse.Input) {
	for !in.End() {
		var s string
		switch {
		case in.Parse("mtu %d", &m.mtuBytes):
		case in.Parse("tap"):
			m.isTun = false
		case in.Parse("tun"):
			m.isTun = true
		case in.Parse("verbose-packets"):
			m.verbose_packets = true
		case in.Parse("verbose-netlink"):
			m.verbose_netlink++
		case in.Parse("filter-accept %s", &s):
			m.interface_filter.add(s, true)
		case in.Parse("filter-reject %s", &s):
			m.interface_filter.add(s, false)
		default:
			panic(parse.ErrInput)
		}
	}
}
