// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"

	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

func (m *mod) parseTypeIpoib() error {
	m.args = m.opt.Parms.More(m.args,
		"pkey",
		"mode", // { datagram | connected }
		"umcast",
	)
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"pkey", rtnl.IFLA_IPOIB_PKEY},
		{"umcast", rtnl.IFLA_IPOIB_UMCAST},
	} {
		var u16 uint16
		s := m.opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		if _, err := fmt.Sscan(s, &u16); err != nil {
			return fmt.Errorf("type ipoib %s: %q %v",
				x.name, s, err)
		}
		m.tinfo = append(m.tinfo, nl.Attr{x.t, nl.Uint16Attr(u16)})
	}
	if s := m.opt.Parms.ByName["mode"]; len(s) > 0 {
		if mode, found := map[string]uint16{
			"datagram":  rtnl.IPOIB_MODE_DATAGRAM,
			"connected": rtnl.IPOIB_MODE_CONNECTED,
		}[s]; !found {
			return fmt.Errorf("type ipoib mode: %q invalid", s)
		} else {
			m.tinfo = append(m.tinfo, nl.Attr{rtnl.IFLA_IPOIB_MODE,
				nl.Uint16Attr(mode)})
		}
	}
	return nil
}
