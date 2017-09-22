// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux

package unix

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/iomux"
	"github.com/platinasystems/go/internal/netlink"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip4"

	"fmt"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

type tuntap_interface struct {
	m  *Main
	mu sync.Mutex
	// Namespace this interface is currently in.
	namespace *net_namespace
	// Raw socket bound to this interface used for provisioning.
	provision_fd   int
	dev_net_tun_fd int
	iomux.File     // /dev/net/tun fd for this interface.
	hi             vnet.Hi
	si             vnet.Si
	name           ifreq_name
	elog_name      elog.StringRef
	ifindex        uint32 // linux interface index

	// Tun (ip4/ip6 header) versus tap (has ethernet header).
	isTun bool

	// Tuntap interface has been created (via TUNSETIFF ioctl).
	created bool
	// True when vnet/kernel interface flag sync has started.
	flag_sync_in_progress bool
	// True when vnet/kernel interface flags have been successfully synchronized.
	flag_sync_done bool
	flags          iff_flag
	operState      netlink.IfOperState

	mtuBytes   uint
	mtuBuffers uint

	active_refs int32
	to_tx       chan *tx_packet_vector
	pv          *tx_packet_vector

	interface_routes ip4.MapFib
}

//go:generate gentemplate -d Package=unix -id ifVec -d VecType=interfaceVec -d Type=*tuntap_interface github.com/platinasystems/go/elib/vec.tmpl

func (i *tuntap_interface) Name() string   { return i.name.String() }
func (i *tuntap_interface) String() string { return i.Name() }

func (i *tuntap_interface) set_name(name string) {
	v := i.m.v
	v.HwIf(i.hi).SetName(v, name)
}

func (i *tuntap_interface) setMtu(m *Main, mtu uint) {
	i.mtuBytes = mtu
	i.mtuBuffers = mtu / m.bufferPool.Size
	if mtu%m.bufferPool.Size != 0 {
		i.mtuBuffers++
	}
}

func (i *tuntap_interface) close(provision bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if provision && i.provision_fd != -1 {
		syscall.Close(i.provision_fd)
		i.provision_fd = -1
	}
	if i.dev_net_tun_fd != -1 {
		if i.Fd != -1 {
			iomux.Del(i)
		}
		i.Fd = -1
		syscall.Close(i.dev_net_tun_fd)
		i.dev_net_tun_fd = -1
	}
}

// Called when interface namespace is added/deleted or when interface moves namespace.
func (i *tuntap_interface) add_del_namespace(m *Main, ns *net_namespace, is_del bool) (err error) {
	if is_del {
		i.close(true)
	} else {
		is_discovery := !i.created && i.namespace == nil
		if is_discovery {
			return
		}
		i.namespace = ns
		// Close sockets before re-opening in new namespace.
		i.close(true)
		if err = i.open_sockets(); err != nil {
			return
		}
		if err = i.create(); err != nil {
			return
		}
		if err = i.bind(); err != nil {
			return
		}
		i.start_up()
	}
	return
}

type tuntap_main struct {
	// Selects whether we create tun or tap interfaces.
	mtuBytes   uint
	bufferPool *vnet.BufferPool
}

func (m *tuntap_main) Init(v *vnet.Vnet) {
	m.bufferPool = vnet.DefaultBufferPool
	v.AddBufferPool(m.bufferPool)
}

const (
	// TUNSETIFF ifReq flags
	iff_tun          = 1 << 0
	iff_tap          = 1 << 1
	iff_multi_queue  = 1 << 8
	iff_attach_queue = 1 << 9
	iff_detach_queue = 1 << 10
	iff_persist      = 1 << 11
	iff_no_pi        = 1 << 12
	iff_one_queue    = 1 << 13
	iff_vnet_hdr     = 1 << 14
	iff_tun_excl     = 1 << 15
	iff_nofilter     = 1 << 12
)

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

type ifreq_flags struct {
	name  ifreq_name
	flags uint16
}

type ifreq_int struct {
	name ifreq_name
	i    int
}

type ifreq_sockaddr_any struct {
	name     ifreq_name
	sockaddr syscall.RawSockaddrAny
}

type ifreq_type int

const (
	ifreq_TUNSETIFF     ifreq_type = syscall.TUNSETIFF
	ifreq_TUNSETPERSIST ifreq_type = syscall.TUNSETPERSIST
	ifreq_GETIFINDEX    ifreq_type = syscall.SIOCGIFINDEX
	ifreq_GETIFFLAGS    ifreq_type = syscall.SIOCGIFFLAGS
	ifreq_SETIFFLAGS    ifreq_type = syscall.SIOCSIFFLAGS
	ifreq_GETIFHWADDR   ifreq_type = syscall.SIOCGIFHWADDR
	ifreq_SETIFHWADDR   ifreq_type = syscall.SIOCSIFHWADDR
	ifreq_SETIFMTU      ifreq_type = syscall.SIOCSIFMTU
	ifreq_SIFTXQLEN     ifreq_type = syscall.SIOCSIFTXQLEN
)

var ifreq_type_names = map[ifreq_type]string{
	ifreq_TUNSETIFF:     "TUNSETIFF",
	ifreq_TUNSETPERSIST: "TUNSETPERSIST",
	ifreq_GETIFINDEX:    "GETIFINDEX",
	ifreq_GETIFFLAGS:    "GETIFFLAGS",
	ifreq_SETIFFLAGS:    "SETIFFLAGS",
	ifreq_GETIFHWADDR:   "GETIFHWADDR",
	ifreq_SETIFHWADDR:   "SETIFHWADDR",
	ifreq_SETIFMTU:      "SETIFMTU",
	ifreq_SIFTXQLEN:     "SIOCSIFTXQLEN",
}

func (t ifreq_type) String() string {
	if s, ok := ifreq_type_names[t]; ok {
		return s
	}
	return fmt.Sprintf("0x%x", int(t))
}

// Create tuntap interfaces for all vnet interfaces not marked as special.
func (m *Main) okHi(hi vnet.Hi) (ok bool) { return m.v.HwIfer(hi).IsUnix() }
func (m *Main) okSi(si vnet.Si) bool      { return m.okHi(m.v.SupHi(si)) }

func (i *tuntap_interface) ioctl(fd int, req ifreq_type, arg uintptr) (err error) {
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(req), arg)

	// Ignore set if flags "no such device" error on if down.
	// This can happen when an interfaces moves to a new namespace and is harmless.
	if req == ifreq_SETIFFLAGS && e == syscall.ENODEV {
		return
	}
	if e != 0 {
		err = fmt.Errorf("tuntap ioctl %s %s: %s", i.name, req, e)
	}
	return
}

func (intf *tuntap_interface) flags_synced() bool { return intf.created && intf.flag_sync_done }

// Set flags and operational state when vnet-owned tuntap interface becomes ready.
func (intf *tuntap_interface) sync_flags() {
	// For startup discovery, interface has not been created yet
	if intf.provision_fd < 0 {
		return
	}
	intf.flag_sync_in_progress = true
	if err := intf.set_flags(); err != nil {
		panic(err)
	}
}

func (intf *tuntap_interface) check_flag_sync_done(msg *netlink.IfInfoMessage) {
	ok := iff_flag(msg.IfInfomsg.Flags)&(iff_up|iff_running|iff_lower_up) == intf.flags
	intf.flag_sync_done = ok
	intf.flag_sync_in_progress = !ok
}

func (m *Main) SwIfAddDel(v *vnet.Vnet, si vnet.Si, isDel bool) (err error) {
	isTun := m.si_is_vnet_tun(si)
	ok := isTun
	hi := vnet.HiNil
	if !isTun {
		hi = m.v.SupHi(si)
		ok = m.okHi(hi)
	}

	if !ok {
		// Unknown interface punts get sent to error node.
		m.rx_node.set_next(si, rx_node_next_error)
		return
	}

	// Tuntap interfaces are never deleted; only created.
	if isDel {
		return
	}

	if si.IsSwSubInterface(m.v) {
		return
	}
	if hi != vnet.HiNil && !hi.IsProvisioned(m.v) {
		return
	}

	intf := &tuntap_interface{
		m:     m,
		hi:    hi,
		si:    si,
		isTun: isTun,
	}

	name := si.Name(v)
	if isTun {
		name = m.vnet_tun_main.linux_interface_name
	}
	copy(intf.name[:], name)
	intf.elog_name = elog.SetString(name)

	if m.vnet_tuntap_interface_by_si == nil {
		m.vnet_tuntap_interface_by_si = make(map[vnet.Si]*tuntap_interface)
	}
	m.vnet_tuntap_interface_by_si[si] = intf

	key := tuntap_address_key(name, uint(si.Id(v)))
	if !isTun {
		key = string(hi.GetAddress(v))
	}
	if m.vnet_tuntap_interface_by_address == nil {
		m.vnet_tuntap_interface_by_address = make(map[string]*tuntap_interface)
	}
	m.vnet_tuntap_interface_by_address[key] = intf
	return
}

func (m *Main) netlink_discovery_done_for_all_namespaces() (err error) {
	nm := &m.net_namespace_main

	// Create any VNET interfaces that were not found via netlink discovery.
	for si, intf := range m.vnet_tuntap_interface_by_si {
		if _, ok := nm.interface_by_si[si]; !ok {
			// vnet-NS devices already have namespace set.
			if intf.namespace == nil {
				intf.namespace = &nm.default_namespace
			}
			err = intf.init(m)
			if err != nil {
				return
			}
		}
	}

	// Initialize other previously existing vnet tuntap interfaces.
	for _, nsi := range nm.interface_by_si {
		if intf := nsi.tuntap; intf != nil {
			intf.namespace = nsi.namespace
			err = intf.init(m)
			if err != nil {
				return
			}

			intf.get_flags()
		}
	}

	// Perform all defered registrations for unix interfaces.
	for _, hw := range m.registered_hwifer_by_si {
		m.RegisterHwInterface(hw)
	}

	return
}

func (intf *tuntap_interface) open_sockets() (err error) {
	ns := intf.namespace
	err, _ = elib.WithNamespace(ns.ns_fd, ns.m.default_namespace.ns_fd, syscall.CLONE_NEWNET, func() (err error) {
		if intf.dev_net_tun_fd, err = syscall.Open("/dev/net/tun", syscall.O_RDWR, 0); err != nil {
			err = fmt.Errorf("tuntap open /dev/net/tun: %s", err)
			return
		}
		if intf.provision_fd == -1 {
			if intf.provision_fd, err = syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(eth_p_all)); err != nil {
				err = fmt.Errorf("tuntap socket AF_PACKET: %s", err)
				return
			}
		}
		return
	})
	return
}

// Create interface (set flags) and make persistent (e.g. interface stays around when we die).
func (intf *tuntap_interface) create() (err error) {
	r := ifreq_flags{name: intf.name}
	r.flags = iff_no_pi
	if intf.isTun {
		r.flags |= iff_tun
	} else {
		r.flags |= iff_tap
	}
	if err = intf.ioctl(intf.dev_net_tun_fd, ifreq_TUNSETIFF, uintptr(unsafe.Pointer(&r))); err != nil {
		return
	}
	if !intf.created {
		intf.created = true
		if err = intf.ioctl(intf.dev_net_tun_fd, ifreq_TUNSETPERSIST, 1); err != nil {
			return
		}
	}
	return
}

// Bind the provisioning socket to the interface.
func (intf *tuntap_interface) bind() (err error) {
	sa := syscall.SockaddrLinklayer{
		Ifindex:  int(intf.ifindex),
		Protocol: eth_p_all,
	}
	if err = syscall.Bind(intf.provision_fd, &sa); err != nil {
		err = fmt.Errorf("tuntap bind: %s", err)
	}
	return
}

func (intf *tuntap_interface) start_up() {
	if intf.to_tx == nil {
		intf.to_tx = make(chan *tx_packet_vector, vnet.MaxOutstandingTxRefs)
	}
	intf.Fd = intf.dev_net_tun_fd
	iomux.Add(intf)
}

func (intf *tuntap_interface) init(m *Main) (err error) {
	// Initial-state.
	intf.dev_net_tun_fd = -1
	intf.provision_fd = -1
	intf.Fd = -1

	err = intf.open_sockets()
	defer func() {
		if err != nil {
			intf.close(true)
		}
	}()

	// Create interface (set flags) and make persistent (e.g. interface stays around when we die).
	if err = intf.create(); err != nil {
		return
	}

	intf.bind()

	// Increase transmit queue.  Default of 500 in drivers/net/tun.c causes packet loss at high rates.
	// FIXME investigate this more.  Maybe MaxOutstandingTxRefs is enough?
	{
		r := ifreq_int{name: intf.name}
		r.i = 5000
		if err = intf.ioctl(intf.provision_fd, ifreq_SIFTXQLEN, uintptr(unsafe.Pointer(&r))); err != nil {
			return
		}
	}

	if eifer, ok := intf.is_ethernet(m); ok {
		if err = intf.configure_ethernet(m, eifer); err != nil {
			return
		}
	}

	// Hook up unix rx node to interface transmit node or inject for tun.
	{
		var next rx_node_next
		if intf.isTun {
			intf.set_mtu(m, m.mtuBytes)
			next = rx_node_next_inject_ip
		} else {
			next = rx_node_next(m.v.AddNamedNext(&m.rx_node, intf.Name()))
		}
		m.rx_node.set_next(intf.si, next)
	}

	if intf.isTun || intf.hi.IsLinkUp(intf.m.v) {
		intf.start_up()
	} else {
		intf.close(false)
	}

	return
}

func (intf *tuntap_interface) is_ethernet(m *Main) (eifer ethernet.HwInterfacer, yes bool) {
	if intf.hi == vnet.HiNil {
		return
	}
	eifer, yes = m.v.HwIfer(intf.hi).(ethernet.HwInterfacer)
	return
}

func (intf *tuntap_interface) set_mtu(m *Main, max_packet_size uint) (err error) {
	if max_packet_size == 0 {
		max_packet_size = m.mtuBytes
	}
	intf.mtuBytes = max_packet_size
	r := ifreq_int{name: intf.name}
	r.i = int(intf.mtuBytes)
	if err = intf.ioctl(intf.provision_fd, ifreq_SETIFMTU, uintptr(unsafe.Pointer(&r))); err != nil {
		return
	}
	intf.setMtu(m, intf.mtuBytes)
	return
}

func (intf *tuntap_interface) configure_ethernet(m *Main, eifer ethernet.HwInterfacer) (err error) {
	ei := eifer.GetInterface()

	if err = intf.set_mtu(m, ei.MaxPacketSize()); err != nil {
		return
	}

	// For tap interfaces, set ethernet address of interface.
	{
		r := ifreq_sockaddr_any{name: intf.name}
		r.sockaddr.Addr.Family = syscall.ARPHRD_ETHER

		// Only set address if it changes.  If address is reset to same value, kernel will remove arps for some reason.
		if err = intf.ioctl(intf.provision_fd, ifreq_GETIFHWADDR, uintptr(unsafe.Pointer(&r))); err != nil {
			err = fmt.Errorf("%s: %s", err, &ei.Address)
			return
		}
		same_address := true
		for i := range ei.Address {
			same_address = r.sockaddr.Addr.Data[i] == int8(ei.Address[i])
			if !same_address {
				break
			}
		}
		if !same_address {
			for i := range ei.Address {
				r.sockaddr.Addr.Data[i] = int8(ei.Address[i])
			}
			if err = intf.ioctl(intf.provision_fd, ifreq_SETIFHWADDR, uintptr(unsafe.Pointer(&r))); err != nil {
				err = fmt.Errorf("%s: %s", err, &ei.Address)
				return
			}
		}
	}

	return
}

func (intf *tuntap_interface) set_flags() (err error) {
	r := ifreq_int{
		name: intf.name,
		i:    int(intf.flags),
	}
	err = intf.ioctl(intf.provision_fd, ifreq_SETIFFLAGS, uintptr(unsafe.Pointer(&r)))
	return
}

func (intf *tuntap_interface) get_flags() (err error) {
	r := ifreq_int{
		name: intf.name,
	}
	err = intf.ioctl(intf.provision_fd, ifreq_GETIFFLAGS, uintptr(unsafe.Pointer(&r)))
	intf.flags = iff_flag(r.i)
	isUp := intf.flags&iff_up != 0
	err = intf.si.SetAdminUp(intf.m.v, isUp)
	intf.flag_sync_in_progress = false
	intf.flag_sync_done = true
	return
}

func (m *Main) maybeChangeFlag(intf *tuntap_interface, isUp bool, flag iff_flag) (err error) {
	change := false
	switch {
	case isUp && intf.flags&flag != flag:
		change = true
		intf.flags |= flag
	case !isUp && intf.flags&flag != 0:
		change = true
		intf.flags &^= flag
	}
	if change && intf.flags_synced() {
		err = intf.set_flags()
	}
	return
}

func (m *Main) SwIfAdminUpDown(v *vnet.Vnet, si vnet.Si, isUp bool) (err error) {
	if intf, ok := m.vnet_tuntap_interface_by_si[si]; ok {
		if err = m.maybeChangeFlag(intf, isUp, iff_up|iff_running); err != nil {
			return
		}
	}
	return
}

func (m *Main) HwIfLinkUpDown(v *vnet.Vnet, hi vnet.Hi, isUp bool) (err error) {
	if !m.okHi(hi) {
		return
	}
	si := v.HwIf(hi).Si()
	intf, ok := m.vnet_tuntap_interface_by_si[si]
	if !ok || !intf.created {
		return
	}
	if isUp {
		intf.open_sockets()
		intf.create()
		intf.start_up()
	} else {
		intf.close(false)
	}
	return
}

var eth_p_all = uint16(vnet.Uint16(syscall.ETH_P_ALL).FromHost())

func (m *Main) Init() (err error) {
	// Suitable defaults for an Ethernet-like tun/tap device.
	m.mtuBytes = 9216

	m.v.RegisterSwIfAddDelHook(m.SwIfAddDel)
	m.v.RegisterSwIfAdminUpDownHook(m.SwIfAdminUpDown)
	m.v.RegisterHwIfLinkUpDownHook(m.HwIfLinkUpDown)
	return
}
