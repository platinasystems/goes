package tomahawk

import (
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/sbus"
)

func (t *tomahawk) PhyIDForPort(block sbus.Block, sub m.SubPortIndex) (id m.PhyID, bus m.PhyBusID) {
	if block == BlockXlport0 {
		bus = 6
		switch sub {
		case 0:
			id = 1
		case 2:
			id = 3
		default:
			panic("sub block")
		}
	} else {
		bi := m.PortBlockIndex(block - BlockClport0)
		p := m.PhyID(1 + 4*uint(bi) + uint(sub))
		switch {
		case p <= 24:
			// bus 0: physical ports 1 to 24 -> 1 to 24
			bus = 0
			id = p
		case p <= 40:
			// bus 1: physical ports 25 to 40 -> 1 to 16
			bus = 1
			id = 1 + (p - 25)
		case p <= 64:
			// bus 2: physical ports 41 to 64 -> 1 to 24
			bus = 2
			id = 1 + (p - 41)
		case p <= 88:
			// bus 3: physical ports 65 to 88 -> 1 to 24
			bus = 3
			id = 1 + (p - 65)
		case p <= 104:
			// bus 4: physical ports 89 to 104 -> 1 to 16
			bus = 4
			id = 1 + (p - 89)
		case p <= 128:
			// bus 5: physical ports 105 to 128 -> 1 to 24
			bus = 5
			id = 1 + (p - 105)
		default:
			panic("port")
		}
	}
	return
}
