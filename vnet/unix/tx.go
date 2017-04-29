// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/vnet"
)

type tx_node struct {
	vnet.OutputNode
}

func (n *tx_node) init(v *vnet.Vnet) {
	v.RegisterOutputNode(n, "punt")
}

func (n *tx_node) NodeOutput(out *vnet.RefIn) {
	n.Suspend(out)
	panic("tx")
}

func (ns *net_namespace) WriteAvailable() bool {
	return false
}

func (ns *net_namespace) WriteReady() (err error) {
	panic("tx")
	return
}

func (ns *net_namespace) ReadReady() (err error) {
	panic("not used")
	return
}
