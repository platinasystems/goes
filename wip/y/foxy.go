//+build foxy

package main

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/wip/y/internal/eeprom"

	"flag"
)

func (p *platform) boardInit() (err error) {
	d := eeprom.Device{
		BusIndex:   0,
		BusAddress: 0x51,
	}
	if e := d.GetInfo(); e != nil {
		p.Vnet.Logf("eeprom read failed: %s; using random address block", e)
		p.AddressBlock = ethernet.AddressBlock{
			Base:  ethernet.RandomAddress(),
			Count: 256,
		}
	} else {
		p.AddressBlock = ethernet.AddressBlock{
			Base:  d.Fields.BaseEthernetAddress,
			Count: uint32(d.Fields.NEthernetAddress),
		}
	}
	return
}

func (p *platform) boardPortInit(s fe1.Switch) (err error) {
	cf := fe1.SwitchConfig{}

	var (
		nPorts uint
		speed  float64
		pm     uint
	)

	flag.UintVar(&nPorts, "n-ports", 32, "Number of ports (32 64 128)")
	flag.Float64Var(&speed, "speed", 100e9, "Port speed in bits per second")
	flag.UintVar(&pm, "pm", 4, "PortMode(0-4x1, 1-2x1_1x2, 2-1x2_2x1, 3-2x2, 4-1x4)")
	flag.Parse()

	switch nPorts {
	case 32, 64, 96, 128:
		break
	default:
		panic("nports")
	}

	cf.Ports = make([]fe1.PortConfig, nPorts)

	switch nPorts {
	case 32:
		for i := range cf.Ports {
			cf.Ports[i] = fe1.PortConfig{
				PortBlockIndex:  uint(i),
				SpeedBitsPerSec: speed,
				LaneMask:        0xf,
				PhyInterface:    fe1.PhyInterfaceOptics,
			}
		}
	case 64:
		for i := range cf.Ports {
			cf.Ports[i] = fe1.PortConfig{
				PortBlockIndex:  uint(i) / 2,
				SpeedBitsPerSec: speed,
				LaneMask:        0x3 << (2 * (uint(i) % 2)),
				PhyInterface:    fe1.PhyInterfaceOptics,
			}
		}
	case 96:
		if pm == 2 {
			// 2x1_1x2
			for i := range cf.Ports {
				if i%3 == 0 {
					cf.Ports[i] = fe1.PortConfig{
						PortBlockIndex:  uint(i) / 3,
						SpeedBitsPerSec: speed,
						LaneMask:        0x3 << 2,
						PhyInterface:    fe1.PhyInterfaceOptics,
					}
				} else if (i+2)%3 == 0 {
					cf.Ports[i] = fe1.PortConfig{
						PortBlockIndex:  (uint(i) - 1) / 3,
						SpeedBitsPerSec: speed / 2,
						LaneMask:        0x1 << 1,
						PhyInterface:    fe1.PhyInterfaceOptics,
					}
				} else if (i+1)%3 == 0 {
					cf.Ports[i] = fe1.PortConfig{
						PortBlockIndex:  (uint(i) - 2) / 3,
						SpeedBitsPerSec: speed / 2,
						LaneMask:        0x1 << 0,
						PhyInterface:    fe1.PhyInterfaceOptics,
					}
				}
			}
		} else if pm == 1 {
			// 1x2_2x1
			for i := range cf.Ports {
				if i%3 == 0 {
					cf.Ports[i] = fe1.PortConfig{
						PortBlockIndex:  uint(i) / 3,
						SpeedBitsPerSec: speed / 2,
						LaneMask:        0x1 << 3,
						PhyInterface:    fe1.PhyInterfaceOptics,
					}
				} else if (i+2)%3 == 0 {
					cf.Ports[i] = fe1.PortConfig{
						PortBlockIndex:  (uint(i) - 1) / 3,
						SpeedBitsPerSec: speed / 2,
						LaneMask:        0x1 << 2,
						PhyInterface:    fe1.PhyInterfaceOptics,
					}
				} else if (i+1)%3 == 0 {
					cf.Ports[i] = fe1.PortConfig{
						PortBlockIndex:  (uint(i) - 2) / 3,
						SpeedBitsPerSec: speed,
						LaneMask:        0x3 << 0,
						PhyInterface:    fe1.PhyInterfaceOptics,
					}
				}
			}
		} else {
			panic("Unsupported 2/1/1 configuration")
		}
	case 128:
		for i := range cf.Ports {
			cf.Ports[i] = fe1.PortConfig{
				PortBlockIndex:  uint(i) / 4,
				SpeedBitsPerSec: speed,
				LaneMask:        0x1 << (uint(i) % 4),
				PhyInterface:    fe1.PhyInterfaceOptics,
			}
		}
	}

	// Management ports.
	for i := uint(0); i < 2; i++ {
		cf.Ports = append(cf.Ports, fe1.PortConfig{
			PortBlockIndex:  0,
			SubPortIndex:    i,
			IsManagement:    true,
			SpeedBitsPerSec: 10e9,
			LaneMask:        1 << (2 * i),
			PhyInterface:    fe1.PhyInterfaceKR,
			InitAutoneg:     true,
		})
	}

	{
		phys := [32 + 1]fe1.PhyConfig{}
		// No front panel remapping; no lane remapping on Foxy.
		for i := range phys {
			p := &phys[i]
			p.Index = uint8(i & 0x1f)
			p.FrontPanelIndex = p.Index ^ 1
			p.IsManagement = i == 32
		}
		cf.Phys = phys[:]
	}

	cf.Configure(p.Vnet, s)
	return
}
