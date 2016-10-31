// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package port

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/sbus"

	"fmt"
)

type Port struct {
	Port m.Porter
}

// A block of ports which may contain 1, 2, 3 or 4 ports.
// Each port block has a single PHY which has 4 SERDES lanes implementing the ports.
type PortBlock struct {
	m.Switch

	Phy m.Phyer

	PortBlockIndex m.PortBlockIndex
	SbusBlock      sbus.Block

	// Index on switch front panel.  Used only to derive name.
	FrontPanelIndex uint8

	IsXlPort bool

	// mask of all possible lanes for this port block.
	AllLanesMask m.LaneMask

	Ports [4]Port

	req dmaRequest
}

func (p *PortBlock) Name() string {
	s := "ce"
	if p.IsXlPort {
		s = "xe"
	}
	return fmt.Sprintf("%s%02d", s, p.FrontPanelIndex)
}

func (p *PortBlock) GetPhy() m.Phyer                     { return p.Phy }
func (p *PortBlock) GetPipe() uint                       { return uint(p.PortBlockIndex / 8) }
func (p *PortBlock) GetPortBlockIndex() m.PortBlockIndex { return p.PortBlockIndex }
func (p *PortBlock) GetPortIndex(port m.Porter) uint {
	return uint(4*p.PortBlockIndex) + m.GetSubPortIndex(port)
}

// Mode as in mode register.
type PortBlockMode uint8

const (
	PortBlockMode4x1     PortBlockMode = iota // 1 port: 1 x 4 lane port
	PortBlockMode2x1_1x2                      // 3 ports: 2 x 1 lane port + 1 x 2 lane port
	PortBlockMode1x2_2x1                      // 3 ports: 1 x 2 lane port + 2 x 1 lane port
	PortBlockMode2x2                          // 2 ports: 2 x 2 lane port
	PortBlockMode1x4                          // 4 ports: 4 x 1 lane port
)

func (p *PortBlock) GetMode() PortBlockMode {
	var m [4]m.LaneMask
	for i := range m {
		port := p.Ports[i].Port
		if port != nil {
			m[i] = port.GetPortCommon().LaneMask
		}
	}

	// Management ports.
	if p.IsXlPort {
		return PortBlockMode4x1
	}

	switch {
	case m[0] == 0xf && m[1] == 0 && m[2] == 0 && m[3] == 0:
		return PortBlockMode1x4
	case m[0] == 3<<0 && m[1] == 0 && m[2] == 3<<2 && m[3] == 0:
		return PortBlockMode2x2
	case m[0] == 3<<0 && m[1] == 0 && m[2] == 1<<2 && m[3] == 1<<3:
		return PortBlockMode2x1_1x2
	case m[0] == 1<<0 && m[1] == 1<<1 && m[2] == 3<<2 && m[3] == 0:
		return PortBlockMode1x2_2x1
	case m[0] == 1<<0 && m[1] == 1<<1 && m[2] == 1<<2 && m[3] == 1<<3:
		return PortBlockMode4x1
	default:
		panic("port mode")
	}
	return 0
}

type dmaRequest struct {
	sbus.DmaRequest
	portBlock *PortBlock
}

func (q *dmaRequest) Do() {
	s := q.portBlock.Switch.GetSwitchCommon()
	s.Dma.Do(&q.DmaRequest)
}

func (p *PortBlock) dmaReq() *dmaRequest {
	p.req.portBlock = p
	return &p.req
}

func (p *PortBlock) Enable(enable bool) {
	r, _, _, _ := p.get_regs()
	q := p.dmaReq()

	var v [2]uint32

	r.mac_control.get(q, &v[0])
	r.tsc_control.get(q, &v[1])
	q.Do()

	const (
		mac_reset_h    uint32 = 1 << 0
		tsc_reset_l    uint32 = 1 << 0
		tsc_power_down uint32 = 1 << 3
	)

	if enable {
		v[0] &^= mac_reset_h
		v[1] |= tsc_reset_l
		v[1] &^= tsc_power_down
	} else {
		v[0] |= mac_reset_h
		v[1] &^= tsc_reset_l
		v[1] |= tsc_power_down
	}

	r.mac_control.set(q, v[0])
	r.tsc_control.set(q, v[1])
	// Reset counters or garbage will be returned
	r.reset_mib_counters.set(q, 0xf)
	r.reset_mib_counters.set(q, 0)

	if enable {
		// Set counters to saturate on overflow; clear on read
		r.counter_mode.set(q, 0)

		// Set block mode.
		mode := p.GetMode()
		r.mode.set(q, uint32(mode)<<3|uint32(mode)<<0)
	}

	q.Do()

	if enable {
		p.Phy.Init()
	}
}

func (p *PortBlock) SetPortEnable(port m.Porter, enable bool) {
	block_regs, mac_regs, _, _ := p.get_regs()
	q := p.dmaReq()
	i := m.GetSubPortIndex(port)
	pc := port.GetPortCommon()
	sp := pc.SpeedBitsPerSec

	p.Phy.SetSpeed(port, sp, false)

	pc.Enabled = enable
	pc.Autoneg = true
	p.Phy.SetAutoneg(port, pc.Autoneg)

	const (
		mac_sw_reset uint64 = 1 << 6
		rx_en        uint64 = 1 << 1
		tx_en        uint64 = 1 << 0
	)
	var (
		v [2]uint64
		w uint32
	)
	block_regs.port_enable.get(q, &w)
	mac_regs.control[i].get(q, &v[0])
	mac_regs.tx_control[i].get(q, &v[1])
	q.Do()

	if enable {
		v[0] |= tx_en | rx_en
		v[0] &^= mac_sw_reset
		w |= 1 << i
	} else {
		v[0] &^= tx_en | rx_en
		v[0] |= mac_sw_reset
		w &^= 1 << i
	}

	mac_regs.control[i].set(q, v[0])
	block_regs.port_enable.set(q, w)

	if enable {
		// We never want to drop packets in MAC on rx.
		mtu := 1<<14 - 1
		mac_regs.rx.max_bytes_per_packet[i].set(q, uint64(mtu))
		block_regs.mib_stats_max_packet_size[i].set(q, uint32(mtu))

		// Default crc mode is 2 replace; set to 0 append.
		v[1] = (v[1] &^ (3 << 0)) | (0 << 0)
		mac_regs.tx_control[i].set(q, v[1])
	}

	q.Do()
}

func (p *PortBlock) SetPortLoopback(port m.Porter, enable bool) {
	_, r, _, _ := p.get_regs()
	q := p.dmaReq()
	i := m.GetSubPortIndex(port)

	const local_loopback_enable = 1 << 2
	v := r.control[i].getDo(q)
	if enable {
		v |= local_loopback_enable
	} else {
		v &^= local_loopback_enable
	}
	r.control[i].set(q, v)
	q.Do()
}

type portStatus struct {
	name          string
	link_up       bool
	signal_detect bool
	pmd_lock      bool
}

func (p *PortBlock) getStatus(port m.Porter) (s portStatus) {
	r, _, _, _ := p.get_regs()
	i := m.GetSubPortIndex(port)
	q := p.dmaReq()
	v := r.tsc_lane_status[i].getDo(q)
	s.name = port.GetPortName()
	s.pmd_lock = v&(1<<2) != 0
	s.signal_detect = v&(1<<1) != 0
	s.link_up = v&(1<<0) != 0
	return s
}

type switchSelect struct{ m.SwitchSelect }

func (ss *switchSelect) showBcmPortStatus(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	var ifs vnet.HwIfChooser
	ifs.Init(ss.Vnet)
	for !in.End() {
		switch {
		case in.Parse("%v", &ifs):
		default:
			err = cli.ParseError
			return
		}
	}
	stats := []portStatus{}
	ifs.Foreach(func(v *vnet.Vnet, r vnet.HwInterfacer) {
		var (
			p  m.Porter
			ok bool
		)
		if p, ok = r.(m.Porter); !ok {
			return
		}
		pb := p.GetPortBlock()
		if pb != nil { // ignore cpu, loopback ports
			stats = append(stats, pb.(*PortBlock).getStatus(p))
		}
	})
	elib.TabulateWrite(w, stats)
	return
}

func Init(v *vnet.Vnet) {
	ss := &switchSelect{}
	ss.Vnet = v
	v.CliAdd(&cli.Command{
		Name:   "show bcm port-status mac",
		Action: ss.showBcmPortStatus,
	})
}
