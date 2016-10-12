package ethernet

import (
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"

	"fmt"
)

func (a *Address) String() string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", a[0], a[1], a[2], a[3], a[4], a[5])
}

func (a *Address) Parse(in *parse.Input) {
	var b [3]vnet.Uint16
	switch {
	case in.Parse("%x:%x:%x:%x:%x:%x", &a[0], &a[1], &a[2], &a[3], &a[4], &a[5]):
	case in.Parse("%x.%x.%x", &b[0], &b[1], &b[2]):
		a[0], a[1] = uint8(b[0]>>8), uint8(b[0])
		a[2], a[3] = uint8(b[1]>>8), uint8(b[1])
		a[4], a[5] = uint8(b[2]>>8), uint8(b[2])
	default:
		panic(parse.ErrInput)
	}
}

func (h *Header) String() (s string) {
	return fmt.Sprintf("%s: %s -> %s", h.GetType().String(), h.Src.String(), h.Dst.String())
}

func (h *Header) Parse(in *parse.Input) {
	if !in.ParseLoose("%v: %v -> %v", &h.Type, &h.Src, &h.Dst) {
		panic(parse.ErrInput)
	}
}

func (h *VlanHeader) String() (s string) {
	return fmt.Sprintf("%s: vlan %d", h.GetType().String(), h.Priority_cfi_and_id.ToHost()&0xfff)
}
