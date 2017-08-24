// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/internal/netlink"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"

	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

func (m *net_namespace_main) read_dir(dir *namespace_search_dir, f func(dir *namespace_search_dir, name string, is_del bool, is_init bool)) (err error) {
	// Collect existing files in /var/run/netns directory.
	// ip netns add X command creates /var/run/netns/X file which when opened becomes ns_fd.
	var fis []os.FileInfo
	if fis, err = ioutil.ReadDir(dir.path); err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
	}
	for _, fi := range fis {
		f(dir, fi.Name(), false, true)
	}
	return
}

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

func (m *net_namespace_main) watch_dir(dir *namespace_search_dir, f func(dir *namespace_search_dir, name string, is_del bool, is_init bool)) (err error) {
	var fd, n int

	// Watch for new files added and existing files deleted.
	fd, err = syscall.InotifyInit()
	if err != nil {
		err = os.NewSyscallError("inotify init", err)
		return
	}

	if _, err = syscall.InotifyAddWatch(fd, dir.path, syscall.IN_CREATE|syscall.IN_DELETE); err != nil {
		err = os.NewSyscallError("inotify add watch", err)
		return
	}

	const is_init = false
	for {
		var buf [4096]byte
		if n, err = syscall.Read(fd, buf[:]); err != nil {
			panic(err)
		}
		for i := 0; i < n; {
			e, name, i_next := decode(buf[:], i)
			switch {
			case e.mask&syscall.IN_CREATE != 0:
				f(dir, name, false, is_init)
			case e.mask&syscall.IN_DELETE != 0:
				f(dir, name, true, is_init)
			}
			i = i_next
		}
	}
}

const (
	default_namespace_name = "default"
)

type namespace_search_dir struct {
	path string
	// Prefix for namespace names.
	prefix string
}

func (d *namespace_search_dir) namespace_name(file_name string) (ns_name string) {
	if d.prefix != "" {
		ns_name = d.prefix + "-" + file_name
	} else {
		ns_name = file_name
	}
	return
}

var netns_search_dirs = [...]namespace_search_dir{
	// docker: do this directory first.
	namespace_search_dir{
		path:   "/var/run/docker/netns",
		prefix: "docker",
	},
	// iproute2
	{
		path: "/var/run/netns",
	},
}

func (m *netlink_main) namespace_register_nodes() {
	nm := &m.m.net_namespace_main
	nm.m = m.m
	nm.rx_node.init(m.m)
	nm.tx_node.init(nm)
}

func (nm *net_namespace_main) init() (err error) {
	// Handcraft default name space.
	{
		ns := &nm.default_namespace
		ns.m = nm
		ns.name = default_namespace_name
		ns.is_default = true
		if ns.ns_fd, err = nm.fd_for_path("", "/proc/self/ns/net"); err != nil {
			return
		}

		ns.index = nm.namespace_pool.GetIndex()
		nm.namespace_pool.entries[ns.index] = ns

		nm.namespace_by_name = make(map[string]*net_namespace)
		nm.namespace_by_name[ns.name] = ns

		if err = ns.netlink_socket_pair.configure(-1, -1); err != nil {
			return
		}

		// Set nsid (if it exists) and inode (which always exists and uniquely identifies namespace).
		ns.nsid, ns.inode, err = nm.nsid_for_fd(ns.ns_fd)
		if err != nil {
			panic(err)
		}
		nm.namespace_by_nsid = make(map[int]*net_namespace)
		nm.namespace_by_nsid[ns.nsid] = ns
		nm.namespace_by_inode = make(map[uint64]*net_namespace)
		nm.namespace_by_inode[ns.inode] = ns

		ns.listen(&nm.m.netlink_main)
		ns.fibInit(false)
		nm.m.vnet_tun_main.create_tun(ns)
	}

	// Setup initial namespaces.
	// 1 for default namespace.
	nm.n_namespace_discovered_at_init = 1
	for i := range netns_search_dirs {
		d := &netns_search_dirs[i]

		// If not existent yet (ie on boot), create the netns search
		// dirs so watch doesn't fail
		if _, err := os.Stat(d.path); os.IsNotExist(err) {
			os.MkdirAll(d.path, os.ModeDir|0755)
		}

		if err = nm.read_dir(d, nm.watch_namespace_add_del); err != nil {
			return
		}
	}
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

func (m *net_namespace_main) watch_for_new_net_namespaces() {
	for i := range netns_search_dirs {
		d := &netns_search_dirs[i]
		go m.watch_dir(d, m.watch_namespace_add_del)
	}
}

type net_namespace_interface struct {
	name                 string
	namespace            *net_namespace
	ifindex              uint32
	address              []byte
	kind                 netlink.InterfaceKind
	tunnel_metadata_mode bool
	si                   vnet.Si
	sup_interface        *net_namespace_interface
	tuntap               *tuntap_interface
}

func tuntap_address_key(name string, index uint) string {
	return fmt.Sprintf("%s-%d", name, index)
}

type si_by_ifindex struct {
	mu sync.RWMutex
	m  map[uint32]vnet.Si
}

func (i *si_by_ifindex) set(x uint32, si vnet.Si) {
	i.mu.Lock()
	i.m[x] = si
	i.mu.Unlock()
}
func (i *si_by_ifindex) unset(x uint32) {
	i.mu.Lock()
	delete(i.m, x)
	i.mu.Unlock()
}
func (i *si_by_ifindex) get(x uint32) (si vnet.Si, ok bool) {
	i.mu.RLock()
	si, ok = i.m[x]
	i.mu.RUnlock()
	return
}

type net_namespace struct {
	m *net_namespace_main

	// Unique index allocated from index_pool.
	index uint

	name string

	// File descriptor for /proc/self/ns/net for default name space or /var/run/netns/NAME.
	ns_fd int

	// NSID from netlink.
	nsid int
	// Inode which uniquely identifies file for namespace.
	// NSID is not required to be set at all.
	inode uint64

	mu sync.Mutex

	vnet_tun_interface               *tuntap_interface
	vnet_tuntap_interface_by_ifindex map[uint32]*tuntap_interface
	dummy_interface_by_ifindex       map[uint32]*dummy_interface
	si_by_ifindex                    si_by_ifindex

	is_default bool

	netlink_socket_fds [2]int
	netlink_socket_pair

	interface_by_index map[uint32]*net_namespace_interface
	interface_by_name  map[string]*net_namespace_interface
}

//go:generate gentemplate -d Package=unix -id net_namespace -d PoolType=net_namespace_pool -d Type=*net_namespace -d Data=entries github.com/platinasystems/go/elib/pool.tmpl

type net_namespace_main struct {
	m                                *Main
	default_namespace                net_namespace
	namespace_by_name                map[string]*net_namespace
	namespace_by_nsid                map[int]*net_namespace
	namespace_by_inode               map[uint64]*net_namespace
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
	tuntap_sendmsg_recvmsg_disable   bool
}

func (m *net_namespace_main) fd_for_path(elem ...string) (fd int, err error) {
	fd, err = syscall.Open(path.Join(elem...), syscall.O_RDONLY, 0)
	return
}

func (m *net_namespace_main) nsid_for_path(elem ...string) (nsid int, inode uint64, err error) {
	var fd int
	if fd, err = m.fd_for_path(elem...); err != nil {
		return
	}
	defer syscall.Close(fd)
	nsid, inode, err = m.nsid_for_fd(fd)
	return
}

func (m *net_namespace_main) nsid_for_fd(fd int) (nsid int, inode uint64, err error) {
	var s syscall.Stat_t
	syscall.Fstat(fd, &s)
	inode = s.Ino

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
	case *netlink.ErrorMessage:
		err = fmt.Errorf("netlink GETNSID: %v", syscall.Errno(v.Errno))
	}
	return
}

type net_namespace_process struct {
	// Namespace for process (nil if unknown).
	ns *net_namespace
	// Command for process (e.g. "bash")
	command string
	// Process id.
	pid uint64
	// Inode /proc/PID/ns/net.
	inode uint64
	// NSID.
	nsid int
}

// Not used now but might be useful someday.
func (m *net_namespace_main) foreachProcFs(f func(p *net_namespace_process)) (err error) {
	var fis []os.FileInfo
	const dir = "/proc"
	if fis, err = ioutil.ReadDir(dir); err != nil {
		return
	}
	for _, fi := range fis {
		if !fi.IsDir() {
			continue
		}
		var p net_namespace_process

		// See if name looks like PID.
		if p.pid, err = strconv.ParseUint(fi.Name(), 10, 0); err != nil {
			err = nil
			continue
		}

		var fd int
		fd, err = m.fd_for_path(dir, fi.Name(), "ns/net")
		if err != nil {
			// Ignore processes which have disappeared.
			if os.IsNotExist(err) {
				continue
			} else {
				return
			}
		}
		p.nsid, p.inode, err = m.nsid_for_fd(fd)
		syscall.Close(fd)
		if err != nil {
			return
		}

		{
			// /proc/PID/stat gives name of process.
			var b []byte
			b, err = ioutil.ReadFile(path.Join(dir, fi.Name(), "stat"))
			if err != nil {
				return
			}
			fs := strings.Fields(string(b))
			p.command = strings.TrimRight(strings.TrimLeft(fs[1], "("), ")")
		}

		p.ns = m.namespace_by_inode[p.inode]
		f(&p)
	}
	return
}

func (e *netlinkEvent) netnsMessage(msg *netlink.NetnsMessage) (err error) {
	// Nothing to do.  Discovery of namespaces is done via files in /var/run
	return
}

func (m *net_namespace_main) add_del_nsid(dir *namespace_search_dir, file_name string, is_del bool) (err error) {
	name := dir.namespace_name(file_name)

	if is_del {
		ns := m.namespace_by_name[name]
		if ns == nil { // delete unknown namespace file
			return
		}
		if ns.nsid != netlink.DefaultNsid {
			delete(m.namespace_by_nsid, ns.nsid)
		}
		delete(m.namespace_by_inode, ns.inode)
		ns.del(m)
		delete(m.namespace_by_name, name)
		return
	}

	var (
		ns *net_namespace
		ok bool
	)
	// If it exists set id; otherwise make a new namespace.
	if ns, ok = m.namespace_by_name[name]; ok {
		ns.nsid, ns.inode, err = m.nsid_for_path(dir.path, file_name)
		return
	}
	ns = &net_namespace{name: name}
	// Namespace may be duplicate.  (e.g. created by docker and then linked to in /var/run/netns)
	if err = ns.add(m, dir.path, file_name); err != nil {
		return
	}
	m.namespace_by_name[name] = ns
	return
}

type add_del_namespace_event struct {
	vnet.Event
	m       *net_namespace_main
	dir     *namespace_search_dir
	name    string
	is_del  bool
	is_init bool
}

func (e *add_del_namespace_event) String() string { return "add-del-namespace-event" }
func (e *add_del_namespace_event) EventAction() {
	if err := e.m.add_del_nsid(e.dir, e.name, e.is_del); err != nil {
		e.m.m.v.Logf("namespace watch: %v %v\n", e.name, err)
	} else if e.is_init {
		e.m.n_namespace_discovered_at_init++
	}
}

func (m *net_namespace_main) watch_namespace_add_del(dir *namespace_search_dir, name string, is_del bool, is_init bool) {
	m.m.v.SignalEvent(&add_del_namespace_event{m: m, dir: dir, name: name, is_del: is_del, is_init: is_init})
}

func (ns *net_namespace) add_del_interface(m *Main, msg *netlink.IfInfoMessage) (err error) {
	is_del := false
	switch msg.Header.Type {
	case netlink.RTM_NEWLINK:
	case netlink.RTM_DELLINK:
		is_del = true
	default:
		return
	}
	name := msg.Attrs[netlink.IFLA_IFNAME].String()
	var address []byte
	switch a := msg.Attrs[netlink.IFLA_ADDRESS].(type) {
	case *netlink.EthernetAddress:
		address = a.Bytes()
	case *netlink.Ip4Address:
		address = a.Bytes()
	case *netlink.Ip6Address:
		address = a.Bytes()
	}
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

		// For tunnels record if in metadata mode.
		if as := msg.GetLinkInfoData(); as != nil {
			switch intf.kind {
			case netlink.InterfaceKindIp4GRE, netlink.InterfaceKindIp4GRETap,
				netlink.InterfaceKindIp6GRE, netlink.InterfaceKindIp6GRETap:
				intf.tunnel_metadata_mode = as.X[netlink.IFLA_GRE_COLLECT_METADATA] != nil
			case netlink.InterfaceKindIpip, netlink.InterfaceKindIp6Tunnel:
				intf.tunnel_metadata_mode = as.X[netlink.IFLA_IPTUN_COLLECT_METADATA] != nil
			}
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

		is_tuntap := intf.kind == netlink.InterfaceKindTun
		tuntap_key := tuntap_address_key(intf.name, ns.index)
		if len(address) > 0 {
			tuntap_key = string(address)
		}
		if tif, ok := m.vnet_tuntap_interface_by_address[tuntap_key]; is_tuntap && ok {
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
				if err = tif.add_del_namespace(m, ns, is_del); err != nil {
					return
				}
			}
		}

		if !exists && intf.kind == netlink.InterfaceKindVlan {
			m.add_del_vlan(intf, msg, is_del)
		}
	} else {
		intf, ok := ns.interface_by_index[index]
		// Ignore deletes of unknown interface.
		if !ok {
			return
		}
		if tif := intf.tuntap; tif != nil {
			tif.add_del_namespace(m, ns, is_del)
			tif.namespace = nil
		}
		if intf.si != vnet.SiNil {
			if intf.kind == netlink.InterfaceKindVlan {
				m.add_del_vlan(intf, msg, is_del)
			}
			ns.si_by_ifindex.unset(index)
			delete(m.interface_by_si, intf.si)
		}
		delete(ns.interface_by_index, index)
		delete(ns.interface_by_name, name)
	}
	return
}

func (m *net_namespace_main) find_interface_with_ifindex(index uint32) (intf *net_namespace_interface) {
	for _, ns := range m.namespace_by_name {
		if i, ok := ns.interface_by_index[index]; ok {
			if intf != nil {
				panic(fmt.Errorf("interface is not uniquely identified by index %d; index exists in namespaces %s and %s",
					index, intf.namespace.name, ns.name))
			}
			intf = i
		}
	}
	return
}

func (m *net_namespace_main) add_del_vlan(intf *net_namespace_interface, msg *netlink.IfInfoMessage, is_del bool) {
	ns := intf.namespace
	sup_index := msg.Attrs[netlink.IFLA_LINK].(netlink.Uint32Attr).Uint()
	sup_si := vnet.SiNil

	// Look in same namespace as target interface; if not found look in all namespaces (ifindex had better be unique!).
	sup_intf := ns.interface_by_index[sup_index]
	if sup_intf == nil {
		sup_intf = m.find_interface_with_ifindex(sup_index)
	}

	sup_si = sup_intf.si
	intf.sup_interface = sup_intf

	// Sup interface not Vnet interface?
	if sup_si == vnet.SiNil {
		return
	}

	ld := msg.GetLinkInfoData()
	v := ns.m.m.v
	if is_del {
		v.DelSwIf(intf.si)
	} else {
		id := vnet.Uint16(ld.X[netlink.IFLA_VLAN_ID].(netlink.Uint16Attr).Uint())
		var eid ethernet.IfId
		if sup_si.IsSwSubInterface(v) {
			eid = ethernet.IfId(v.SwIf(sup_si).Id(v))
			outer, _ := eid.OuterVlan()
			eid.Set2(outer, id)
		} else {
			eid.Set(id)
		}
		hi := v.SupHi(sup_si)
		hw := v.HwIf(hi)
		si := ns.m.m.v.NewSwSubInterface(hw.Si(), vnet.IfId(eid))
		m.set_si(intf, si)
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
	if ns.si_by_ifindex.m == nil {
		ns.si_by_ifindex.m = make(map[uint32]vnet.Si)
	}
	ns.si_by_ifindex.set(intf.ifindex, si)

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

func (m *netlink_main) show_net_namespaces(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	nm := &m.m.net_namespace_main

	var matching_ns_names []parse.Regexp
	show_procs := false
	detail := false
	for !in.End() {
		var re parse.Regexp
		switch {
		case in.Parse("p%*rocess"):
			show_procs = true
		case in.Parse("d%*etail"):
			detail = true
		case in.Parse("m%*atching %v", &re):
			matching_ns_names = append(matching_ns_names, re)
		case in.Parse("%v", &re):
			matching_ns_names = append(matching_ns_names, re)
		default:
			err = cli.ParseError
			return
		}
	}

	if show_procs {
		type proc struct {
			Command   string `format:"%-40s"`
			Pid       uint64 `format:"%8d"`
			Namespace string `format:"%s" align:"center" width:"30"`
			NSID      string `format:"%s" align:"center"`
			Inode     uint64 `format:"0x%8x" width:"16" align:"center"`
		}
		var ps []proc
		nm.foreachProcFs(func(p *net_namespace_process) {
			x := proc{
				Command: p.command,
				Pid:     p.pid,
				Inode:   p.inode,
			}
			if p.nsid != netlink.DefaultNsid {
				x.NSID = fmt.Sprintf("%d", p.nsid)
			}
			if p.ns != nil {
				x.Namespace = p.ns.name
			}
			if !detail && x.Namespace == default_namespace_name {
				return
			}
			ps = append(ps, x)
		})
		sort.Slice(ps, func(i, j int) bool {
			pi, pj := &ps[i], &ps[j]
			if pi.Namespace == pj.Namespace {
				return pi.Command < pj.Command
			}
			return pi.Namespace < pj.Namespace
		})
		elib.Tabulate(ps).Write(w)
		return
	}

	type nsIf struct {
		si        vnet.Si
		Interface string `format:"%-30s"`
		Type      string `format:"%s" align:"center"`
		Namespace string `format:"%s" align:"center" width:"30"`
		NSID      string `format:"%s" align:"center"`
	}
	var ifs []nsIf
	for _, ns := range nm.namespace_by_name {
		for _, intf := range ns.interface_by_index {
			// Filter by namespace name.
			if len(matching_ns_names) > 0 {
				found := false
				for i := range matching_ns_names {
					if found = matching_ns_names[i].MatchString(ns.name); found {
						break
					}
				}
				if !found {
					continue
				}
			}

			x := nsIf{Namespace: ns.name, Interface: intf.name, Type: intf.kind.String(), si: vnet.SiNil}
			if intf.tuntap != nil {
				x.si = intf.tuntap.si
			}
			if ns.nsid != -1 {
				x.NSID = fmt.Sprintf("%d", ns.nsid)
			}
			ifs = append(ifs, x)
		}
	}
	sort.Slice(ifs, func(i, j int) bool {
		ni, nj := &ifs[i], &ifs[j]
		if ni.Namespace == nj.Namespace {
			if ni.si != vnet.SiNil && nj.si != vnet.SiNil {
				ifi, ifj := m.m.v.SwIf(ni.si), m.m.v.SwIf(nj.si)
				return m.m.v.SwLessThan(ifi, ifj)
			}
			return ni.Interface < nj.Interface
		}
		return ni.Namespace < nj.Namespace
	})
	colMap := map[string]bool{
		"si": false,
	}
	elib.Tabulate(ifs).WriteCols(w, colMap)
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

func (ns *net_namespace) add(m *net_namespace_main, dir, name string) (err error) {
	// Allocate unique index for namespace.
	ns.index = m.namespace_pool.GetIndex()
	m.namespace_pool.entries[ns.index] = ns

	defer func() {
		if err != nil {
			if ns.ns_fd > 0 {
				syscall.Close(ns.ns_fd)
			}
			m.namespace_pool.PutIndex(ns.index)
		}
	}()

	// Loop until namespace sockets are allocated.
	time_start := time.Now()
	var first_setns_errno syscall.Errno
	for {
		ns.m = m
		if ns.ns_fd, err = m.fd_for_path(dir, name); err != nil {
			return
		}
		// First setns may return EINVAL until "ip netns add X" performs mount bind; so we need to retry.
		err, first_setns_errno = elib.WithNamespace(ns.ns_fd, m.default_namespace.ns_fd, syscall.CLONE_NEWNET, ns.allocate_sockets)
		if err == nil {
			break
		}
		if time.Since(time_start) > 500*time.Millisecond {
			return
		}
		syscall.Close(ns.ns_fd)
		ns.ns_fd = -1
		if first_setns_errno == syscall.EINVAL {
			time.Sleep(1 * time.Millisecond)
		}
	}
	ns.nsid, ns.inode, err = m.nsid_for_fd(ns.ns_fd)
	if err != nil {
		return
	}
	// Check if namespace inode already exists.
	// This can happen when a link is made to an existing namespace.
	if ns1, ok := m.namespace_by_inode[ns.inode]; ok {
		err = fmt.Errorf("namespace add %s: already exists as %s", ns.name, ns1.name)
		return
	}

	if ns.nsid != netlink.DefaultNsid {
		m.namespace_by_nsid[ns.nsid] = ns
	}
	m.namespace_by_inode[ns.inode] = ns

	if err = ns.netlink_socket_pair.configure(ns.netlink_socket_fds[0], ns.netlink_socket_fds[1]); err != nil {
		syscall.Close(ns.ns_fd)
		ns.ns_fd = -1
		return
	}
	ns.listen(&m.m.netlink_main)
	ns.fibInit(false)
	intf := m.m.vnet_tun_main.create_tun(ns)
	if m.discovery_is_done() {
		if err = intf.init(m.m); err != nil {
			return
		}
	}
	return
}

func (ns *net_namespace) is_deleted() bool { return ns.ns_fd < 0 }

func (ns *net_namespace) del(m *net_namespace_main) {
	for index, intf := range ns.interface_by_index {
		if intf.si != vnet.SiNil {
			m.m.v.DelSwIf(intf.si)
		}
		if intf.tuntap != nil {
			intf.tuntap.close()
		}
		delete(ns.interface_by_index, index)
	}

	ns.m.namespace_pool.PutIndex(ns.index)
	ns.m.namespace_pool.entries[ns.index] = nil
	ns.fibInit(true)
	if ns.ns_fd > 0 {
		syscall.Close(ns.ns_fd)
		ns.ns_fd = -1
	}
	ns.netlink_socket_pair.close()
	ns.index = ^uint(0)
}
