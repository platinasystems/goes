package elib

import (
	"fmt"
	"runtime"
	"unsafe"
)

const (
	supportsUnaligned = runtime.GOARCH == "386" || runtime.GOARCH == "amd64" || runtime.GOARCH == "ppc64" || runtime.GOARCH == "ppc64le" || runtime.GOARCH == "s390x"
	isBig             = runtime.GOARCH == "ppc64" || runtime.GOARCH == "mips64"
)

func _ua16(p unsafe.Pointer, i, j uint) uint16 {
	q := (*[2]byte)(p)
	return uint16(q[i]) << (8 * j)
}

func _ua32(p unsafe.Pointer, i, j uint) uint32 {
	q := (*[4]byte)(p)
	return uint32(q[i]) << (8 * j)
}

func _ua64(p unsafe.Pointer, i, j uint) uint64 {
	q := (*[8]byte)(p)
	return uint64(q[i]) << (8 * j)
}

func UnalignedUint16(p unsafe.Pointer, i uintptr) uint16 {
	p = PointerAdd(p, i)
	if supportsUnaligned {
		return *(*uint16)(p)
	}
	if isBig {
		return _ua16(p, 0, 1) + _ua16(p, 1, 0)
	} else {
		return _ua16(p, 0, 0) + _ua16(p, 1, 1)
	}
}

func UnalignedUint32(p unsafe.Pointer, i uintptr) uint32 {
	p = PointerAdd(p, i)
	if supportsUnaligned {
		return *(*uint32)(p)
	}
	if isBig {
		return _ua32(p, 0, 3) + _ua32(p, 1, 2) + _ua32(p, 2, 1) + _ua32(p, 3, 0)
	} else {
		return _ua32(p, 0, 0) + _ua32(p, 1, 1) + _ua32(p, 2, 2) + _ua32(p, 3, 3)
	}
}

func UnalignedUint64(p unsafe.Pointer, i uintptr) uint64 {
	p = PointerAdd(p, i)
	if supportsUnaligned {
		return *(*uint64)(p)
	}
	if isBig {
		return _ua64(p, 0, 7) + _ua64(p, 1, 6) + _ua64(p, 2, 5) + _ua64(p, 3, 4) +
			_ua64(p, 4, 3) + _ua64(p, 5, 2) + _ua64(p, 6, 1) + _ua64(p, 7, 0)
	} else {
		return _ua64(p, 0, 0) + _ua64(p, 1, 1) + _ua64(p, 2, 2) + _ua64(p, 3, 3) +
			_ua64(p, 4, 4) + _ua64(p, 5, 5) + _ua64(p, 6, 6) + _ua64(p, 7, 7)
	}
}

func PointerAdd(p unsafe.Pointer, i uintptr) unsafe.Pointer { return unsafe.Pointer(uintptr(p) + i) }

func PointerPoison(p unsafe.Pointer, n uintptr) {
	poison := uint64(0xdeadbeefdeadbeef)
	i := uintptr(0)
	for i+2*8 <= n {
		*(*uint64)(PointerAdd(p, i+0*8)) = poison
		*(*uint64)(PointerAdd(p, i+1*8)) = poison
		i += 2 * 8
	}
	for i+1*8 <= n {
		*(*uint64)(PointerAdd(p, i+0*8)) = poison
		i += 1 * 8
	}
	for i+1*4 <= n {
		*(*uint32)(PointerAdd(p, i+0*4)) = uint32(poison)
		i += 1 * 4
	}
	for i+1*2 <= n {
		*(*uint16)(PointerAdd(p, i+0*2)) = uint16(poison)
		i += 1 * 2
	}
	if i < n {
		*(*uint8)(PointerAdd(p, i)) = uint8(poison)
	}
}

type MemorySize uint64

func (s MemorySize) String() (v string) {
	u, c := uint64(1), rune(0)
	switch {
	case s >= 1<<40:
		u, c = 1<<40, 'T'
	case s >= 1<<30:
		u, c = 1<<30, 'G'
	case s >= 1<<20:
		u, c = 1<<20, 'M'
	case s >= 1<<10:
		u, c = 1<<10, 'K'
	}
	x := float64(s) / float64(u)
	if c == 0 {
		v = fmt.Sprintf("%d", s)
	} else {
		if x != float64(int64(x)) {
			v = fmt.Sprintf("%.2f%c", x, c)
		} else {
			v = fmt.Sprintf("%.0f%c", x, c)
		}
	}
	return
}
