//+build ppc ppc64

package vnet

// No byte swapping required.
func swap16(x uint16) uint16 { return x }
func swap32(x uint32) uint32 { return x }
func swap64(x uint64) uint64 { return x }
