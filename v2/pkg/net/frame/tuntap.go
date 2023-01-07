// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/host"
)

type TunPIData struct {
	*TunPI
	Data []byte
}

func NewTunPIData(data []byte) TunPIData {
	pi := (*TunPI)(unsafe.Pointer(&data[0]))
	return TunPIData{pi, data[unsafe.Sizeof(*pi):]}
}

// Returns AF_INET or AF_INET6 if the top 4 bits of the first data byte are 4
// or 6 respectively; otherwise 0.
func (pi *TunPIData) AF() (af uint16) {
	switch pi.Data[0] >> 4 {
	case 4:
		af = syscall.AF_INET
	case 6:
		af = syscall.AF_INET6
	}
	return
}

func (pi TunPIData) Format(w fmt.State, verb rune) {
	fmt.Fprintf(w, "tun: %#x", pi.Flags.Value())
	switch proto := pi.Proto.Value(); proto {
	case syscall.AF_INET:
		fmt.Fprint(w, ProtoMark, NewIPv4Data(pi.Data))
	case syscall.AF_INET6:
		fmt.Fprint(w, ProtoMark, NewIPv6Data(pi.Data))
	default:
		fmt.Fprintf(w, ", proto[%#x]", proto)
	}
}

type TunPI struct {
	Flags host.Uint16
	Proto big.Uint16
}
