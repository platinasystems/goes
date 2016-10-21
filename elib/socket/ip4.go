// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package socket

import (
	"fmt"
	"net"
)

// 16 bit port; ^uint32(0) means no port given.
type IpPort uint32

const (
	NilIpPort IpPort = 0xffffffff
)

func (p *IpPort) Scan(ss fmt.ScanState, verb rune) (err error) {
	*p = NilIpPort
	tok, err := ss.Token(false, func(r rune) bool {
		switch {
		case r >= '0' && r <= '9':
			return true
		}
		return false
	})
	if len(tok) > 0 {
		var x uint32
		_, err = fmt.Sscanf(string(tok), "%d", &x)
		if err != nil {
			return
		}
		if x>>16 != 0 {
			err = fmt.Errorf("out or range: %s", string(tok))
			return
		}
		*p = IpPort(x)
	}
	return
}

type Ip4Address [4]uint8

func (a *Ip4Address) Scan(ss fmt.ScanState, verb rune) (err error) {
	tok, err := ss.Token(false, func(r rune) bool {
		switch {
		case r >= 'a' && r <= 'z':
			return true
		case r >= 'A' && r <= 'Z':
			return true
		case r >= '0' && r <= '9':
			return true
		case r == '.':
			return true
		case r == '-':
			return true
		}
		return false
	})

	as, err := net.LookupHost(string(tok))
	if err != nil {
		return
	}
	for i := range as {
		_, err = fmt.Sscanf(as[i], "%d.%d.%d.%d", &a[0], &a[1], &a[2], &a[3])
		if err == nil {
			return
		}
	}
	return
}

type Ip4Socket struct {
	Address Ip4Address
	Port    IpPort
}

func (s *Ip4Socket) String() string {
	return fmt.Sprintf("%d.%d.%d.%d:%d", s.Address[0], s.Address[1], s.Address[2], s.Address[3], s.Port)
}

func (s *Ip4Socket) Scan(ss fmt.ScanState, verb rune) (err error) {
	r, _, err := ss.ReadRune()
	if err != nil {
		return
	}
	if r == ':' {
		_, err = fmt.Fscanf(ss, "%d", &s.Port)
	} else {
		err = ss.UnreadRune()
		if err != nil {
			return
		}
		_, err = fmt.Fscanf(ss, "%s:%d", &s.Address, &s.Port)
	}
	return
}
