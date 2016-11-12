// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/cpu"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/packet"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/phy"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/port"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"
	"github.com/platinasystems/go/vnet/ethernet"

	"fmt"
	"sync"
	"time"
)

const (
	n_port_block_100g = 32
	n_port_block_10g  = 1
)

type Port struct {
	packet.InterfaceNode
	ethernet.Interface
	m.PortCommon
	physical_port_number
}

type loopbackType int

const (
	loopbackNone loopbackType = iota
	loopbackPhy
	loopbackExtCable
)

func (p *Port) GetPortName() string { return p.HwIf.Name() }

type portMain struct {
	port_blocks_100g [n_port_block_100g]port.PortBlock
	port_blocks_10g  [n_port_block_10g]port.PortBlock
	phys_100g        [n_port_block_100g]phy.HundredGig
	phys_10g         [n_port_block_10g]phy.FortyGig
	loopbackPorts    [n_pipe]Port
	cpuPort          Port

	packet_dma_ports [n_pipe_ports]packet.Port

	// Port pointers indexed by physical ports.  Only applies to data ports and management ports.
	port_by_phys_port [n_phys_ports]*Port

	// Data ports.
	ports_100g [n_port_block_100g][4]Port
	// Management ports.
	ports_10g [2]Port

	rx_pipe_port_table [n_phys_ports]rx_port_table_entry
	tx_pipe_port_table [n_phys_ports]tx_port_table_entry
}

func revb_kludge(c m.PhyConfig) (v m.PhyConfig) {
	var tmp [2][4]bool
	v = c
	if v.Rx_invert_lane_polarity == nil {
		v.Rx_invert_lane_polarity = tmp[0][:]
	}
	if v.Tx_invert_lane_polarity == nil {
		v.Tx_invert_lane_polarity = tmp[1][:]
	}
	v.Rx_logical_lane_by_phys_lane = make([]uint8, 4)
	v.Tx_logical_lane_by_phys_lane = make([]uint8, 4)
	for i := range v.Rx_logical_lane_by_phys_lane {
		v.Rx_invert_lane_polarity[i] = !v.Rx_invert_lane_polarity[i]
		v.Tx_invert_lane_polarity[i] = !v.Tx_invert_lane_polarity[i]
		v.Rx_logical_lane_by_phys_lane[i] = c.Rx_logical_lane_by_phys_lane[3-i]
		v.Tx_logical_lane_by_phys_lane[i] = c.Tx_logical_lane_by_phys_lane[3-i]
	}
	return
}

func (p *Port) IsUnix() (ok bool) {
	ok = true
	n := p.physical_port_number
	switch {
	case n == phys_port_cpu:
		ok = false
	case n >= phys_port_loopback_pipe_0 && n < phys_port_loopback_pipe_0+n_pipe:
		ok = false
	}
	return
}
func (p *Port) ValidateSpeed(speed vnet.Bandwidth) (err error) {
	pc := p.GetPortCommon()
	new := float64(speed)
	was_auto := pc.Autoneg
	is_auto := new == 0
	phy := p.GetPhy()

	// cpu/loopback ports have no phy.
	if phy == nil {
		pc.SpeedBitsPerSec = new
		return
	}

	// See if PHY is ok with given speed.
	isHiGig := false
	if !is_auto && !phy.ValidateSpeed(p, new, isHiGig) {
		err = fmt.Errorf("invalid speed")
		return
	}

	pc.SpeedBitsPerSec = new
	pc.Autoneg = is_auto

	// Don't actually realize speed until port is enabled.
	if !pc.Enabled {
		return
	}

	if is_auto {
		phy.SetAutoneg(p, pc.Autoneg)
	} else {
		if was_auto {
			phy.SetAutoneg(p, false)
		}
		phy.SetSpeed(p, pc.SpeedBitsPerSec, isHiGig)
	}
	return
}
func (p *Port) GetHwInterfaceCounterNames() vnet.InterfaceCounterNames {
	return p.Switch.(*fe1a).InterfaceCounterNames
}

func (p *Port) GetHwInterfaceCounterValues(th *vnet.InterfaceThread) {
	p.Switch.(*fe1a).get_port_counters(p, th)
}

func (p *Port) registerEthernet(v *vnet.Vnet, name string, provision bool, isUnixInterface bool) {
	var ok bool
	config := &ethernet.InterfaceConfig{}
	if isUnixInterface {
		if config.Address, ok = m.GetPlatform(v).AddressBlock.Alloc(); !ok {
			panic("ran out of platform ethernet addresses")
		}
	}
	config.Unprovisioned = !provision
	ethernet.RegisterInterface(v, p, config, name)
}

func (t *fe1a) addPort(p *Port, pb *port.PortBlock) {
	c := t.GetSwitchCommon()
	pb.Ports[p.SubPortIndex].Port = p
	c.Ports = append(c.Ports, p)

	v := t.Vnet
	var name string
	if p.IsManagement {
		u := 0
		if p.LaneMask != 1<<0 {
			u = 1
		}
		name = fmt.Sprintf("meth-%d", u)
	} else {
		name = fmt.Sprintf("eth-%d-%d", p.FrontPanelIndex, p.SubPortIndex)
	}

	p.registerEthernet(v, name, m.IsProvisioned(p), true)
	p.HwIf.SetSpeed(vnet.Bandwidth(p.SpeedBitsPerSec))
	hi := p.HwIf.Hi()
	cf := packet.InterfaceNodeConfig{
		Use_module_header: true,
		Dst_port:          uint8(p.physical_port_number.toPipe()),
	}
	p.InterfaceNode.Init(&t.CpuMain.PacketDma, hi, &cf)
	v.RegisterOutputInterfaceNode(&p.InterfaceNode, hi, name)

	t.linkScanAdd(p, pb)

	t.packet_dma_ports[p.physical_port_number.toPipe()] = packet.Port{
		Name: p.GetPortName(),
		Si:   p.HwIf.Si(),
	}
}

func (t *fe1a) addCpuLoopbackPort(p *Port, name string, pn physical_port_number, speed vnet.Bandwidth) {
	v := t.Vnet
	p.Switch = t
	p.physical_port_number = pn
	p.registerEthernet(v, name, true, false)
	p.HwIf.SetSpeed(speed)
	hi := p.HwIf.Hi()
	in := &p.InterfaceNode
	cf := packet.InterfaceNodeConfig{
		Use_module_header:   false,
		Dst_port:            uint8(pn.toPipe()),
		Use_loopback_header: false,
		// cpu port is on pipe 0
		Loopback_port:        uint8(phys_port_loopback_pipe_0),
		Is_visibility_packet: true,
	}
	in.Init(&t.CpuMain.PacketDma, hi, &cf)
	v.RegisterOutputInterfaceNode(in, hi, name)

	t.port_by_phys_port[p.physical_port_number] = p

	t.packet_dma_ports[pn.toPipe()] = packet.Port{
		Name: p.GetPortName(),
		Si:   p.HwIf.Si(),
	}
}

func (t *fe1a) addCpuLoopbackPorts() {
	t.addCpuLoopbackPort(&t.cpuPort, "fe1-cpu", phys_port_cpu, 10e9)
	for pipe := uint(0); pipe < n_pipe; pipe++ {
		t.addCpuLoopbackPort(&t.loopbackPorts[pipe],
			fmt.Sprintf("fe1-pipe%d-loopback", pipe),
			phys_port_loopback_for_pipe(pipe),
			10e9)
	}
}

func (t *fe1a) PortInit(vn *vnet.Vnet) {
	port.Init(vn)
	phy.Init(vn)
	t.cliInit()

	c := t.GetSwitchCommon()
	cf := &c.SwitchConfig

	t.set_mmu_pipe_map()

	t.l3_main.fe1a = t
	t.register_hooks(vn)

	for p := range t.port_blocks_100g {
		t.port_blocks_100g[p] = port.PortBlock{
			Switch:          t,
			Phy:             &t.phys_100g[p],
			PortBlockIndex:  m.PortBlockIndex(p),
			FrontPanelIndex: ^uint8(0),
			SbusBlock:       BlockClport0 + sbus.Block(p),
			IsXlPort:        false,
			AllLanesMask:    m.LaneMask(0xf),
		}
		c.PortBlocks = append(c.PortBlocks, &t.port_blocks_100g[p])
		t.phys_100g[p] = phy.HundredGig{
			Common: phy.Common{
				Switch:    t,
				PortBlock: &t.port_blocks_100g[p],
			},
		}
		c.Phys = append(c.Phys, &t.phys_100g[p])
	}

	for p := range t.port_blocks_10g {
		t.port_blocks_10g[p] = port.PortBlock{
			Switch:          t,
			Phy:             &t.phys_10g[p],
			PortBlockIndex:  m.PortBlockIndex(0),
			FrontPanelIndex: ^uint8(0),
			SbusBlock:       BlockXlport0,
			IsXlPort:        true,
			AllLanesMask:    m.LaneMask(1 << uint(2*p)),
		}
		c.PortBlocks = append(c.PortBlocks, &t.port_blocks_10g[p])
		t.phys_10g[p] = phy.FortyGig{
			Common: phy.Common{
				Switch:    t,
				PortBlock: &t.port_blocks_10g[p],
			},
		}
		c.Phys = append(c.Phys, &t.phys_10g[p])
	}

	for p := range cf.Phys {
		pc := cf.Phys[p]

		if pc.Rx_logical_lane_by_phys_lane == nil {
			pc.Rx_logical_lane_by_phys_lane = []uint8{0, 1, 2, 3}
		}
		if pc.Tx_logical_lane_by_phys_lane == nil {
			pc.Tx_logical_lane_by_phys_lane = []uint8{0, 1, 2, 3}
		}

		// Parts after Revision A do lane swapping & polarity reversal to work around internal signal issues.
		if t.revision_id.getId() != revision_id_a0 && !pc.IsManagement {
			switch pc.Index {
			case 5, 7, 8, 10, 20, 22, 25, 27:
				pc = revb_kludge(pc)
			}
		}

		if pc.IsManagement {
			t.phys_10g[pc.Index].PhyConfig = pc
		} else {
			t.phys_100g[pc.Index].PhyConfig = pc
			t.port_blocks_100g[pc.Index].FrontPanelIndex = pc.FrontPanelIndex
		}
	}

	// Data ports.
	for pbi := 0; pbi < n_port_block_100g; pbi++ {
		for i := 0; i < 4; i++ {
			p := &t.ports_100g[pbi][i]
			pb := &t.port_blocks_100g[pbi]
			phy := &t.phys_100g[pbi]

			p.Switch = t
			p.Phy = phy
			p.PortBlock = pb

			p.FrontPanelIndex = uint(pb.FrontPanelIndex)
			p.SubPortIndex = m.SubPortIndex(i)
			p.physical_port_number = phys_port_data_lo +
				physical_port_number(uint(pb.PortBlockIndex)*4+uint(p.SubPortIndex))
			t.port_by_phys_port[p.physical_port_number] = p

			p.PhyInterface = m.PhyInterfaceOptics

			// Enable sub index 0 as 100G 4 lanes.
			if i == 0 {
				p.SpeedBitsPerSec = 100e9
				p.LaneMask = 0xf
			}

			t.addPort(p, pb)
		}
	}

	// Management ports.
	t.ports_10g[0].physical_port_number = phys_port_management_0
	t.ports_10g[1].physical_port_number = phys_port_management_1
	for i := range t.ports_10g {
		p := &t.ports_10g[i]
		pb := &t.port_blocks_10g[0]
		phy := &t.phys_10g[0]

		p.Switch = t
		p.Phy = phy
		p.PortBlock = pb
		t.port_by_phys_port[p.physical_port_number] = p

		p.FrontPanelIndex = uint(pb.FrontPanelIndex)
		p.SubPortIndex = m.SubPortIndex(2 * i)

		p.SpeedBitsPerSec = 10e9
		p.LaneMask = 1 << p.SubPortIndex
		p.PhyInterface = m.PhyInterfaceOptics
		p.IsManagement = true

		t.addPort(p, pb)
	}

	t.addCpuLoopbackPorts()

	for pi := range cf.Ports {
		pc := &cf.Ports[pi]
		var p *Port
		if pc.IsManagement {
			p = &t.ports_10g[pc.SubPortIndex]
		} else {
			p = &t.ports_100g[pc.PortBlockIndex][pc.SubPortIndex]
		}
		p.PhyInterface = pc.PhyInterface
		p.SpeedBitsPerSec = pc.SpeedBitsPerSec
		// fixme derive this
		p.LaneMask = pc.LaneMask
	}

	t.set_port_bitmaps()

	// Initialize misc/mmu stuff
	t.misc_init()
	t.shared_lookup_sram_init()
	t.enablePorts(true)
	t.init_port_table()
	t.garbage_dump_init()

	// Link/admin is always up for cpu/loopback ports.
	t.cpuPort.SetLinkUp(true)
	t.cpuPort.SetAdminUp(true)
	for pipe := uint(0); pipe < n_pipe; pipe++ {
		t.loopbackPorts[pipe].SetLinkUp(true)
		t.loopbackPorts[pipe].SetAdminUp(true)
	}

	{
		nm := "fe1-rx"
		tn := t.Name()
		if len(tn) > 0 {
			nm += "-" + tn
		}
		t.CpuMain.PacketDma.StartRx(vn, nm, t.packet_dma_ports[:])
		c.CpuMain.StartPacketDma()
	}

	t.CpuMain.LinkScanEnable(vn, true)
}

type port_bitmap_main struct {
	all_ports        port_bitmap
	cpu_ports        port_bitmap
	loopback_ports   port_bitmap
	management_ports port_bitmap
}

func (t *fe1a) set_port_bitmaps() {
	bm := &t.port_bitmap_main
	bm.cpu_ports.add(phys_port_cpu.toPipe())
	for pipe := uint(0); pipe < n_pipe; pipe++ {
		bm.loopback_ports.add(phys_port_loopback_for_pipe(pipe).toPipe())
	}
	bm.management_ports.add(phys_port_management_0.toPipe())
	bm.management_ports.add(phys_port_management_1.toPipe())

	for phys := phys_port_data_lo; phys <= phys_port_data_hi; phys++ {
		p := t.port_by_phys_port[phys]
		if m.IsProvisioned(p) {
			bm.all_ports.add(phys.toPipe())
		}
	}
	bm.all_ports.or(&bm.management_ports)
	bm.all_ports.add(phys_port_cpu.toPipe())
}

// For link scan: data ports 0-127; management ports 128, 129
func (p physical_port_number) toLinkScan() uint16 {
	switch {
	case p >= phys_port_data_lo && p <= phys_port_data_hi:
		return uint16(p - phys_port_data_lo)
	case p == phys_port_management_0:
		return 128
	case p == phys_port_management_1:
		return 130
	default:
		panic("bad port")
	}
}

func (p *physical_port_number) fromLinkScan(i uint16) (ok bool) {
	ok = true
	switch {
	case i < n_data_ports:
		*p = phys_port_data_lo + physical_port_number(i)
	case i == 128:
		*p = phys_port_management_0
	case i == 130:
		*p = phys_port_management_1
	default:
		ok = false
	}
	return
}

func (t *fe1a) linkScanAdd(p *Port, pb *port.PortBlock) {
	var ls cpu.LinkScanPort
	id, bus := t.PhyIDForPort(pb.SbusBlock, p.SubPortIndex)
	ls.Index = p.physical_port_number.toLinkScan()
	ls.PhyId, ls.PhyBusId = uint8(id), uint8(bus)
	ls.Enable = true
	ls.IsExternal = false
	ls.IsClause45 = true
	t.CpuMain.LinkScanAdd(&ls)
}

const (
	ledDataRamLinkStatus = 1 << 0
	ledDataRamTurbo      = 1 << 7
)

func (t *fe1a) setLedPortState(p physical_port_number, bit uint8, isSet bool) {
	var li uint
	pi := uint(p - phys_port_data_lo)
	switch {
	case p.is_data_port_in_range(0, 32) || p.is_data_port_in_range(96, 32):
		li = 0
	case p.is_data_port_in_range(32, 64):
		li = 1
	case p == phys_port_management_0:
		li, pi = 2, 0
	case p == phys_port_management_1:
		li, pi = 2, 1
	default:
		panic(fmt.Errorf("physical port number out of range %d", p))
	}
	t.CpuMain.Leds[li].SetPortState(pi, bit, isSet)
}

func (t *fe1a) LinkStatusChange(v *cpu.LinkStatus) {
	n := uint16(0)
	for i := range v {
		for j := uint(0); j < cpu.LinkStatusWordBits; j++ {
			var p physical_port_number
			if ok := p.fromLinkScan(n); ok {
				linkUp := v[i]&(1<<j) != 0
				t.port_by_phys_port[p].SetLinkUp(linkUp)
				t.setLedPortState(p, ledDataRamLinkStatus, linkUp)
			}
			n++
		}
	}
}

func (t *fe1a) enablePortBlock(pb *port.PortBlock, enable bool) {
	// Power up phys and port blocks; run phy init sequence.
	pb.Enable(enable)

	// Bring up provisioned ports (ports with non-zero lane masks).
	if pb.IsXlPort {
		for i := range t.ports_10g {
			p := &t.ports_10g[i]
			if p.LaneMask != 0 {
				pb.SetPortEnable(p, enable)
			}
		}
	} else {
		for i := 0; i < 4; i++ {
			p := &t.ports_100g[pb.PortBlockIndex][i]
			if p.LaneMask != 0 {
				pb.SetPortEnable(p, enable)
				if SnakeLoopMode == loopbackPhy {
					pb.SetPortLoopback(p, enable)
				}
			}
		}
	}
}

func (t *fe1a) enablePorts(enable bool) {
	var wg sync.WaitGroup

	start := time.Now()

	for pbi := 0; pbi < n_port_block_100g+n_port_block_10g; pbi++ {
		var pb *port.PortBlock
		if pbi < len(t.port_blocks_100g) {
			pb = &t.port_blocks_100g[pbi]
		} else {
			pb = &t.port_blocks_10g[pbi-n_port_block_100g]
		}
		wg.Add(1)
		go func() {
			t.enablePortBlock(pb, enable)
			wg.Add(-1)
		}()
	}

	wg.Wait()

	if elog.Enabled() {
		elog.GenEventf("enablePorts %e secs", time.Since(start).Seconds())
	}
}

// Set default VLAN to 1.
func (p *Port) DefaultId() vnet.IfIndex { return 1 }

func (a *Port) LessThan(b vnet.HwInterfacer) bool {
	return a.PortCommon.LessThan(&b.(*Port).PortCommon)
}

func (p *Port) SetLoopback(v vnet.IfLoopbackType) (err error) {
	var enable bool
	switch v {
	case vnet.IfLoopbackMac:
		enable = true
	case vnet.IfLoopbackNone:
		enable = false
	default:
		err = vnet.ErrNotSupported
		return
	}
	p.PortBlock.SetPortLoopback(p, enable)
	return
}

func (p physical_port_number) toGpp() global_physical_port_number { return p.toPipe().toGpp() }
func (p pipe_port_number) toGpp() global_physical_port_number     { return global_physical_port_number(p) }

func (t *fe1a) port_mapping_init() {
	q := t.getDmaReq()

	// IDB to pipe port number mapping.
	for i := rx_pipe_mmu_port_number(0); i < n_rx_pipe_port; i++ {
		for pipe := uint(0); pipe < n_pipe; pipe++ {
			phys := i.toPhys(pipe)
			if phys != phys_port_invalid {
				pipePort := phys.toPipe()
				t.rx_pipe_mems.rx_pipe_to_pipe_port_number_mapping_table[i].seta(q, sbus.Unique(pipe), uint32(pipePort))
			}
		}
	}
	q.Do()

	for phys := physical_port_number(0); phys < n_phys_ports; phys++ {
		pipePort := phys.toPipe()
		if pipePort == pipe_port_invalid {
			continue
		}

		gpp := pipePort.toGpp()

		// GPP to pipe port number mapping.
		t.rx_pipe_mems.device_port_by_global_physical[gpp].set(q, uint32(pipePort))

		// Tx_pipe device port to physical.
		t.tx_pipe_controller.device_to_physical_port_number_mapping[pipePort].seta(q, sbus.AddressSplit, uint32(phys))

		mmu_port := phys.toGlobalMmu(t)
		if mmu_port != mmu_global_port_number_invalid {
			t.mmu_global_controller.device_port_by_mmu_port[mmu_port].set(q, uint32(pipePort))
			t.mmu_global_controller.physical_port_by_mmu_port[mmu_port].set(q, uint32(phys))
			t.mmu_global_controller.global_physical_port_by_mmu_port[mmu_port].set(q, uint32(gpp))
		}
	}
	q.Do()
}
