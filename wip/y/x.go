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
