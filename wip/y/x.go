package main

import (
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/internal/prog"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/bus/pci"
	"github.com/platinasystems/go/vnet/devices/ethernet/ixge"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/firmware"
	"github.com/platinasystems/go/vnet/ethernet"
	ipcli "github.com/platinasystems/go/vnet/ip/cli"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
	"github.com/platinasystems/go/vnet/pg"
	"github.com/platinasystems/go/vnet/unix"

	"fmt"
	"os"
)

type platform struct {
	vnet.Package
	*fe1.Platform
}

func (p *platform) Init() (err error) {
	v := p.Vnet
	p.Platform = fe1.GetPlatform(v)
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

func main() {
	var err error
	defer func() {
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	err = firmware.Extract(prog.Name())
	if err != nil {
		if e := firmware.Extract("fe1a.zip"); e != nil {
			return
		} else {
			err = nil
		}
	}

	var in parse.Input
	in.Add(os.Args[1:]...)

	v := &vnet.Vnet{}

	// Select packages we want to run with.
	fe1.Init(v)
	ethernet.Init(v)
	ip4.Init(v)
	ip6.Init(v)
	if false {
		ixge.Init(v)
	}
	pci.Init(v)
	pg.Init(v)
	ipcli.Init(v)
	unix.Init(v)

	p := &platform{}
	v.AddPackage("platform", p)
	p.DependsOn("pci-discovery") // after pci discovery
	p.DependedOnBy("ip4")        // adjacencies/fib init needs to happen after fe1 init.
	p.DependedOnBy("ip6")

	err = v.Run(&in)
}
