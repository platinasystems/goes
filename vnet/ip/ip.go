package ip

import (
	"strconv"
)

type Family uint8

const (
	Ip4 Family = iota
	Ip6
	NFamily
)

// Generic ip4/ip6 address: big enough for either.
type Address [16]uint8

type Prefix struct {
	Address
	Len uint32
}

func (p *Prefix) String(m *Main) string {
	return m.AddressStringer(&p.Address) + "/" + strconv.Itoa(int(p.Len))
}
