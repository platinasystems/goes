package main

import (
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/internal/eeprom"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/platforms/fe1"
	"github.com/platinasystems/go/vnet/platforms/mk1"

	"fmt"
	"os"
)

func platformConfigFromEEPROM(p *fe1.Platform) {
	d := eeprom.Device{
		BusIndex:   0,
		BusAddress: 0x51,
	}
	if e := d.GetInfo(); e != nil {
		fmt.Printf("eeprom read failed: %s; using random address block", e)
		p.BaseEthernetAddress = ethernet.RandomAddress()
		p.NEthernetAddress = 256
	} else {
		p.BaseEthernetAddress = ethernet.Address(d.Fields.BaseEthernetAddress)
		p.NEthernetAddress = d.Fields.NEthernetAddress
		p.Version = uint(d.Fields.DeviceVersion)
	}
}

func main() {
	var err error
	defer func() {
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	var in parse.Input
	in.Add(os.Args[1:]...)

	v := &vnet.Vnet{}
	p := &fe1.Platform{}

	platformConfigFromEEPROM(p)

	{
		var wip_in parse.Input
		cf := &p.PlatformConfig

		cf.EnableMsiInterrupt = true
		cf.DisableGpioSwitchReset = false
		cf.EnableCpuSwitchReset = false

		if in.Parse("wip %v", &wip_in) {
			for !wip_in.End() {
				switch {
				case wip_in.Parse("gpio-reset"):
					cf.DisableGpioSwitchReset = false
				case wip_in.Parse("no-gpio-reset"):
					cf.DisableGpioSwitchReset = true
				case wip_in.Parse("cpu-reset"):
					cf.EnableCpuSwitchReset = true
				case wip_in.Parse("no-cpu-reset"):
					cf.EnableCpuSwitchReset = false
				case wip_in.Parse("enable-msi"):
					cf.EnableMsiInterrupt = true
				case wip_in.Parse("disable-msi"):
					cf.EnableMsiInterrupt = false
				default:
					err = parse.ErrInput
					return
				}
			}
			// Make sure we reset switch either via gpio or cpu.
			// Its much safer to reset the switch via either method; so we enforce that here.
			if cf.DisableGpioSwitchReset && !cf.EnableCpuSwitchReset {
				cf.DisableGpioSwitchReset = false
			}
		}
	}

	if err = mk1.PlatformInit(v, p); err != nil {
		return
	}
	if err = v.Run(&in); err != nil {
		return
	}
	if err = mk1.PlatformExit(v, p); err != nil {
		return
	}
}
