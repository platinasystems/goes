// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package m

import (
	"github.com/platinasystems/go/elib/hw/pci"
	"github.com/platinasystems/go/elib/hw/pcie"
	"github.com/platinasystems/go/vnet"

	"encoding/binary"
	"unsafe"
)

func (p *Platform) DeviceMatch(pd *pci.Device) (pci.DriverDevice, error) {
	s := p.switchByDeviceID[pd.Config.DeviceID.Device]
	c := s.GetSwitchCommon()

	c.PciDev = pd

	if c.PciFunction != pd.Addr.Fn {
		return nil, nil
	}

	haveIcpu := false
	if b, _, ok := pd.FindExtCap(pci.ExtVendorSpecific); ok {
		r := binary.LittleEndian.Uint32(b[8:])
		haveIcpu = r == 0x101 || r == 0x100
	}

	if pr := pcie.GetCapabilityHeader(pd); pr != nil {
		max := uint16(2)
		v := pr.Device.Control.Get(pd)
		v = (v &^ (7 << 5)) | (max << 5)   // max payload
		v = (v &^ (7 << 12)) | (max << 12) // max read request
		v &^= 1 << 4                       // disable relaxed ordering
		pr.Device.Control.Set(pd, v)
	}

	// PCI BAR 0
	r, err := pd.MapResource(&pd.Resources[0])
	if err != nil {
		panic(err)
	}
	var rr unsafe.Pointer
	if haveIcpu {
		rr = r

		// For icpu devices cpu controller is PCI BAR 2.
		r, err = pd.MapResource(&pd.Resources[1])
		if err != nil {
			panic(err)
		}
	}

	c.index = uint(len(p.Switches))
	p.Switches = append(p.Switches, s)

	c.CpuMain.Init(r, rr)

	return s, err
}

func RegisterDeviceIDs(v *vnet.Vnet, s Switch, devs []pci.VendorDeviceID) {
	c := s.GetSwitchCommon()
	c.Vnet = v
	p := GetPlatform(v)
	err := pci.SetDriver(p, pci.Broadcom, devs)
	if err != nil {
		panic(err)
	}
	if p.switchByDeviceID == nil {
		p.switchByDeviceID = make(map[pci.VendorDeviceID]Switch)
	}
	for i := range devs {
		p.switchByDeviceID[devs[i]] = s
	}
}
