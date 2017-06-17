// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/internal/netlink"
	"github.com/platinasystems/go/vnet"

	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

type inotify_event struct {
	watch_descriptor int32
	mask             uint32
	cookie           uint32
	len              uint32
}

func decode(b []byte, i int) (e *inotify_event, name string, i_next int) {
	e = (*inotify_event)(unsafe.Pointer(&b[i]))
	j := i + 16
	len := strings.IndexByte(string(b[j:]), 0)
	name = string(b[j : j+len])
	i_next = j + int(e.len)
	return
}

func (m *net_namespace_main) read_dir(dir_name string, f func(dir, name string, is_del bool)) (err error) {
	// Collect existing files in directory.
	var fis []os.FileInfo
	if fis, err = ioutil.ReadDir(dir_name); err != nil {
		return
	}
	for _, fi := range fis {
		f(dir_name, fi.Name(), false)
	}
	return
}

const (
	netnsDir               = "/var/run/netns"
	default_namespace_name = "default"
)

func (m *netlink_main) namespace_register_nodes() {
	nm := &m.net_namespace_main
	nm.m = m.m
	nm.rx_node.init(m.m.v)
	nm.tx_node.init(nm)
}

func (m *netlink_main) namespace_init() (err error) {
	nm := &m.net_namespace_main

	// Handcraft default name space.
	{
		ns := &nm.default_namespace
		ns.m = nm
		ns.name = default_namespace_name
		ns.nsid = -1
		ns.is_default = true
		if ns.ns_fd, err = nm.fd_for_path("", "/proc/self/ns/net"); err != nil {
			return
		}

		ns.index = nm.namespace_pool.GetIndex()
		nm.namespace_pool.entries[ns.index] = ns

		m.namespace_by_name = make(map[string]*net_namespace)
		m.namespace_by_name[ns.name] = ns

		if err = ns.configure(-1, -1); err != nil {
			return
		}
		ns.listen(m)
		ns.fibInit(false)
	}

	// Setup initial namespaces.
	m.read_dir(netnsDir, m.watch_namespace_add_del)
	m.n_namespace_discovered_at_init = uint32(len(nm.namespace_by_name))
	return
}

// True when all namespaces have been discovered.
func (m *net_namespace_main) discovery_is_done() bool {
	return m.n_namespace_discovery_done > 0 && m.n_namespace_discovery_done >= m.n_namespace_discovered_at_init
}

// Called when initial netlink dump via netlink.Listen is done.
func (ns *net_namespace) netlink_dump_done(m *Main) (err error) {
	nm := &m.net_namespace_main
	if atomic.AddUint32(&nm.n_namespace_discovery_done, 1) == nm.n_namespace_discovered_at_init {
		err = m.netlink_discovery_done_for_all_namespaces()
	}
	return
}

func (m *netlink_main) watch_for_new_net_namespaces() {
	go m.watch_dir(netnsDir, m.watch_namespace_add_del)
}

func (m *net_namespace_main) watch_dir(dir_name string, f func(dir, name string, is_del bool)) (err error) {
	var fd, n int

	// Watch for new files added and existing files deleted.
	fd, err = syscall.InotifyInit()
	if err != nil {
		err = os.NewSyscallError("inotify init", err)
		return
	}

	if _, err = syscall.InotifyAddWatch(fd, dir_name, syscall.IN_CREATE|syscall.IN_DELETE); err != nil {
		err = os.NewSyscallError("inotify add watch", err)
		return
	}

	for {
		var buf [4096]byte
		if n, err = syscall.Read(fd, buf[:]); err != nil {
			panic(err)
		}
		for i := 0; i < n; {
			e, name, i_next := decode(buf[:], i)
			switch {
			case e.mask&syscall.IN_CREATE != 0:
				f(dir_name, name, false)
			case e.mask&syscall.IN_DELETE != 0:
				f(dir_name, name, true)
			}
			i = i_next
		}
	}
}

type net_namespace_interface struct {
	name      string
	namespace *net_namespace
	ifindex   uint32
	address   []byte
	kind      netlink.InterfaceKind
	si        vnet.Si
	tuntap    *tuntap_interface
}

type net_namespace struct {
	m *net_namespace_main

	// Unique index allocated from index_pool.
	index uint

	name string

	// File descriptor for /proc/self/ns/net for default name space or /var/run/netns/NAME.
	ns_fd int

	nsid int

	mu sync.Mutex

	vnet_tuntap_interface_by_ifindex map[uint32]*tuntap_interface
	dummy_interface_by_ifindex       map[uint32]*dummy_interface
	si_by_ifindex                    map[uint32]vnet.Si

	is_default bool

	netlink_socket_fds [2]int
	netlink_socket_pair

	current_event *netlinkEvent

	interface_by_index map[uint32]*net_namespace_interface
	interface_by_name  map[string]*net_namespace_interface
}

//go:generate gentemplate -d Package=unix -id net_namespace -d PoolType=net_namespace_pool -d Type=*net_namespace -d Data=entries github.com/platinasystems/go/elib/pool.tmpl

type net_namespace_main struct {
	m                                *Main
	default_namespace                net_namespace
	namespace_by_name                map[string]*net_namespace
	vnet_tuntap_interface_by_si      map[vnet.Si]*tuntap_interface
	vnet_tuntap_interface_by_address map[string]*tuntap_interface
	n_namespace_discovery_done       uint32
	n_namespace_discovered_at_init   uint32
	interface_by_si                  map[vnet.Si]*net_namespace_interface
	registered_hwifer_by_si          map[vnet.Si]vnet.HwInterfacer
	registered_hwifer_by_address     map[string]vnet.HwInterfacer
	namespace_pool                   net_namespace_pool
	rx_node                          rx_node
	tx_node                          tx_node
}

func (m *net_namespace_main) fd_for_path(dir, name string) (fd int, err error) {
	fd, err = syscall.Open(path.Join(dir, name), syscall.O_RDONLY, 0)
	return
}

func (m *netlink_main) nsid_for_path(dir, name string) (nsid int, err error) {
	var fd int
	if fd, err = m.fd_for_path(dir, name); err != nil {
		return
	}
	defer syscall.Close(fd)

	req := netlink.NewNetnsMessage()
	req.Type = netlink.RTM_GETNSID
	req.Flags = netlink.NLM_F_REQUEST
	req.AddressFamily = netlink.AF_UNSPEC
	req.Attrs[netlink.NETNSA_FD] = netlink.Uint32Attr(fd)
	rep := m.default_namespace.NetlinkTx(req, true)
	nsid = netlink.DefaultNsid
	switch v := rep.(type) {
	case *netlink.NetnsMessage:
		nsid = int(v.Attrs[netlink.NETNSA_NSID].(netlink.Int32Attr).Int())
	}
	return
}

func (e *netlinkEvent) netnsMessage(msg *netlink.NetnsMessage) (err error) {
	// Re-read directory to refresh name to nsid mapping.
	e.m.read_dir(netnsDir, e.m.watch_namespace_add_del)
	return
}

func (m *netlink_main) add_del_nsid(name string, nsid int, is_del bool) {
	if is_del {
		delete(m.namespace_by_name, name)
	} else {
		ns := &net_namespace{name: name, nsid: nsid}
		m.namespace_by_name[name] = ns
		ns.add(m)
	}
}

func (m *netlink_main) watch_namespace_add_del(dir, name string, is_del bool) {
	var (
		nsid int
		err  error
	)
	if !is_del {
		if nsid, err = m.nsid_for_path(dir, name); err != nil {
			panic(err)
		}
	} else {
		var ok bool
		if _, ok = m.namespace_by_name[name]; !ok {
			panic("delete unknown namespace " + name)
		}
	}
	m.add_del_nsid(name, nsid, is_del)
}

func (ns *net_namespace) add_del_interface(m *Main, msg *netlink.IfInfoMessage) {
	is_del := false
	switch msg.Header.Type {
	case netlink.RTM_NEWLINK:
	case netlink.RTM_DELLINK:
		is_del = true
	default:
		return
	}
	name := msg.Attrs[netlink.IFLA_IFNAME].String()
	address := msg.Attrs[netlink.IFLA_ADDRESS].(*netlink.EthernetAddress).Bytes()
	index := msg.Index
	if !is_del {
		if ns.interface_by_index == nil {
			ns.interface_by_index = make(map[uint32]*net_namespace_interface)
			ns.interface_by_name = make(map[string]*net_namespace_interface)
		}
		intf, exists := ns.interface_by_index[index]
		name_changed := false
		if !exists {
			intf = &net_namespace_interface{
				namespace: ns,
				name:      name,
				ifindex:   index,
				kind:      msg.InterfaceKind(),
				si:        vnet.SiNil,
			}
			ns.interface_by_index[index] = intf
			ns.interface_by_name[name] = intf
		} else {
			name_changed = intf.name != name
		}
		if exists && string(intf.address) != string(address) {
			// fixme address change
		}
		intf.address = make([]byte, len(address))
		copy(intf.address[:], address[:])
		if name_changed {
			delete(ns.interface_by_name, name)
			ns.interface_by_name[name] = intf
			intf.name = name

			// Change name of corresponding vnet interface.
			if tif, ok := ns.vnet_tuntap_interface_by_ifindex[index]; ok {
				tif.set_name(name)
			}
		}

		// Ethernet address uniquely identifies register hw interfaces.
		if h, ok := m.registered_hwifer_by_address[string(address)]; ok {
			m.set_si(intf, h.GetHwIf().Si())
		}

		if tif, ok := m.vnet_tuntap_interface_by_address[string(address)]; ok {
			m.set_si(intf, tif.si)
			intf.tuntap = tif

			if ns.vnet_tuntap_interface_by_ifindex == nil {
				ns.vnet_tuntap_interface_by_ifindex = make(map[uint32]*tuntap_interface)
			}
			tif.ifindex = index
			ns.vnet_tuntap_interface_by_ifindex[tif.ifindex] = tif

			if tif.created && !tif.flag_sync_done && !tif.flag_sync_in_progress {
				tif.sync_flags()
			}

			// Interface moved to a new namespace?
			if tif.namespace != ns {
				tif.add_del_namespace(m, ns, is_del)
			}
		}
	} else {
		intf := ns.interface_by_index[index]
		if tif := intf.tuntap; tif != nil {
			tif.add_del_namespace(m, ns, is_del)
			tif.namespace = nil
		}
		if intf.si != vnet.SiNil {
			delete(ns.si_by_ifindex, index)
			delete(m.interface_by_si, intf.si)
		}
		delete(ns.interface_by_index, index)
		delete(ns.interface_by_name, name)
	}
}

func (m *net_namespace_main) interface_by_name(name string) (ns *net_namespace, intf *net_namespace_interface) {
	for _, s := range m.namespace_by_name {
		if i, ok := s.interface_by_name[name]; ok {
			ns, intf = s, i
			break
		}
	}
	return
}

func (m *net_namespace_main) set_si(intf *net_namespace_interface, si vnet.Si) {
	intf.si = si

	ns := intf.namespace

	// Set up ifindex to vnet Si mapping.
	if ns.si_by_ifindex == nil {
		ns.si_by_ifindex = make(map[uint32]vnet.Si)
	}
	ns.si_by_ifindex[intf.ifindex] = si

	// Set up si to interface mapping.
	if m.interface_by_si == nil {
		m.interface_by_si = make(map[vnet.Si]*net_namespace_interface)
	}
	m.interface_by_si[si] = intf
}

func (m *net_namespace_main) RegisterHwInterface(h vnet.HwInterfacer) {
	hw := h.GetHwIf()
	si := hw.Si()
	// Defer registration until after discovery is done.
	if m.registered_hwifer_by_si == nil {
		m.registered_hwifer_by_si = make(map[vnet.Si]vnet.HwInterfacer)
	}
	m.registered_hwifer_by_si[si] = h

	if !m.discovery_is_done() {
		return
	}

	ns, intf := m.interface_by_name(hw.Name())
	if ns == nil {
		panic("unknown interface: " + hw.Name())
	}
	m.set_si(intf, si)
	h.SetAddress(intf.address)

	if m.registered_hwifer_by_address == nil {
		m.registered_hwifer_by_address = make(map[string]vnet.HwInterfacer)
	}
	m.registered_hwifer_by_address[string(intf.address)] = h
}

func (ns *net_namespace) String() (s string) {
	s = ns.name
	if s == "" {
		s = "default-namespace"
	}
	return
}

type showNsLine struct {
	Interface string `format:"%-30s"`
	Type      string `format:"%s" align:"center"`
	Namespace string `format:"%s" align:"center"`
	NSID      string `format:"%s" align:"center"`
	si        vnet.Si
}
type showNsLines struct {
	lines []showNsLine
	v     *vnet.Vnet
}

func (ns showNsLines) Less(i, j int) bool {
	ni, nj := &ns.lines[i], &ns.lines[j]
	if ni.Namespace == nj.Namespace {
		if ni.si == vnet.SiNil || nj.si == vnet.SiNil {
			return ni.Interface < nj.Interface
		}
		ifi, ifj := ns.v.SwIf(ni.si), ns.v.SwIf(nj.si)
		return ns.v.SwLessThan(ifi, ifj)
	}
	return ni.Namespace < nj.Namespace
}
func (ns showNsLines) Swap(i, j int) { ns.lines[i], ns.lines[j] = ns.lines[j], ns.lines[i] }
func (ns showNsLines) Len() int      { return len(ns.lines) }

func (m *netlink_main) show_net_namespaces(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	ms := showNsLines{v: m.m.v}
	for _, ns := range m.namespace_by_name {
		for _, intf := range ns.interface_by_index {
			x := showNsLine{Namespace: ns.name, Interface: intf.name, Type: intf.kind.String(), si: vnet.SiNil}
			if intf.tuntap != nil {
				x.si = intf.tuntap.si
			}
			if ns.nsid != -1 {
				x.NSID = fmt.Sprintf("%d", ns.nsid)
			}
			ms.lines = append(ms.lines, x)
		}
	}
	sort.Sort(ms)
	colMap := map[string]bool{
		"si": false,
	}
	elib.Tabulate(ms.lines).WriteCols(w, colMap)
	return
}

func (ns *net_namespace) allocate_sockets() (err error) {
	ns.netlink_socket_fds[0], err = syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW, syscall.NETLINK_ROUTE)
	if err == nil {
		ns.netlink_socket_fds[1], err = syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW, syscall.NETLINK_ROUTE)
	}
	return
}

func (m *net_namespace_main) max_n_namespace() uint { return uint(len(m.namespace_by_name)) }

func (ns *net_namespace) add(m *netlink_main) {
	// Allocate unique index for namespace.
	nm := &m.net_namespace_main
	ns.index = nm.namespace_pool.GetIndex()
	nm.namespace_pool.entries[ns.index] = ns

	// Loop until namespace sockets are allocated.
	time_start := time.Now()
	var (
		err               error
		first_setns_errno syscall.Errno
	)
	for {
		ns.m = nm
		if ns.ns_fd, err = m.fd_for_path(netnsDir, ns.name); err != nil {
			panic(err)
		}
		// First setns may return EINVAL until "ip netns add X" performs mount bind; so we need to retry.
		err, first_setns_errno = elib.WithNamespace(ns.ns_fd, m.default_namespace.ns_fd, syscall.CLONE_NEWNET, ns.allocate_sockets)
		if err == nil {
			break
		}
		if time.Since(time_start) > 10*time.Millisecond {
			panic(err)
		}
		syscall.Close(ns.ns_fd)
		ns.ns_fd = -1
		if first_setns_errno == syscall.EINVAL {
			time.Sleep(1 * time.Millisecond)
		}
	}
	if err = ns.netlink_socket_pair.configure(ns.netlink_socket_fds[0], ns.netlink_socket_fds[1]); err != nil {
		panic(err)
	}
	ns.listen(m)
	ns.fibInit(false)
}

func (ns *net_namespace) del(m *netlink_main) {
	ns.m.namespace_pool.PutIndex(ns.index)
	ns.m.namespace_pool.entries[ns.index] = nil
	ns.index = ^uint(0)

	if ns.ns_fd > 0 {
		syscall.Close(ns.ns_fd)
		ns.ns_fd = -1
	}
	ns.netlink_socket_pair.close()
	ns.fibInit(true)
}
