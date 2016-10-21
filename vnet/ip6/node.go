// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip6

import (
	"github.com/platinasystems/go/vnet"
)

func GetHeader(r *vnet.Ref) *Header { return (*Header)(r.Data()) }

type nodeMain struct {
	inputNode   inputNode
	rewriteNode vnet.Node
	arpNode     vnet.Node
}

func (m *Main) nodeInit(v *vnet.Vnet) {
	m.inputNode.Next = []string{
		input_next_drop: "error",
		input_next_punt: "punt",
	}
	v.RegisterInOutNode(&m.inputNode, "ip6-input")
}

const (
	input_next_drop = iota
	input_next_punt
)

type inputNode struct{ vnet.InOutNode }

func (node *inputNode) NodeInput(in *vnet.RefIn, out *vnet.RefOut) {
	node.Redirect(in, out, input_next_punt)
}
