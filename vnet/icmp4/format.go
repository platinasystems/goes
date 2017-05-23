package icmp4

import (
	"github.com/platinasystems/go/elib/parse"
)

func (h *Header) Parse(in *parse.Input) {
	if !in.ParseLoose("%v", &h.Type) {
		panic(parse.ErrInput)
	}
	return
}
