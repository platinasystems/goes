package main

import (
	"github.com/platinasystems/fe1"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/internal/sriovs"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/bus/pci"
	"github.com/platinasystems/go/vnet/devices/ethernet/ixge"
	"github.com/platinasystems/go/vnet/ethernet"
	ipcli "github.com/platinasystems/go/vnet/ip/cli"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
	"github.com/platinasystems/go/vnet/pg"
	"github.com/platinasystems/go/vnet/unix"

	"fmt"
	"net"
	"os"
)

type platform struct {
	vnet.Package
	*fe1.Platform
	sriov_mode bool
}

func (p *platform) Init() (err error) {
	v := p.Vnet
	p.Platform = fe1.GetPlatform(v)

	if !p.sriov_mode {
		// 2 ixge ports are used to inject packets.
		ns := ixge.GetPortNames(v)
		p.SingleTaggedInjectNexts = make([]uint, len(ns))
		for i := range ns {
			p.SingleTaggedInjectNexts[i] = v.AddNamedNext(&p.Platform.SingleTaggedPuntInjectNodes.Inject, ns[i])
		}
	}

	if err = p.boardInit(); err != nil {
		return
	}
	for _, s := range p.Switches {
		if err = p.boardPortInit(s); err != nil {
			return
		}
	}
	return
}

func newSriovs(ver int) error {
	if ver > 0 {
		sriovs.VfName = func(port, subport uint) string {
			return fmt.Sprintf("eth-%d-%d", port+1, subport+1)
		}
	}
	eth0, err := net.InterfaceByName("eth0")
	if err != nil {
		return err
	}
	mac := sriovs.Mac(eth0.HardwareAddr)
	// skip over eth0, eth1, and eth2
	mac.Plus(3)
	sriovs.VfMac = mac.VfMac
	return sriovs.New(vfs)
}

func vlan_for_port(port, subport sriovs.Vf) (vf sriovs.Vf) {
	// physical port number for data ports are numbered starting at 1.
	// (phys 0 is cpu port...)
	phys := sriovs.Vf(1)

	// 4 sub-ports per port; mk1 ports are even/odd swapped.
	phys += 4 * (port ^ 1)

	phys += subport

	// Vlan is 1 plus physical port number.
	return sriovs.Vf(1 + phys)
}

// The vfs table is 0 based and is adjusted to 1 based beta and production
// units with VfName
var vfs = make_vfs()

func make_vfs() [][]sriovs.Vf {
	// pf0 = fe1 pipes 0 & 1; only 63 vfs supported so last sub port is not accessible.
	// pf1 = fe1 pipes 2 & 3; only 63 vfs supported so last sub port is not accessible.
	var pfs [2][63]sriovs.Vf
	for port := sriovs.Vf(0); port < 32; port++ {
		for subport := sriovs.Vf(0); subport < 4; subport++ {
			vf := port<<sriovs.PortShift | subport<<sriovs.SubPortShift | vlan_for_port(port, subport)
			pf := port / 16
			i := 4*(port%16) + subport
			if i < sriovs.Vf(len(pfs[pf])) {
				pfs[pf][i] = vf
			}
		}
	}
	return [][]sriovs.Vf{pfs[0][:], pfs[1][:]}
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
	p := &platform{}

	fns, err := sriovs.NumvfsFns()
	p.sriov_mode = err == nil && len(fns) > 0

	// Select packages we want to run with.
	fe1.Init(v)
	ethernet.Init(v)
	ip4.Init(v)
	ip6.Init(v)
	if !p.sriov_mode {
		ixge.Init(v, ixge.Config{DisableUnix: true, PuntNode: "fe1-single-tagged-punt"})
	} else if err = newSriovs(0); err != nil {
		return
	}
	pci.Init(v)
	pg.Init(v)
	ipcli.Init(v)
	unix.Init(v)

	v.AddPackage("platform", p)
	p.DependsOn("pci-discovery") // after pci discovery
	p.DependedOnBy("ip4")        // adjacencies/fib init needs to happen after fe1 init.
	p.DependedOnBy("ip6")

	err = v.Run(&in)
}
