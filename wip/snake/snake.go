package main

/*

Set SDK=xxx to where you installed bcmsdk via make install.

SDK=/home/edressel/wip/snake CGO_CFLAGS="-I${SDK}/bcmsdk/include" CGO_LDFLAGS="-L${SDK}/bcmsdk/lib" go build -o ~/gopath/src/github.com/platinasystems/machines/tbd_lc_dp/x -gcflags "-N -l" -ldflags "-extldflags -static -linkmode external" wip/snake

*/

/*
#cgo LDFLAGS: -lbcmsdk -lpthread -lm
#include <bcmsdk/bcmsdk.h>
#include <stdlib.h>
*/
import "C"

import (
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/layer"

	"flag"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"
	"unsafe"
)

type bcmError struct {
	error
	code      C.int // returned from bcm_* call
	tag       string
	func_name string
	file      string
	line      int
}

func (e *bcmError) Error() string {
	s := ""
	if e.func_name != "" {
		s += e.func_name + " "
	}
	s += fmt.Sprintf("%s:%d: %s %s", e.file, e.line, e.tag, C.GoString(C.bcm_strerror(e.code)))
	return s
}

func check(tag string, rv C.int) (err error) {
	if rv > 0 {
		rv = C.BCM_E_NONE
	}
	if rv != C.BCM_E_NONE {
		var pc uintptr
		be := &bcmError{code: rv}
		pc, be.file, be.line, _ = runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		if f != nil {
			be.func_name = f.Name()
		}
		be.tag = tag
		err = be
	}
	return
}

// Switch devices either manage left/right front-panel leaf-switch ports or are spine ports
type dev_role int

const (
	ROLE_LEAF dev_role = iota
	ROLE_SPINE
)

const (
	SPINE_TH = 0
)

type dev struct {
	role         dev_role
	index        uint8
	unit         C.int
	vendorID     uint16
	deviceID     uint16
	revisionID   uint8
	ports        []port
	counterReg   map[string]int
	counterDesc  map[string]string
	counterIndex map[string]int
	portConfig   C.bcm_port_config_t
	socInfo      *C.soc_info_t
}

type counterCode int

const (
	TBYT counterCode = iota
	TPKT
	Max
)

func (c counterCode) String() string {
	switch c {
	case TBYT:
		return "TBYT"
	case TPKT:
		return "TPKT"
	case Max:
		return ""
	default:
		return fmt.Sprintf("counterCode(%d)", c)
	}
}

type counter struct {
	code      counterCode
	reg       C.soc_reg_t
	name      string
	desc      string
	value     uint64
	lastValue uint64
	lastDiff  uint64
}

type counterByName []counter

func (a counterByName) Len() int           { return len(a) }
func (a counterByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a counterByName) Less(i, j int) bool { return a[i].name < a[j].name }

// Uniquely identifies device and port.  Used to form routing table.
type portUID uint64

type port struct {
	dev_index     int
	is_backplane  bool
	port_index    int
	uid           portUID
	egress_id     C.bcm_if_t
	bcm_port      C.bcm_port_t
	intf          C.bcm_if_t
	vrf           C.bcm_vrf_t
	vlanID        C.bcm_vlan_t
	name          string
	counters      []counter
	counterByCode []*counter
}

type snakeTest struct {
	cfg           config
	devs          []dev
	indexByUnit   []int
	unitByIndex   []int
	ports_by_uid  []*port
	next_port_uid []portUID
	eth           *ethernet.Header
	ip4           *ip4.Header
	payload       *layer.Incrementing
	packetBuffer  []byte
}

func (d *dev) isTomahawk() bool {
	if d.vendorID == C.BROADCOM_VENDOR_ID {
		switch uint16(d.deviceID) {
		case C.BCM56960_DEVICE_ID, C.BCM56961_DEVICE_ID, C.BCM56962_DEVICE_ID, C.BCM56963_DEVICE_ID:
			return true
		}
	}
	return false
}

// Call function foreach port in port bitmap.
func (d *dev) pbmp_foreach(b C.bcm_pbmp_t, f func(i C.bcm_port_t) (err error)) (err error) {
	for i := C.int(0); i < C._SHR_PBMP_PORT_MAX; i++ {
		i0 := uint(i / C._SHR_PBMP_WORD_WIDTH)
		i1 := uint(i % C._SHR_PBMP_WORD_WIDTH)
		if b.pbits[i0]&(1<<i1) != 0 {
			err = f(C.bcm_port_t(i))
			if err != nil {
				return
			}
		}
	}
	return
}

func (m *snakeTest) initDevs(nDevs int) (err error) {

	m.devs = make([]dev, nDevs)
	m.unitByIndex = make([]int, nDevs)
	m.indexByUnit = make([]int, nDevs)

	for i := range m.devs {
		d := &m.devs[i]

		d.unit = C.bcm_unit_for_dev(C.int(i))
		if m.cfg.isPlatinaChassis {
			switch d.unit {
			case 0:
				d.index = 2 // th2
			case 1:
				d.index = 1 // th1
			case 2:
				d.index = 0 // th0
			default:
				panic(d.unit)
			}
		} else {
			d.index = 0
		}
		m.indexByUnit[i] = int(d.index)
		m.unitByIndex[d.index] = i

		// Ugly fixme is it i == 0 1 or 2?
		// This has to be set at config time so we're forced
		// to hardcode a number here.
		d.role = ROLE_LEAF
		if i == SPINE_TH {
			d.role = ROLE_SPINE
		}

		if C.sysconf_attach(d.unit) < 0 {
			err = fmt.Errorf("sysconf_attach")
			return
		}

		x := uint32(C.soc_cm_pci_conf_read(d.unit, C.PCI_CONF_VENDOR_ID))
		d.vendorID = uint16(x & 0xffff)
		d.deviceID = uint16((x >> 16) & 0xffff)
		d.revisionID = uint8(C.soc_cm_pci_conf_read(d.unit, C.PCI_CONF_REVISION_ID)) & 0xff

		d.socInfo = C.bcm_soc_info_for_unit(d.unit)

		if m.cfg.verbose {
			fmt.Printf("%d: device 0x%x rev %d, bandwidth %.0f Gbits/sec\n",
				int(d.unit),
				d.deviceID, d.revisionID,
				float64(d.socInfo.bandwidth)*1e-3)
		}

		d.counterReg = make(map[string]int)
		d.counterDesc = make(map[string]string)
		for r := 0; r < C.NUM_SOC_REG; r++ {
			v := C.bcmsdk_counter_reg_name(d.unit, C.soc_reg_t(r))
			if v != nil {
				n := C.GoString(v)
				d.counterReg[n] = r
				w := C.bcmsdk_counter_reg_desc(d.unit, C.soc_reg_t(r))
				if w != nil {
					lines := strings.Split(C.GoString(w), "\n")
					d.counterDesc[n] = strings.TrimSpace(lines[0])
				}
			}
		}

		err = m.initStation(d, 4)
		if err != nil {
			return
		}
	}

	return
}

func (m *snakeTest) initPort(p *port) (err error) {
	d := &m.devs[p.dev_index]

	n := "xe"
	if d.isTomahawk() {
		n = "ce"
	}
	p.name = fmt.Sprintf("%s%d.%d", n, d.index, p.port_index)

	if m.cfg.verbose {
		fmt.Printf("Port init: %s\n", p.name)
	}

	p.counters = make([]counter, len(d.counterReg))
	i := 0
	d.counterIndex = make(map[string]int)
	for name, reg := range d.counterReg {
		p.counters[i] = counter{
			reg:  C.soc_reg_t(reg),
			name: name,
			desc: "",
			code: Max,
		}
		if v, ok := d.counterDesc[name]; ok {
			p.counters[i].desc = v
		}
		i++
	}
	sort.Sort(counterByName(p.counters))
	for i, c := range p.counters {
		d.counterIndex[c.name] = i
	}

	p.counterByCode = make([]*counter, Max)
	for c := counterCode(0); c < Max; c++ {
		s := c.String()
		if i, ok := d.counterIndex[s]; ok {
			p.counters[i].code = c
			p.counterByCode[c] = &p.counters[i]
		} else {
			err = fmt.Errorf("no counter with code %s", s)
			return
		}
	}

	/* Even leaf chassis ports are front panel ports; odd are backplane spine. */
	// FP ports >= 16 and BP ports < 16
	p.is_backplane = m.cfg.isPlatinaChassis && (d.role == ROLE_SPINE || (p.port_index < 16))

	/* Set port attributes */
	{
		var i C.bcm_port_info_t

		i.enable = 1
		i.action_mask |= C.BCM_PORT_ATTR_ENABLE_MASK

		i.linkscan = 1
		i.action_mask |= C.BCM_PORT_ATTR_LINKSCAN_MASK

		i.autoneg = 1
		i.action_mask |= C.BCM_PORT_ATTR_AUTONEG_MASK

		// Speed in Mbit/sec
		if d.isTomahawk() {
			i.speed = 100 * 1000
		} else {
			i.speed = 40 * 1000
		}
		i.action_mask |= C.BCM_PORT_ATTR_SPEED_MASK

		i._interface = C.BCM_PORT_IF_CR4
		if p.is_backplane {
			i._interface = C.BCM_PORT_IF_KR4
		}
		i.action_mask |= C.BCM_PORT_ATTR_INTERFACE_MASK

		switch m.cfg.loopback {
		case "phy":
			i.loopback = C.BCM_PORT_LOOPBACK_PHY
		case "mac":
			i.loopback = C.BCM_PORT_LOOPBACK_MAC
		case "none":
			i.loopback = C.BCM_PORT_LOOPBACK_NONE
		default:
			panic(fmt.Errorf("unkown loopback %s", m.cfg.loopback))
		}
		i.action_mask |= C.BCM_PORT_ATTR_LOOPBACK_MASK

		i.frame_max = C.int(m.cfg.frameSize)
		i.action_mask |= C.BCM_PORT_ATTR_FRAME_MAX_MASK

		i.untagged_vlan = p.vlanID
		i.action_mask |= C.BCM_PORT_ATTR_UNTAG_VLAN_MASK

		i.encap_mode = C.BCM_PORT_ENCAP_IEEE
		i.action_mask |= C.BCM_PORT_ATTR_ENCAP_MASK

		i.stp_state = C.BCM_STG_STP_FORWARD
		i.action_mask |= C.BCM_PORT_ATTR_STP_STATE_MASK

		err = check("port_selective_set", C.bcm_port_selective_set(d.unit, p.bcm_port, &i))
		if err != nil {
			return err
		}
	}

	/* Create L3 interface. */
	{
		var intf C.bcm_l3_intf_t

		/* Each port gets its own VRF. */
		p.vrf = C.bcm_vrf_t(1 + p.port_index) // must be > 0
		intf.l3a_vrf = p.vrf
		intf.l3a_mtu = C.int(m.cfg.frameSize - ethernet.HeaderBytes)

		dst := m.eth.Src
		dst.Add(uint64(p.uid))
		for i := 0; i < ethernet.AddressBytes; i++ {
			intf.l3a_mac_addr[i] = C.uint8(dst[i])
		}
		intf.l3a_vid = p.vlanID

		err = check("l3_intf_create", C.bcm_l3_intf_create(d.unit, &intf))
		if err != nil {
			return err
		}

		p.intf = intf.l3a_intf_id
	}

	/* Assign per-port VRF. */
	{
		err = check("vrf port_control_set", C.bcm_port_control_set(d.unit, p.bcm_port, C.bcmPortControlVrf, C.int(p.vrf)))
		if err != nil {
			return err
		}
	}

	{
		err = check("l3 egress mode switch_control_port_set", C.bcm_switch_control_port_set(d.unit, p.bcm_port, C.bcmSwitchL3EgressMode, C.int(1)))
		if err != nil {
			return err
		}
	}

	return
}

func (m *snakeTest) initL3Port(p *port) (err error) {
	d := &m.devs[p.dev_index]
	nextPort := m.ports_by_uid[m.next_port_uid[p.uid]]

	if m.cfg.verbose {
		fmt.Printf("Port L3 init: %s -> %s\n", p.name, nextPort.name)
	}

	/* Next hop. */
	if true {
		var e C.bcm_l3_egress_t

		e.intf = nextPort.intf
		e.port = nextPort.bcm_port
		e.vlan = nextPort.vlanID
		if true {
			e.flags = C.BCM_L3_KEEP_TTL
		}

		dst := m.eth.Dst
		dst.Add(uint64(nextPort.uid))
		for i := 0; i < ethernet.AddressBytes; i++ {
			e.mac_addr[i] = C.uint8(dst[i])
		}

		// Set these or else sdk fails
		e.dynamic_scaling_factor = C.BCM_L3_ECMP_DYNAMIC_SCALING_FACTOR_INVALID
		e.dynamic_load_weight = C.BCM_L3_ECMP_DYNAMIC_LOAD_WEIGHT_INVALID

		err = check("l3_egress_create", C.bcm_l3_egress_create(d.unit, e.flags, &e, &p.egress_id))
		if err != nil {
			return err
		}
	}

	/* Add route for host. */
	if true {
		var h C.bcm_l3_host_t

		h.l3a_vrf = p.vrf
		h.l3a_intf = p.egress_id
		h.l3a_ip_addr = C.bcm_ip_t(m.ip4.Dst.Uint32())
		if false {
			h.l3a_port_tgid = nextPort.bcm_port
			h.l3a_flags = C.BCM_L3_KEEP_TTL
			dst := m.eth.Dst
			dst.Add(uint64(nextPort.uid))
			for i := 0; i < ethernet.AddressBytes; i++ {
				h.l3a_nexthop_mac[i] = C.uint8(dst[i])
			}
		}

		err = check("l3_host_add", C.bcm_l3_host_add(d.unit, &h))
		if err != nil {
			return err
		}
	} else {
		var r C.bcm_l3_route_t

		r.l3a_vrf = p.vrf
		r.l3a_intf = p.egress_id
		r.l3a_subnet = C.bcm_ip_t(m.ip4.Dst.Uint32())
		r.l3a_ip_mask = C.bcm_ip_t(0xffffffff)
		r.l3a_flags = C.BCM_L3_KEEP_TTL
		err = check("l3_route_add", C.bcm_l3_route_add(d.unit, &r))
		if err != nil {
			return err
		}
	}

	return
}

func (m *snakeTest) initL2Port(p *port) (err error) {
	d := &m.devs[p.dev_index]
	nextPort := m.ports_by_uid[m.next_port_uid[p.uid]]

	if m.cfg.verbose {
		fmt.Printf("Port L2 init: %s -> %s\n", p.name, nextPort.name)
	}

	// Make static L2 entry {port vlan, dst mac} -> next port in routing table. */
	{
		var x C.bcm_l2_addr_t

		x.flags = C.BCM_L2_STATIC
		x.vid = p.vlanID
		for i := 0; i < ethernet.AddressBytes; i++ {
			x.mac[i] = C.uint8(m.eth.Dst[i])
		}
		x.port = C.int(nextPort.bcm_port)
		err = check("l2_addr_add", C.bcm_l2_addr_add(d.unit, &x))
		if err != nil {
			return err
		}
	}

	return
}

/* Set up MY_STATION table to recognize packets as L3 packets. */
func (m *snakeTest) initStation(d *dev, n_match int) (err error) {
	var s C.bcm_l2_station_t
	var station_id C.int

	if !m.cfg.isLayer3 {
		return
	}

	/* Match given number of bytes of ethernet destination. */
	for i := 0; i < n_match; i++ {
		s.dst_mac[i] = C.uint8(m.eth.Dst[i])
		s.dst_mac_mask[i] = 0xff
	}

	s.flags = C.BCM_L2_STATION_IPV4

	err = check("l2_station_add", C.bcm_l2_station_add(d.unit, &station_id, &s))
	return
}

type bcmConfig struct {
	name, value string
}

func sdkConfig(cs []bcmConfig) (err error) {
	for _, c := range cs {
		ca := C.CString(c.name)
		cb := C.CString(c.value)
		defer func() {
			C.free(unsafe.Pointer(ca))
			C.free(unsafe.Pointer(cb))
		}()
		if r := int(C.sal_config_set(ca, cb)); r < 0 {
			err = fmt.Errorf("sal_config_set returns %d", r)
		}
	}
	return
}

func sdkInit(cs []bcmConfig) (nDevs int, err error) {
	if C.sal_core_init() < 0 {
		err = fmt.Errorf("sal_core_init")
		return
	}

	if C.sal_appl_init() < 0 {
		err = fmt.Errorf("sal_appl_init")
		return
	}

	err = sdkConfig(cs)
	if err != nil {
		return
	}

	if C.sysconf_init() < 0 {
		err = fmt.Errorf("sysconf_init")
		return
	}
	if C.sysconf_probe() < 0 {
		err = fmt.Errorf("sysconf_probe")
		return
	}
	if C.soc_ndev == 0 {
		err = fmt.Errorf("no switching devices")
		return
	}

	nDevs = int(C.soc_ndev)
	return
}

func sdkColdBoot(unit C.int) (err error) {
	err = check("soc_reset_init", C.soc_reset_init(unit))
	if err != nil {
		return
	}

	err = check("soc_misc_init", C.soc_misc_init(unit))
	if err != nil {
		return
	}

	err = check("soc_mmu_init", C.soc_mmu_init(unit))
	if err != nil {
		return
	}

	err = check("bcm_init", C.bcm_init(unit))
	if err != nil {
		return
	}

	err = check("l3_enable_set", C.bcm_l3_enable_set(unit, C.int(1)))
	if err != nil {
		return
	}

	err = check("bcm_l3_init", C.bcm_l3_init(unit))
	if err != nil {
		return
	}

	usec := 10000
	err = check("bcm_linkscan_enable_set", C.bcm_linkscan_enable_set(unit, C.int(usec)))
	if err != nil {
		return
	}

	return
}

// Send a packet out first port on first device.
// Packet will infinitly loop through all switch ports
func (m *snakeTest) tx(d *dev, nPacket int) (err error) {
	for i := 0; i < nPacket; i++ {
		err = check("tx_packet", C.bcmsdk_tx_packet(d.unit, d.ports[0].bcm_port,
			(*C.uint8)(unsafe.Pointer(&m.packetBuffer[0])),
			C.int(len(m.packetBuffer))))
		if err != nil {
			return
		}
	}
	return
}

func (m *snakeTest) printLinkStates() (nUp int) {
	nUp = 0
	for di := range m.devs {
		d := &m.devs[di]
		for pi := range d.ports {
			p := &d.ports[pi]
			var rv, linkstatus C.int
			rv = C.bcm_port_link_status_get(d.unit, p.bcm_port, &linkstatus)
			fmt.Printf("%s %d %d [%d]\n", p.name, p.bcm_port, linkstatus, rv)
			if linkstatus == 1 {
				nUp += 1
			}
		}
	}
	return
}

func run(c config) (err error) {
	m := snakeTest{
		cfg: c,
		eth: &ethernet.Header{
			Type: ethernet.IP4,
			Src:  ethernet.Address{0xe0, 0xe1, 0xe2, 0xe3, 0xe4, 0xe5},
			Dst:  ethernet.Address{0xea, 0xeb, 0xec, 0xed, 0xee, 0xef},
		},
		ip4: &ip4.Header{
			Protocol: ip.UDP,
			Src:      ip4.Address{0x1, 0x2, 0x3, 0x4},
			Dst:      ip4.Address{0x5, 0x6, 0x7, 0x8},
			Tos:      0,
			Ttl:      255,
			Ip_version_and_header_length: 0x45,
			Fragment_id:                  0x1234,
			Flags_and_fragment_offset:    ip4.DONT_FRAGMENT,
		},
		payload: &layer.Incrementing{Count: c.frameSize - ethernet.HeaderBytes - ip4.HeaderBytes},
	}
	m.packetBuffer = layer.Make(m.eth, m.ip4, m.payload)

	// Configure SDK
	{
		// These maps are for backplane cables and fp cables wired port <-> port+1
		// Should we restore old lane-maps for individual TH loopbacks ?
		rxLaneMap := [3][32]int{
			// unit 0 => th 2 spine backplane traces 23 & 24
			{
				0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123,
				0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123,
				0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x3210,
				0x3210, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123,
			},
			// unit 1 => th 1 leaf right side ports
			{
				0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210,
				0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210,
				0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210,
				0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210,
			},
			// unit 2 => th 0 leaf left side ports
			{
				0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210,
				0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210,
				0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210,
				0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210,
			},
		}
		txLaneMap := [3][32]int{
			// unit 0 => th 2 spine
			{
				0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123,
				0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123,
				0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x3210,
				0x3210, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123, 0x0123,
			},
			// unit 1 => th 1 leaf right side ports ports 0-15 backplane ports 16-31 front panel
			{
				0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210,
				0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210,
				0x2301, 0x2301, 0x2301, 0x2301, 0x2301, 0x2301, 0x2301, 0x2301,
				0x2301, 0x2301, 0x2301, 0x2301, 0x2301, 0x2301, 0x2301, 0x2301,
			},
			// unit 2 => th 0 leaf left side ports
			{
				0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210,
				0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210, 0x3210,
				0x2301, 0x2301, 0x2301, 0x2301, 0x2301, 0x2301, 0x2301, 0x2301,
				0x2301, 0x2301, 0x2301, 0x2301, 0x2301, 0x2301, 0x2301, 0x2301,
			},
		}
		configs := []bcmConfig{
			{name: "pbmp_xport_xe.BCM56850", value: "0x1fffffffe"},
			{name: "pbmp_oversubscribe.BCM56850", value: "0x1fffffffe"},
			{name: "pbmp_xport_xe.BCM56960", value: "0x0000000000000000000000000000003fffffffdffffffff7fffffffdfffffffe"},
			{name: "pbmp_oversubscribe.BCM56960", value: "0x0000000000000000000000000000000000003fc000000ff0000003fc000001fe"},
			{name: "bcm_linkscan_interval", value: "250000"},
		}

		for ui := 0; ui < 3; ui++ {
			for pi := 0; pi < 32; pi++ {
				var pm int
				switch {
				case pi < 8*1:
					pm = 1 + (pi - 0)
				case pi < 8*2:
					pm = 34 + (pi - 8)
				case pi < 8*3:
					pm = 68 + (pi - 16)
				default:
					pm = 102 + (pi - 24)
				}
				polarity_flip := 0x0
				if ui == SPINE_TH && pi != 23 && pi != 24 {
					polarity_flip = 0xf
				}
				is_revb_kludge_port := false
				switch pi {
				case 5, 7, 8, 10, 20, 22, 25, 27:
					is_revb_kludge_port = true
				}
				if is_revb_kludge_port {
					polarity_flip ^= 0xf
				}
				configs = append(configs,
					bcmConfig{
						name:  fmt.Sprintf("portmap_%d.BCM56850", 1+pi),
						value: fmt.Sprintf("%d:40:OSG.0xf", 1+4*pi),
					},
					bcmConfig{
						name:  fmt.Sprintf("portmap_%d.BCM56960", pm),
						value: fmt.Sprintf("%d:100", 1+4*pi),
					},
					bcmConfig{
						name:  fmt.Sprintf("xgxs_rx_lane_map_core0_ce%d.%d", pi, ui),
						value: fmt.Sprintf("0x%x", rxLaneMap[ui][pi]),
					},
					bcmConfig{
						name:  fmt.Sprintf("xgxs_tx_lane_map_core0_ce%d.%d", pi, ui),
						value: fmt.Sprintf("0x%x", txLaneMap[ui][pi]),
					},
					bcmConfig{
						name:  fmt.Sprintf("phy_xaui_rx_polarity_flip_ce%d.%d", pi, ui),
						value: fmt.Sprintf("0x%x", polarity_flip),
					},
					bcmConfig{
						name:  fmt.Sprintf("phy_xaui_tx_polarity_flip_ce%d.%d", pi, ui),
						value: fmt.Sprintf("0x%x", polarity_flip),
					})
			}
		}

		var nDevs int
		nDevs, err = sdkInit(configs)
		if err != nil {
			return err
		}
		err = m.initDevs(nDevs)
		if err != nil {
			return err
		}
	}

	// Cold boot chips.
	for _, d := range m.devs {
		if c.verbose {
			fmt.Printf("Cold boot unit %d\n", d.unit)
		}
		err = sdkColdBoot(d.unit)
		if err != nil {
			return
		}
	}

	// Initialize ports.
	nPortUID := portUID(0)
	for dev_index := range m.devs {
		d := &m.devs[dev_index]

		err = check("port_config_get", C.bcm_port_config_get(d.unit, &d.portConfig))
		if err != nil {
			return
		}

		bm := d.portConfig.ce
		n := int(C.bcm_pbmp_count(bm))
		d.ports = make([]port, n)

		// Create vlan for each port; all ports are members of each vlan
		for pi := 0; pi < n; pi++ {
			vid := C.bcm_vlan_t(2 + pi)
			err = check("bcm_vlan_create", C.bcm_vlan_create(d.unit, vid))
			if err != nil {
				return
			}
			err = check("vlan_port_add", C.bcm_vlan_port_add(d.unit, vid, d.portConfig.all, d.portConfig.all))
			if err != nil {
				return
			}
			d.ports[pi].vlanID = vid
		}

		// Set up each port and assign a unique id which uniquely identifies a device port pair.
		pi := 0
		err = d.pbmp_foreach(bm, func(bcm_port C.bcm_port_t) (suberr error) {
			p := &d.ports[pi]
			p.dev_index = dev_index
			p.port_index = pi
			p.uid = nPortUID
			p.bcm_port = bcm_port
			suberr = m.initPort(p)
			if suberr != nil {
				return
			}
			nPortUID++
			pi++
			return
		})
		if err != nil {
			return
		}
	}

	// Map unique ID to port
	m.ports_by_uid = make([]*port, int(nPortUID))
	for di := range m.devs {
		for pi := range m.devs[di].ports {
			p := m.devs[di].ports[pi]
			m.ports_by_uid[p.uid] = &p
		}
	}

	// Compute routing table
	m.next_port_uid = make([]portUID, len(m.ports_by_uid))
	if m.cfg.isPlatinaChassis {
		m.platina1RURoute()
	} else {
		for uid, p := range m.ports_by_uid {
			d := &m.devs[p.dev_index]
			ni := (p.port_index + 1) % len(d.ports)
			n := d.ports[ni]
			m.next_port_uid[uid] = n.uid
		}
	}

	// Setup l2/l3 routing
	for _, p := range m.ports_by_uid {
		if m.cfg.isLayer3 {
			err = m.initL3Port(p)
		} else {
			err = m.initL2Port(p)
		}
		if err != nil {
			return
		}
	}

	// Debug - print link state
	for {
		fmt.Printf("Linkstates check\n")
		n_up := m.printLinkStates()
		if n_up == len(m.ports_by_uid) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Start counters
	for _, d := range m.devs {
		err = check("soc_counter_start", C.soc_counter_start(d.unit, C.SOC_COUNTER_F_DMA,
			/* 5 secs in usec */ C.int(5e6),
			/* all ports */ d.portConfig.all))
		if err != nil {
			return
		}
	}

	if m.cfg.isPlatinaChassis {
		// For chassis all devices are connected so we only send packets on the left leaf port 0.
		for i := 0; i < 2; i++ {
			d := &m.devs[m.unitByIndex[i]]
			err = m.tx(d, m.cfg.nFrames)
			if err != nil {
				return
			}
		}
	} else {
		// Otherwise send packets out the first port on all devices.
		for i := range m.devs {
			err = m.tx(&m.devs[i], m.cfg.nFrames)
			if err != nil {
				return
			}
		}
	}

	// Run test
	tStart := time.Now()
	tEnd := tStart.Add(m.cfg.TestDuration)
	tLast := tStart
	for {
		tNow := time.Now()
		if tNow.After(tEnd) {
			break
		}

		time.Sleep(m.cfg.StatsDuration)

		for _, d := range m.devs {
			err = check("soc_counter_sync", C.soc_counter_sync(d.unit))
			if err != nil {
				return
			}
		}

		tNow = time.Now()
		fmt.Printf("Stat changes %v\n", tNow)
		for _, d := range m.devs {
			txPackets := uint64(0)
			txPacketsDiff := uint64(0)
			for _, p := range d.ports {
				printedPort := false
				for ci, c := range p.counters {
					var val C.uint64
					err = check("soc_counter_get", C.soc_counter_get(d.unit, C.soc_port_t(p.bcm_port), c.reg, C.int(-1), &val))
					if err != nil {
						return
					}
					c.value = uint64(val)
					if c.value != c.lastValue {
						portName := ""
						if !printedPort {
							portName = p.name
							printedPort = true
						}
						c.lastDiff = c.value - c.lastValue
						if m.cfg.verbose {
							fmt.Printf("%-8s%-8s%-50s%16d%+16d\n", portName, c.name, c.desc, val, c.lastDiff)
						}
						c.lastValue = c.value
						p.counters[ci] = c
					}
				}
				tpkt := p.counterByCode[TPKT]
				txPackets += tpkt.value
				txPacketsDiff += tpkt.lastDiff
			}

			{
				dev := float64(m.devs[0].socInfo.io_bandwidth) * 1e6 // socInfo.io_bandwidth in M bits/sec
				/* Include 8 byte ethernet preamble 12 byte inter frame gap and 4 byte CRC. */
				wireFrameSize := float64(8 + 12 + 4 + m.cfg.frameSize)
				cur := float64(txPackets) * wireFrameSize * 8 / tNow.Sub(tStart).Seconds()
				diff := float64(txPacketsDiff) * wireFrameSize * 8 / tNow.Sub(tLast).Seconds()
				fmt.Printf("TH%d Wire bandwidth ave %.2fG, current interval %.2fG, max %.2fG, fraction %.2f\n",
					d.index, 1e-9*cur, 1e-9*diff, 1e-9*dev, cur/dev)
			}

			// Measure temperature.
			{
				var ts [64]C.bcm_switch_temperature_monitor_t
				var nt C.int
				C.bcm_switch_temperature_monitor_get(d.unit, C.int(len(ts)), &ts[0], &nt)
				fmt.Printf("Die temperatures current/peak:\n")
				for i := 0; i < int(nt); i++ {
					t := &ts[i]
					fmt.Printf("%8.1fC %8.1fC\n", float64(t.curr)/10, float64(t.peak)/10)
				}
			}

			tLast = tNow
		}
	}

	fmt.Println("Done\n")

	return
}

type config struct {
	verbose          bool
	isPlatinaChassis bool
	isLayer3         bool
	loopback         string
	nFrames          int
	frameSize        int
	// public so %+v prints as 10s instead of in nsec
	TestDuration  time.Duration
	StatsDuration time.Duration
}

func main() {
	var c config

	flag.BoolVar(&c.isPlatinaChassis, "platina", false, "Snake Test for Platina chassis")
	flag.BoolVar(&c.isLayer3, "layer3", false, "Snake Test with Layer3")
	flag.DurationVar(&c.TestDuration, "time", 10*time.Second, "Snake test duration")
	flag.DurationVar(&c.StatsDuration, "print", 5*time.Second, "Statistics print interval")
	flag.IntVar(&c.nFrames, "n-frames", 10, "Number of frames to send")
	flag.IntVar(&c.frameSize, "frame-size", 9416, "Ethernet frame size (0 means use max)")
	flag.BoolVar(&c.verbose, "verbose", false, "Verbose output")
	flag.StringVar(&c.loopback, "loopback", "phy", "Put ports into loopback (none, phy, mac)")
	flag.Parse()

	if c.verbose {
		fmt.Printf("Config: %+v\n", c)
	}
	e := run(c)
	if e != nil {
		panic(e)
	}
}
