// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !linux && !darwin

package netif

func (sock Inet) Admin(ifname string, up bool) error {
	return ErrNoAdmin
}

func (sock Inet) Flags(ifname string) (uint16, error) {
	return 0, ErrNoAdmin
}
