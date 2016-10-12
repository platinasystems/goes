package ip4

import (
	"github.com/platinasystems/go/elib/parse"

	"fmt"
)

func (a *Address) String() string { return fmt.Sprintf("%d.%d.%d.%d", a[0], a[1], a[2], a[3]) }
func (a *Address) Parse(in *parse.Input) {
	if !in.Parse("%d.%d.%d.%d", &a[0], &a[1], &a[2], &a[3]) {
		panic(parse.ErrInput)
	}
}

func (p *Prefix) String() string { return fmt.Sprintf("%s/%d", &p.Address, p.Len) }
func (p *Prefix) Parse(in *parse.Input) {
	if !in.Parse("%v/%d", &p.Address, &p.Len) {
		panic(parse.ErrInput)
	}
}

func (h *Header) String() (s string) {
	s = fmt.Sprintf("%s: %s -> %s", h.Protocol.String(), h.Src.String(), h.Dst.String())
	if h.Ip_version_and_header_length != 0x45 {
		s += fmt.Sprintf(", version: 0x%02x", h.Ip_version_and_header_length)
	}
	if got, want := h.Checksum, h.ComputeChecksum(); got != want {
		s += fmt.Sprintf(", checksum: 0x%04x (should be 0x%04x)", got.ToHost(), want.ToHost())
	}
	return
}

func (h *Header) Parse(in *parse.Input) {
	h.Ip_version_and_header_length = 0x45
	if !in.ParseLoose("%v: %v -> %v", &h.Protocol, &h.Src, &h.Dst) {
		panic(parse.ErrInput)
	}
	return
}
