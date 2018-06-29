// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"
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
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/platforms/fe1"
	"github.com/platinasystems/go/vnet/platforms/mk1"
	"github.com/platinasystems/go/vnet/unix"

	"gopkg.in/yaml.v2"
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

type mk1Main struct {
	fe1.Platform
}

func vnetdInit() {
	var err error
	// FIXME vnet shouldn't be so bursty
	const nports = 4 * 32
	const ncounters = 512
	xeth.EthtoolFlags = flags
	xeth.EthtoolStats = stats
	vnet.PortPrefixer = &mk1.PortPrefix
	vnet.Xeth, err = xeth.New(machine.Name,
		xeth.SizeofTxchOpt(nports*ncounters))

	if err != nil {
		panic(err)
	}
	eth1, err := net.InterfaceByName("eth1")
	if err != nil {
		panic(err)
	}
	eth2, err := net.InterfaceByName("eth2")
	if err != nil {
		panic(err)
	}

	vnet.PortIsCopper = func(ifname string) bool {
		if p, found := vnet.Ports[ifname]; found {
			return p.Flags.Test(CopperBit)
		}
		return false
	}
	vnet.PortIsFec74 = func(ifname string) bool {
		if p, found := vnet.Ports[ifname]; found {
			return p.Flags.Test(Fec74Bit)
		}
		return false
	}
	vnet.PortIsFec91 = func(ifname string) bool {
		if p, found := vnet.Ports[ifname]; found {
			return p.Flags.Test(Fec91Bit)
		}
		return false
	}
	p := new(mk1Main)
	vnet.Xeth.DumpIfinfo()
	err = vnet.Xeth.UntilBreak(func(buf []byte) error {
		ptr := unsafe.Pointer(&buf[0])
		switch xeth.KindOf(buf) {
		case xeth.XETH_MSG_KIND_ETHTOOL_FLAGS:
			msg := (*xeth.MsgEthtoolFlags)(ptr)
			ifname := xeth.Ifname(msg.Ifname)
			entry, found := vnet.Ports[ifname.String()]
			if found {
				entry.Flags = xeth.EthtoolFlagBits(msg.Flags)
			}
			if vnet.LogSvi {
				fmt.Printf("XETH_MSG_KIND_ETHTOOL_FLAGS: found:%v %+v\n",
					found, *msg)
			}
		case xeth.XETH_MSG_KIND_ETHTOOL_SETTINGS:
			msg := (*xeth.MsgEthtoolSettings)(ptr)
			ifname := xeth.Ifname(msg.Ifname)
			entry, found := vnet.Ports[ifname.String()]
			if found {
				entry.Speed = xeth.Mbps(msg.Speed)
			}
			if vnet.LogSvi {
				fmt.Printf("XETH_MSG_KIND_ETHTOOL_SETTINGS: found:%v %+v\n",
					found, *msg)
			}
		case xeth.XETH_MSG_KIND_IFINFO:
			var punt_index uint8
			msg := (*xeth.MsgIfinfo)(ptr)

			// convert eth1/eth2 to meth-0/meth-1
			switch msg.Iflinkindex {
			case int32(eth1.Index):
				punt_index = 0
			case int32(eth2.Index):
				punt_index = 1
			}

			ifname := xeth.Ifname(msg.Ifname)

			switch msg.Devtype {
			case xeth.XETH_DEVTYPE_PORT:
				err = unix.ProcessInterfaceInfo((*xeth.MsgIfinfo)(ptr), vnet.PreVnetd, nil, punt_index)
			case xeth.XETH_DEVTYPE_BRIDGE:
				be := vnet.SetBridge(msg.Id)
				be.Ifindex = msg.Ifindex
				be.Iflinkindex = msg.Iflinkindex
				be.PuntIndex = punt_index
				copy(be.Addr[:], msg.Addr[:])
				be.Net = msg.Net
			case xeth.XETH_DEVTYPE_UNTAGGED_BRIDGE_PORT:
				brm := vnet.SetBridgeMember(ifname.String())
				brm.Vid = msg.Id
				brm.IsTagged = false
				brm.PortVid = uint16(msg.Portid)
			case xeth.XETH_DEVTYPE_TAGGED_BRIDGE_PORT:
				brm := vnet.SetBridgeMember(ifname.String())
				brm.Vid = msg.Id // customer vlan
				brm.IsTagged = true
				brm.PortVid = uint16(msg.Portid)
			}
			if vnet.LogSvi {
				fmt.Printf("XETH_MSG_KIND_IFINFO: %+v\n", *msg)
			}
		case xeth.XETH_MSG_KIND_IFA:
			err = unix.ProcessInterfaceAddr((*xeth.MsgIfa)(ptr), vnet.PreVnetd, nil)
		}
		if err != nil {
			fmt.Println("Error processing xeth msg:", xeth.KindOf(buf), err)
		}
		return nil
	}, true)
	if err != nil {
		panic(err)
	}
	if vnet.LogSvi {
		for ifname, entry := range vnet.Ports {
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

func (p *mk1Main) parsePortConfig() (err error) {
	plat := &p.Platform
	if false { // /etc/goes/portprovision
		filename := "/etc/goes/portprovision"
		source, err := ioutil.ReadFile(filename)
		// If no file PortConfig will be left empty and lower layers will default
		if err == nil {
			err = yaml.Unmarshal(source, &plat.PortConfig)
			if err != nil {
				fmt.Println("yaml unmarshal failed", err)
				panic(err)
			}
			for _, p := range plat.PortConfig.Ports {
				fmt.Printf("Provision: %s speed %s lanes %d count %d\n", p.Name, p.Speed, p.Lanes, p.Count)
			}
		}
	} else { // ethtool
		// Massage ethtool port-provision format into fe1 format
		var pp fe1.PortProvision
		for ifname, entry := range vnet.Ports {
			pp.Name = ifname
			pp.Portindex = entry.Portindex
			pp.Subportindex = entry.Subportindex
			pp.Vid = ethernet.VlanTag(entry.Vid)
			pp.PuntIndex = entry.PuntIndex
			pp.Speed = fmt.Sprintf("%dg", entry.Speed/1000)
			// Need some more help here from ethtool to disambiguate
			// 40G 2-lane and 40G 4-lane
			// 20G 2-lane and 20G 1-lane
			// others?
			if vnet.LogSvi {
				fmt.Printf("From ethtool: name %v entry %+v pp %+v\n",
					ifname, entry, pp)
			}
			pp.Count = 1
			switch entry.Speed {
			case 100000, 40000:
				pp.Lanes = 4
			case 50000:
				pp.Lanes = 2
			case 25000, 20000, 10000, 1000:
				pp.Lanes = 1
			case 0: // need to calculate autoneg defaults
				if false {
					pp.Lanes = 1
				}
				pp.Lanes = p.getDefaultLanes(uint(pp.Portindex), uint(pp.Subportindex))
			}
			if false {
				fmt.Printf("PortConfig %s: %v\n", ifname, pp)
			}
			plat.PortConfig.Ports = append(plat.PortConfig.Ports, pp)
		}
	}
	return
}

func (p *mk1Main) parseBridgeConfig() (err error) {
	plat := &p.Platform

	if plat.BridgeConfig.Bridges == nil {
		plat.BridgeConfig.Bridges = make(map[ethernet.VlanTag]*fe1.BridgeProvision)
	}

	// for each bridge entry, create bridge config
	for vid, entry := range vnet.Bridges {
		bp, found := plat.BridgeConfig.Bridges[ethernet.VlanTag(vid)]
		if !found {
			bp = new(fe1.BridgeProvision)
			plat.BridgeConfig.Bridges[ethernet.VlanTag(vid)] = bp
		}
		bp.PuntIndex = entry.PuntIndex
		bp.Addr = entry.Addr
		if vnet.LogSvi {
			fmt.Printf("parse bridge %v\n", vid)
		}
	}

	// for each bridgemember entry, add to pbm or ubm of matching bridge config
	for ifname, entry := range vnet.BridgeMembers {
		bp, found := plat.BridgeConfig.Bridges[ethernet.VlanTag(entry.Vid)]
		if found {
			if entry.IsTagged {
				bp.TaggedPortVids =
					append(bp.TaggedPortVids,
						ethernet.VlanTag(entry.PortVid))
			} else {
				bp.UntaggedPortVids =
					append(bp.UntaggedPortVids,
						ethernet.VlanTag(entry.PortVid))
			}
			if vnet.LogSvi {
				fmt.Printf("bridgemember %v added to vlan %v\n",
					ifname,
					entry.Vid)
				fmt.Printf("bridgemember %+v\n", bp)
			}
		} else {
			if vnet.LogSvi {
				fmt.Printf("bridgemember %v ignored, vlan %v not found\n",
					ifname,
					entry.Vid)
			}
		}
	}
	return
}

func (p *mk1Main) parseFibConfig(v *vnet.Vnet) (err error) {
	// Process Interface addresses that have been learned from platina xeth driver
	// ip4IfaddrMsg(msg.Prefix, isDel)
	// Process Route data that have been learned from platina xeth driver
	// Since TH/Fp-ports are not initialized what could these be?
	//for _, fe := range vnet.FdbRoutes {
	//ip4IfaddrMsg(fe.Address, fe.Mask, isDel)
	//}
	return
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

	// Get initial port config from platina-mk1
	p.parsePortConfig()
	p.parseBridgeConfig()

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
		// this path only for tuntap, so eth- vs xeth names doesn't matter
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

func (p *mk1Main) getDefaultLanes(port, subport uint) (lanes uint) {
	lanes = 1

	// Two cases covered:
	// * 4-lane
	//         if first subport of port and only subport in set number of lanes should be 4
	// * 2-lane
	//         if first and third subports of port are present then number of lanes should be 2
	//         Unfortunately, 2-lane autoneg doesn't work for TH but leave this code here
	//         for possible future chipsets.
	//
	if p.Version == 0 { // alpha
		numSubPorts, subportList := subportsmatchingPort(port)
		if subport == 0 && numSubPorts == 1 {
			lanes = 4
		} else {
			if subport == 0 && numSubPorts == 2 && subportList.contains(2) {
				lanes = 2
			}
			if subport == 2 && numSubPorts == 2 && subportList.contains(0) {
				lanes = 2
			}
			lanes = 1 // override to have some function
		}
	} else { // beta/production
		numSubPorts, subportList := subportsmatchingPort(port)

		if subport == 1 && numSubPorts == 1 {
			lanes = 4
		} else {
			if subport == 1 && numSubPorts == 2 && subportList.contains(3) {
				lanes = 2
			}
			if subport == 3 && numSubPorts == 2 && subportList.contains(1) {
				lanes = 2
			}
			lanes = 1 // override to have some function
		}
	}
	return
}

type spList []uint

func subportsmatchingPort(targetport uint) (numsubports uint, subportlist spList) {
	var port, subport uint
	subportlist = []uint{0xf, 0xf, 0xf, 0xf}
	for ifname, _ := range vnet.Ports {
		fmt.Sscanf(ifname, "eth-%d-%d", &port, &subport)
		if port == targetport {
			subportlist[numsubports] = subport
			numsubports++
		}
	}
	return
}

func (subportlist spList) contains(targetsubport uint) bool {
	for _, subport := range subportlist {
		if subport == targetsubport {
			return true
		}
	}
	return false
}
