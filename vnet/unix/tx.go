// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/vnet"
)

type tx_node struct {
	rx_tx_node_common
	vnet.OutputNode
}

func (n *tx_node) add(m *net_namespace_main, ns *net_namespace) {
	n.rx_tx_node_common.add(m, ns, "tx")
	m.m.v.RegisterOutputNode(n, n.name)
}

func (n *tx_node) NodeOutput(out *vnet.RefIn) {
	panic("tx")
}

func (ns *net_namespace) WriteAvailable() bool {
	return false
}

func (ns *net_namespace) WriteReady() (err error) {
	panic("tx")
	return
}
