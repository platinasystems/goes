// +build amd64 arm

package vnet

func swap16(x uint16) uint16 { return x<<8 | x>>8 }

func swap32(x uint32) uint32 {
	c8 := uint32(0x00ff00ff)
	a := x >> 8 & c8
	b := (x & c8) << 8
	x = a | b
	c16 := uint32(0x0000ffff)
	a = x >> 16 & c16
	b = (x & c16) << 16
	x = a | b
	return x
}

func swap64(x uint64) uint64 {
	c8 := uint64(0x00ff00ff00ff00ff)
	a := x >> 8 & c8
	b := (x & c8) << 8
	x = a | b
	c16 := uint64(0x0000ffff0000ffff)
	a = x >> 16 & c16
	b = (x & c16) << 16
	x = a | b
	c32 := uint64(0x00000000ffffffff)
	a = x >> 32 & c32
	b = (x & c32) << 32
	x = a | b
	return x
}
