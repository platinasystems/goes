// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/vnet"
)

type tx_node struct {
	v  *vnet.Vnet
	ns *net_namespace
	vnet.OutputNode
}

func (n *tx_node) add(ns *net_namespace, v *vnet.Vnet) {
	n.v = v
	n.ns = ns
	node_name := "unix-tx"
	if len(ns.name) > 0 {
		node_name += "-" + ns.name
	}
	v.RegisterOutputNode(n, node_name)
}

func (n *tx_node) NodeOutput(out *vnet.RefIn) {
	panic("tx")
}

func (ns *net_namespace) WriteReady() (err error) {
	return
}

func (ns *net_namespace) WriteAvailable() bool {
	return false
}
