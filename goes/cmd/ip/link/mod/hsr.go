// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"
	"syscall"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

// ip link COMMAND type hsr slave1 SLAVE1 slave2 SLAVE2
//	[ subversion BYTE ]
//	[ version VERSION ]
func (m *mod) parseTypeHsr() error {
	var s string
	var u8 int8
	var err error
	m.args = m.opt.Parms.More(m.args,
		"slave1",     // IFNAME
		"slave2",     // IFNAME
		"subversion", // ADDR_BYTE
		"version",    // { 0 | 1 }
	)
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"slave1", rtnl.IFLA_HSR_SLAVE1},
		{"slave2", rtnl.IFLA_HSR_SLAVE2},
	} {
		s = m.opt.Parms.ByName[x.name]
		if len(s) == 0 {
			return fmt.Errorf("missing %s", x.name)
		}
		idx, found := m.ifindexByName[s]
		if !found {
			return fmt.Errorf("%s: %q not found", x.name, s)
		}
		m.tinfo = append(m.tinfo, rtnl.Attr{x.t, rtnl.Uint32Attr(idx)})
	}
	s = m.opt.Parms.ByName["subversion"]
	if len(s) > 0 {
		if _, err = fmt.Sscan(s, &u8); err != nil {
			return fmt.Errorf("subversion: %q %v", s, err)
		}
		m.tinfo = append(m.tinfo,
			rtnl.Attr{rtnl.IFLA_HSR_MULTICAST_SPEC,
				rtnl.Uint8Attr(u8)})
	}
	s = m.opt.Parms.ByName["version"]
	if len(s) > 0 {
		if _, err = fmt.Sscan(s, &u8); err != nil {
			return fmt.Errorf("version: %q %v", s, err)
		}
		if u8 > 1 {
			return fmt.Errorf("version: %q %v", s, syscall.ERANGE)
		}
		m.tinfo = append(m.tinfo,
			rtnl.Attr{rtnl.IFLA_HSR_VERSION,
				rtnl.Uint8Attr(u8)})
	}
	return nil
}
