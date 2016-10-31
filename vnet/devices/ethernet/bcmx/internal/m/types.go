package m

import (
	"github.com/platinasystems/go/elib"
)

// Unicast or multicast.
type Cast int

const (
	Unicast Cast = iota
	Multicast
	N_cast
)

var castNames = [...]string{
	Unicast:   "unicast",
	Multicast: "multicast",
}

func (c Cast) String() string {
	return elib.Stringer(castNames[:], int(c))
}
