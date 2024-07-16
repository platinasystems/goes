// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

import (
	"crypto/rand"
	"net"
)

const HardwareAddrSize = 6

type HardwareAddr []byte

func NewHardwareAddr() HardwareAddr {
	return make(HardwareAddr, HardwareAddrSize, HardwareAddrSize)
}

func (ha HardwareAddr) MarshalText() ([]byte, error) {
	return []byte(ha.String()), nil
}

func (ha HardwareAddr) Rand() error {
	_, err := rand.Read(ha)
	if err == nil {
		ha.Unicast()
	}
	return err
}

func (ha HardwareAddr) String() string {
	return net.HardwareAddr(ha).String()
}

// clear multicast bit
func (ha HardwareAddr) Unicast() { ha[0] &= 0xfe }

func (ha HardwareAddr) UnmarshalText(text []byte) error {
	mac, err := net.ParseMAC(string(text))
	if err == nil {
		copy(ha, mac)
	}
	return err
}
