// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

import (
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/universal"
)

type SockaddrIn struct {
	Len    *universal.Uint8
	Family *universal.Uint8
	Port   *big.Uint16
	Addr   *universal.IPv4
}

func (sain *SockaddrIn) Write(b []byte) (int, error) {
	var data []byte
	sain.Len, data = universal.NewUint8(b)
	sain.Family, data = universal.NewUint8(data)
	sain.Port, data = big.NewUint16(data)
	sain.Addr, data = universal.NewIPv4(data)
	return len(data) - len(b), nil
}
