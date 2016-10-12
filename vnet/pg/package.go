package pg

import (
	"github.com/platinasystems/go/vnet"
)

var packageIndex uint

type main struct {
	vnet.Package
	node
}

func Init(v *vnet.Vnet) {
	m := &main{}
	packageIndex = v.AddPackage("pg", m)
}

func (m *main) Init() (err error) {
	m.node.init(m.Vnet)
	m.cli_init()
	return
}

func GetMain(v *vnet.Vnet) *main { return v.GetPackage(packageIndex).(*main) }
