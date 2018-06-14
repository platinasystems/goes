// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/internal/netlink"
	"github.com/platinasystems/go/vnet"

	"regexp"
	"strings"
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
	case netlink.InterfaceKindDummy, netlink.InterfaceKindTun, netlink.InterfaceKindVeth, netlink.InterfaceKindVlan:
		ok = true
	}
	return
}

type Main struct {
	vnet.Package
	v               *vnet.Vnet
	verbose_netlink bool
	interface_filter
	net_namespace_main
	netlink_main
	Config
}

func GetMain(v *vnet.Vnet) *Main { return v.GetPackage(packageIndex).(*Main) }

type Config struct {
	RxInjectNodeName string
}

func Init(v *vnet.Vnet, cf Config) {
	m := &Main{}
	m.v = v
	m.Config = cf
	m.netlink_main.Init(m)
	packageIndex = v.AddPackage("unix", m)
}

func (m *Main) Configure(in *parse.Input) {
	for !in.End() {
		var s string
		switch {
		case in.Parse("verbose-netlink"):
			m.verbose_netlink = true
		case in.Parse("filter-accept %s", &s):
			m.interface_filter.add(s, true)
		case in.Parse("filter-reject %s", &s):
			m.interface_filter.add(s, false)
		default:
			in.ParseError()
		}
	}
}

type ifreq_name [16]byte

func (n ifreq_name) String() string { return strings.TrimRight(string(n[:]), "\x00") }

// Linux interface flags
type iff_flag int

const (
	iff_up_bit, iff_up iff_flag = iota, 1 << iota
	iff_broadcast_bit, iff_broadcast
	iff_debug_bit, iff_debug
	iff_loopback_bit, iff_loopback
	iff_pointopoint_bit, iff_pointopoint
	iff_notrailers_bit, iff_notrailers
	iff_running_bit, iff_running
	iff_noarp_bit, iff_noarp
	iff_promisc_bit, iff_promisc
	iff_allmulti_bit, iff_allmulti
	iff_master_bit, iff_master
	iff_slave_bit, iff_slave
	iff_multicast_bit, iff_multicast
	iff_portsel_bit, iff_portsel
	iff_automedia_bit, iff_automedia
	iff_dynamic_bit, iff_dynamic
	iff_lower_up_bit, iff_lower_up
	iff_dormant_bit, iff_dormant
	iff_echo_bit, iff_echo
)

var iff_flag_names = [...]string{
	iff_up_bit:          "admin-up",
	iff_broadcast_bit:   "broadcast",
	iff_debug_bit:       "debug",
	iff_loopback_bit:    "loopback",
	iff_pointopoint_bit: "point-to-point",
	iff_notrailers_bit:  "no-trailers",
	iff_running_bit:     "running",
	iff_noarp_bit:       "no-arp",
	iff_promisc_bit:     "promiscuous",
	iff_allmulti_bit:    "all-multicast",
	iff_master_bit:      "master",
	iff_slave_bit:       "slave",
	iff_multicast_bit:   "multicast",
	iff_portsel_bit:     "portsel",
	iff_automedia_bit:   "automedia",
	iff_dynamic_bit:     "dynamic",
	iff_lower_up_bit:    "link-up",
	iff_dormant_bit:     "dormant",
	iff_echo_bit:        "echo",
}

func (f iff_flag) String() string { return elib.FlagStringer(iff_flag_names[:], elib.Word(f)) }

func (m *Main) netlink_discovery_done_for_all_namespaces() (err error) {
	// Perform all defered registrations for unix interfaces.
	for _, hw := range m.registered_hwifer_by_si {
		m.RegisterHwInterface(hw)
	}

	return
}

func (m *Main) Init() (err error) {
	//Suitable defaults for an Ethernet-like tun/tap device.
	return
}
