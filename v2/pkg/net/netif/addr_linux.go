// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

import (
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/host"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/universal"
)

type SockaddrIn struct {
	Family *host.Uint16
	Port   *big.Uint16
	Addr   *universal.IPv4
}

func (sain *SockaddrIn) Write(b []byte) (int, error) {
	var data []byte
	sain.Family, data = host.NewUint16(b)
	sain.Port, data = big.NewUint16(data)
	sain.Addr, data = universal.NewIPv4(data)
	return len(data) - len(b), nil
}
