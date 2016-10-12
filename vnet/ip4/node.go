package ip4

import (
	"github.com/platinasystems/go/vnet"
)

func GetHeader(r *vnet.Ref) *Header { return (*Header)(r.Data()) }

type nodeMain struct {
	inputNode              inputNode
	inputValidChecksumNode inputValidChecksumNode
	rewriteNode            vnet.Node
	arpNode                vnet.Node
}

func (m *Main) nodeInit(v *vnet.Vnet) {
	m.inputNode.Next = []string{
		input_next_drop: "error",
		input_next_punt: "punt",
	}
	v.RegisterInOutNode(&m.inputNode, "ip4-input")
	m.inputValidChecksumNode.Next = m.inputNode.Next
	v.RegisterInOutNode(&m.inputValidChecksumNode, "ip4-input-valid-checksum")
}

const (
	input_next_drop = iota
	input_next_punt
)

type inputNode struct{ vnet.InOutNode }

func (node *inputNode) NodeInput(in *vnet.RefIn, out *vnet.RefOut) {
	node.Redirect(in, out, input_next_punt)
}

type inputValidChecksumNode struct{ vnet.InOutNode }

func (node *inputValidChecksumNode) NodeInput(in *vnet.RefIn, out *vnet.RefOut) {
	node.Redirect(in, out, input_next_punt)
}
