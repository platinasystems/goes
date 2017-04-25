// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/vnet"
)

type rx_node struct {
	v  *vnet.Vnet
	ns *net_namespace
	vnet.InputNode
}

func (n *rx_node) add(ns *net_namespace, v *vnet.Vnet) {
	n.v = v
	n.ns = ns
	node_name := "unix-rx"
	if len(ns.name) > 0 {
		node_name += "-" + ns.name
	}
	v.RegisterInputNode(n, node_name)
}

func (n *rx_node) NodeInput(out *vnet.RefOut) {
	panic("rx")
}

func (ns *net_namespace) ReadReady() (err error) {
	return
}
