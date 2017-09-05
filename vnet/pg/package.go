// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pg

import (
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"
)

var packageIndex uint

type main struct {
	vnet.Package
	stream_type_map parse.StringMap
	stream_types    []StreamType
	nodes           []node
}

func Init(v *vnet.Vnet) {
	m := &main{}
	packageIndex = v.AddPackage("pg", m)
}

func (m *main) Configure(in *parse.Input) {
	n_nodes := 1
	for !in.End() {
		switch {
		case in.Parse("nodes %d", &n_nodes):
		default:
			panic(cli.ParseError)
		}
	}
	m.nodes = make([]node, n_nodes)
}

func (m *main) Init() (err error) {
	if len(m.nodes) == 0 {
		m.nodes = make([]node, 1)
	}
	for i := range m.nodes {
		m.nodes[i].init(m.Vnet, uint(i))
	}
	m.cli_init()
	return
}

func GetMain(v *vnet.Vnet) *main { return v.GetPackage(packageIndex).(*main) }
