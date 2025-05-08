// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import (
	"bytes"
	"net"
)

// net.HardwareAddr with Compare and UnmarshalText methods.
type HardwareAddr net.HardwareAddr

func (ha HardwareAddr) Compare(t HardwareAddr) int {
	return bytes.Compare([]byte(ha), []byte(t))
}

func (ha HardwareAddr) String() string {
	return net.HardwareAddr(ha).String()
}

func (ha *HardwareAddr) UnmarshalText(data []byte) error {
	nha, err := net.ParseMAC(string(data))
	*ha = HardwareAddr(nha)
	return err
}
