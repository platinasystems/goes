// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tuntap

import "net"

type Link struct{ net.HardwareAddr }

func (l Link) IsAutogen() bool { return len(l.HardwareAddr) == 0 }

func (l Link) MarshalText() ([]byte, error) { return []byte(l.String()), nil }

func (l *Link) UnmarshalText(text []byte) error {
	ha, err := net.ParseMAC(string(text))
	if err == nil {
		l.HardwareAddr = ha
	}
	return err
}
