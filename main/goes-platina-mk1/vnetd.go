// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"unsafe"

	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/goes/cmd/ip"
	"github.com/platinasystems/go/goes/cmd/vnetd"
	"github.com/platinasystems/go/internal/machine"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/xeth"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/platforms/fe1"
	"github.com/platinasystems/go/vnet/platforms/mk1"
)

var vnetdCounterSeparators *strings.Replacer

var vnetdLinkStatTranslation = map[string]string{
	"port-rx-multicast-packets":     "multicast",
	"port-rx-bytes":                 "rx-bytes",
	"port-rx-crc_error-packets":     "rx-crc-errors",
	"port-rx-runt-packets":          "rx-fifo-errors",
	"port-rx-undersize-packets":     "rx-length-errors",
	"port-rx-oversize-packets":      "rx-over-errors",
	"port-rx-packets":               "rx-packets",
	"port-tx-total-collisions":      "collisions",
	"port-tx-fifo-underrun-packets": "tx-aborted-errors",
	"port-tx-bytes":                 "tx-bytes",
	"port-tx-runt-packets":          "tx-fifo-errors",
	"port-tx-packets":               "tx-packets",
}

type PortProvisionEntry struct {
	Flags xeth.EthtoolFlagBits
	Speed xeth.Mbps
}

type PortProvision map[string]*PortProvisionEntry

type mk1Main struct {
	fe1.Platform
	PortProvision
}

func vnetdInit() {
	var err error
	// FIXME vnet shouldn't be so bursty
	const nports = 4 * 32
	const ncounters = 512
	xeth.EthtoolFlags = flags
	xeth.EthtoolStats = stats
	vnet.Xeth, err = xeth.New(machine.Name,
		xeth.SizeofTxchOpt(nports*ncounters))
	if err != nil {
		panic(err)
	}
	p := new(mk1Main)
	p.PortProvision = make(PortProvision)
	vnet.Xeth.EthtoolDump()
	vnet.Xeth.UntilBreak(func(buf []byte) error {
		ptr := unsafe.Pointer(&buf[0])
		hdr := (*xeth.Hdr)(ptr)
		if !hdr.IsHdr() {
			return fmt.Errorf("invalid xeth msg: %#x", buf)
		}
		switch xeth.Op(hdr.Op) {
		case xeth.XETH_ETHTOOL_FLAGS_OP:
			msg := (*xeth.EthtoolFlagsMsg)(ptr)
			p.PortProvisionEntry(msg.Ifname.String()).Flags =
				msg.Flags
		case xeth.XETH_ETHTOOL_SETTINGS_OP:
			msg := (*xeth.EthtoolSettingsMsg)(ptr)
			p.PortProvisionEntry(msg.Ifname.String()).Speed =
				msg.Settings.Speed
		default:
			return fmt.Errorf("invalid op: %d", hdr.Op)
		}
		return nil
	})
	if true {
		for ifname, entry := range p.PortProvision {
			fmt.Print(ifname, ".flags: ", entry.Flags, "\n")
			fmt.Print(ifname, ".speed: ", entry.Speed, "\n")
		}
	}
	vnetd.Hook = p.vnetdHook
	vnetd.CloseHook = p.stopHook
	vnetd.Counter = func(s string) string {
		s = vnetdCounterSeparators.Replace(s)
		if x, found := vnetdLinkStatTranslation[s]; found {
			s = x
		}
		return s
	}
	vnetdCounterSeparators =
		strings.NewReplacer(" ", "-", ".", "-", "_", "-")
}

func (p *mk1Main) vnetdHook(init func(), v *vnet.Vnet) error {
	p.Init = init

	s, err := redis.Hget(machine.Name, "eeprom.DeviceVersion")
	if err != nil {
		return err
	}
	if _, err = fmt.Sscan(s, &p.Version); err != nil {
		return err
	}
	s, err = redis.Hget(machine.Name, "eeprom.NEthernetAddress")
	if err != nil {
		return err
	}
	if _, err = fmt.Sscan(s, &p.NEthernetAddress); err != nil {
		return err
	}
	s, err = redis.Hget(machine.Name, "eeprom.BaseEthernetAddress")
	if err != nil {
		return err
	}
	input := new(parse.Input)
	input.SetString(s)
	p.BaseEthernetAddress.Parse(input)

	fi, err := os.Stat("/sys/bus/pci/drivers/ixgbe")
	p.KernelIxgbe = err == nil && fi.IsDir()

	vnetd.UnixInterfacesOnly = !p.KernelIxgbe

	// Default to using MSI versus INTX for switch chip.
	p.EnableMsiInterrupt = true

	if err = mk1.PlatformInit(v, &p.Platform); err != nil {
		return err
	}

	return nil
}

func (p *mk1Main) stopHook(i *vnetd.Info, v *vnet.Vnet) error {
	if vnet.Xeth != nil {
		vnet.Xeth.Close()
	}
	if p.KernelIxgbe {
		return mk1.PlatformExit(v, &p.Platform)
	} else {
		// FIXME why isn't this done in mk1.PlatformExit?
		interfaces, err := net.Interfaces()
		if err != nil {
			return err
		}
		for _, dev := range interfaces {
			if strings.HasPrefix(dev.Name, "eth-") ||
				dev.Name == "vnet" {
				args := []string{"link", "delete", dev.Name}
				err = ip.Goes.Main(args...)
				if err != nil {
					fmt.Println("write err", err)
					return err
				}
			}
		}
		return nil
	}
}

func (p PortProvision) PortProvisionEntry(ifname string) *PortProvisionEntry {
	entry, found := p[ifname]
	if !found {
		entry = new(PortProvisionEntry)
		p[ifname] = entry
	}
	return entry
}

func (p PortProvision) PortProvisionMbps(ifname string) uint32 {
	return uint32(p.PortProvisionEntry(ifname).Speed)
}

func (p PortProvision) PortProvisionCopper(ifname string) bool {
	return p.PortProvisionEntry(ifname).Flags.Test(CopperBit)
}

func (p PortProvision) PortProvisionFec74(ifname string) bool {
	return p.PortProvisionEntry(ifname).Flags.Test(Fec74Bit)
}

func (p PortProvision) PortProvisionFec91(ifname string) bool {
	return p.PortProvisionEntry(ifname).Flags.Test(Fec91Bit)
}
