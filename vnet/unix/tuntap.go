// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux

package unix

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/iomux"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/internal/netlink"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"

	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

type tuntap_interface struct {
	m *Main
	// Namespace this interface is currently in.
	namespace *net_namespace
	// Raw socket bound to this interface used for provisioning.
	provision_fd               int
	dev_net_tun_fd             int
	iomux.File                 // /dev/net/tun fd for this interface.
	hi                         vnet.Hi
	si                         vnet.Si
	name                       ifreq_name
	ifindex                    uint32 // linux interface index
	flags                      iff_flag
	set_flags_at_first_up_down bool
	mtuBytes                   uint
	mtuBuffers                 uint
	operState                  netlink.IfOperState
}

//go:generate gentemplate -d Package=unix -id ifVec -d VecType=interfaceVec -d Type=*tuntap_interface github.com/platinasystems/go/elib/vec.tmpl

func (m *Main) interfaceForSi(si vnet.Si) *tuntap_interface { return m.ifVec[si] }

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

func AddExternalInterface(v *vnet.Vnet, ifIndex int, si vnet.Si) {
	m := GetMain(v)
	if m.siByIfIndex == nil {
		m.siByIfIndex = make(map[int]vnet.Si)
	}
	m.siByIfIndex[ifIndex] = si
}

func (i *tuntap_interface) add_del_namespace(m *Main, ns *net_namespace, is_del bool) {
	if is_del {
	}
}

type tuntapMain struct {
	// Selects whether we create tun or tap interfaces.
	isTun bool

	mtuBytes uint

	ifVec  interfaceVec
	ifBySi map[vnet.Si]*tuntap_interface

	bufferPool *vnet.BufferPool
}

func (m *tuntapMain) Init(v *vnet.Vnet) {
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
	if e != 0 {
		err = fmt.Errorf("tuntap ioctl %s: %s", req, e)
	}
	return
}

func (i *tuntap_interface) setOperState(from_init bool) {
	return
	os := netlink.IF_OPER_DOWN
	if i.si.IsAdminUp(i.m.v) && i.hi.IsLinkUp(i.m.v) {
		os = netlink.IF_OPER_UP
	}
	if os != i.operState || from_init {
		i.operState = os
		if !i.is_initialized() {
			return
		}
		msg := netlink.NewIfInfoMessage()
		msg.Header.Type = netlink.RTM_SETLINK
		msg.Header.Flags = netlink.NLM_F_REQUEST
		msg.Index = uint32(i.ifindex)
		msg.L2IfType = uint16(netlink.ARPHRD_ETHER)
		msg.Attrs[netlink.IFLA_OPERSTATE] = os
		i.namespace.NetlinkTx(msg, false)
	}
}

func (m *Main) SwIfAddDel(v *vnet.Vnet, si vnet.Si, isDel bool) (err error) {
	hi := m.v.SupHi(si)
	if !m.okHi(hi) {
		// Unknown interface punts get sent to error node.
		m.rx_node.set_next(si, rx_node_next_error)
		return
	}

	// Tuntap interfaces are never deleted; only created.
	if isDel {
		return
	}

	intf := &tuntap_interface{
		m:  m,
		hi: hi,
		si: si,
	}

	name := si.Name(v)
	copy(intf.name[:], name)

	m.ifVec.Validate(uint(intf.si))
	m.ifVec[intf.si] = intf
	if m.ifBySi == nil {
		m.ifBySi = make(map[vnet.Si]*tuntap_interface)
	}
	m.ifBySi[intf.si] = intf

	if m.tuntap_interface_by_name == nil {
		m.tuntap_interface_by_name = make(map[string]*tuntap_interface)
	}
	m.tuntap_interface_by_name[name] = intf
	return
}

func (intf *tuntap_interface) is_initialized() bool { return intf.namespace != nil }

func (m *Main) netlink_dump_done_for_all_namespaces() (err error) {
	nm := &m.net_namespace_main
	for _, intf := range m.ifBySi {
		name := intf.name.String()
		ns, i := nm.interface_by_name(name)
		if ns == nil {
			ns = &nm.default_namespace
		}
		if i != nil {
			i.tuntap = intf
		}
		intf.namespace = ns
		err = intf.create(m)
		if err != nil {
			return
		}

		next := m.v.AddNamedNext(&nm.rx_node, intf.Name())
		m.rx_node.set_next(intf.si, rx_node_next(next))
	}
	return
}

func (intf *tuntap_interface) create(m *Main) (err error) {
	ns := intf.namespace

	// Create provisioning socket in given namespace.
	elib.WithNamespace(ns.ns_fd, ns.m.default_namespace.ns_fd, syscall.CLONE_NEWNET, func() (err error) {
		if intf.dev_net_tun_fd, err = syscall.Open("/dev/net/tun", syscall.O_RDWR, 0); err != nil {
			err = fmt.Errorf("tuntap /dev/net/tun open: %s", err)
			return
		}
		if intf.provision_fd, err = syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(eth_p_all)); err != nil {
			err = fmt.Errorf("tuntap socket AF_PACKET: %s", err)
		}
		return
	})
	defer func() {
		if err != nil {
			if intf.provision_fd != 0 {
				syscall.Close(intf.provision_fd)
				intf.provision_fd = -1
			}
			if intf.dev_net_tun_fd != 0 {
				syscall.Close(intf.dev_net_tun_fd)
				intf.provision_fd = -1
			}
		}
	}()

	// Create interface (set flags) and make persistent (e.g. interface stays around when we die).
	{
		r := ifreq_flags{name: intf.name}
		r.flags = iff_no_pi
		if m.isTun {
			r.flags |= iff_tun
		} else {
			r.flags |= iff_tap
		}
		if err = intf.ioctl(intf.dev_net_tun_fd, ifreq_TUNSETIFF, uintptr(unsafe.Pointer(&r))); err != nil {
			return
		}
		if err = intf.ioctl(intf.dev_net_tun_fd, ifreq_TUNSETPERSIST, 0); err != nil {
			return
		}
	}

	// Find linux interface index.
	{
		r := ifreq_int{name: intf.name}
		if err = intf.ioctl(intf.provision_fd, ifreq_GETIFINDEX, uintptr(unsafe.Pointer(&r))); err != nil {
			return
		}
		intf.ifindex = uint32(r.i)
		if ns.tuntap_interface_by_ifindex == nil {
			ns.tuntap_interface_by_ifindex = make(map[uint32]*tuntap_interface)
		}
		ns.tuntap_interface_by_ifindex[intf.ifindex] = intf
	}

	// Bind the provisioning socket to the interface.
	{
		sa := syscall.SockaddrLinklayer{
			Ifindex:  int(intf.ifindex),
			Protocol: eth_p_all,
		}
		if err = syscall.Bind(intf.provision_fd, &sa); err != nil {
			err = fmt.Errorf("tuntap bind: %s", err)
			return
		}
	}

	if !intf.set_flags_at_first_up_down {
		// Fetch initial interface flags.
		r := ifreq_int{name: intf.name}
		if err = intf.ioctl(intf.provision_fd, ifreq_GETIFFLAGS, uintptr(unsafe.Pointer(&r))); err != nil {
			return
		}
		intf.flags = iff_flag(r.i)
	}

	if eifer, ok := m.v.HwIfer(intf.hi).(ethernet.HwInterfacer); ok {
		ei := eifer.GetInterface()

		// Set MTU.
		{
			intf.mtuBytes = ei.MaxPacketSize()
			if intf.mtuBytes == 0 {
				intf.mtuBytes = m.mtuBytes
			}
			r := ifreq_int{name: intf.name}
			r.i = int(intf.mtuBytes)
			if err = intf.ioctl(intf.provision_fd, ifreq_SETIFMTU, uintptr(unsafe.Pointer(&r))); err != nil {
				return
			}
			intf.setMtu(m, intf.mtuBytes)
		}

		// Increase transmit queue.  Default of 500 in drivers/net/tun.c causes packet loss at high rates.
		{
			r := ifreq_int{name: intf.name}
			r.i = 5000
			if err = intf.ioctl(intf.provision_fd, ifreq_SIFTXQLEN, uintptr(unsafe.Pointer(&r))); err != nil {
				return
			}
		}

		// For tap interfaces, set ethernet address of interface.
		if !m.isTun {
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
	}

	intf.setOperState(true)

	intf.Fd = intf.provision_fd
	if err := syscall.SetsockoptInt(intf.Fd, syscall.SOL_SOCKET, syscall.SO_RCVBUF, 1<<20); err != nil {
		panic(err)
	}
	intf.File.SetReadOnly()
	iomux.Add(intf)

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
	// fixme use netlink
	if change {
		if intf.is_initialized() {
			err = intf.set_flags()
		} else {
			intf.set_flags_at_first_up_down = true
		}
	}
	return
}

func (m *Main) SwIfAdminUpDown(v *vnet.Vnet, si vnet.Si, isUp bool) (err error) {
	if !m.okSi(si) {
		return
	}
	intf := m.interfaceForSi(si)
	if intf.is_initialized() && intf.set_flags_at_first_up_down {
		intf.set_flags_at_first_up_down = false
		err = intf.set_flags()
		return
	}
	err = m.maybeChangeFlag(intf, isUp, iff_up|iff_running)
	if err != nil {
		return
	}
	intf.setOperState(false)
	return
}

func (m *Main) HwIfLinkUpDown(v *vnet.Vnet, hi vnet.Hi, isUp bool) (err error) {
	if !m.okHi(hi) {
		return
	}
	intf := m.interfaceForSi(v.HwIf(hi).Si())
	err = m.maybeChangeFlag(intf, isUp, iff_lower_up)
	if err != nil {
		return
	}
	intf.setOperState(false)
	return
}

var eth_p_all = uint16(vnet.Uint16(syscall.ETH_P_ALL).FromHost())

func (i *tuntap_interface) close() (err error) {
	err = syscall.Close(i.provision_fd)
	i.provision_fd = -1
	return
}

func (m *Main) Init() (err error) {
	// Suitable defaults for an Ethernet-like tun/tap device.
	m.mtuBytes = 4096 + 256

	m.v.RegisterSwIfAddDelHook(m.SwIfAddDel)
	m.v.RegisterSwIfAdminUpDownHook(m.SwIfAdminUpDown)
	m.v.RegisterHwIfLinkUpDownHook(m.HwIfLinkUpDown)
	return
}

func (m *Main) Configure(in *parse.Input) {
	for !in.End() {
		switch {
		case in.Parse("mtu %d", &m.mtuBytes):
		case in.Parse("tap"):
			m.isTun = false
		case in.Parse("tun"):
			m.isTun = true
		case in.Parse("verbose-packets"):
			m.verbosePackets = true
		case in.Parse("verbose-netlink"):
			m.verboseNetlink++
		default:
			panic(parse.ErrInput)
		}
	}
}
