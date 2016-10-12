package mpls

import (
	"github.com/platinasystems/go/vnet"

	"unsafe"
)

// MPLS header:
//   [31:12] 20 bit label
//   [11:9] traffic class (EXP)
//   [8] bottom of stack (BOS) bit
//   [7:0] time to live (TTL)
type Header [4]uint8

type Label uint32

func (h *Header) AsUint32() vnet.Uint32    { return *(*vnet.Uint32)(unsafe.Pointer(&h[0])) }
func (h *Header) FromUint32(x vnet.Uint32) { *(*vnet.Uint32)(unsafe.Pointer(&h[0])) = x }
func (h *Header) GetLabel() Label          { return Label(h.AsUint32().ToHost() >> 12) }
func (h *Header) GetTTL() uint8            { return h[3] }
func (h *Header) IsBottomOfStack() bool    { return h[2]&1 != 0 }

// Special labels 0-15
// 16-239 Unassigned.
// 240-255 Reserved for experimental use.
const (
	Ip4ExplicitNullLabel Label = 0
	RouterAlertLabel     Label = 1
	Ip6ExplicitNullLabel Label = 2
	ImplicitNullLabel    Label = 3
	EntropyLabel         Label = 7
	GALLabel             Label = 13
	OAMAlertLabel        Label = 14
	ExtensionLabel       Label = 15
)
