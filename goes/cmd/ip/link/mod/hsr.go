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
func (c *Command) parseTypeHsr() error {
	var s string
	var u8 int8
	var err error
	c.args = c.opt.Parms.More(c.args,
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
		s = c.opt.Parms.ByName[x.name]
		if len(s) == 0 {
			return fmt.Errorf("missing %s", x.name)
		}
		idx, found := c.ifindexByName[s]
		if !found {
			return fmt.Errorf("%s: %q not found", x.name, s)
		}
		c.tinfo = append(c.tinfo, rtnl.Attr{x.t, rtnl.Uint32Attr(idx)})
	}
	s = c.opt.Parms.ByName["subversion"]
	if len(s) > 0 {
		if _, err = fmt.Sscan(s, &u8); err != nil {
			return fmt.Errorf("subversion: %q %v", s, err)
		}
		c.tinfo = append(c.tinfo,
			rtnl.Attr{rtnl.IFLA_HSR_MULTICAST_SPEC,
				rtnl.Uint8Attr(u8)})
	}
	s = c.opt.Parms.ByName["version"]
	if len(s) > 0 {
		if _, err = fmt.Sscan(s, &u8); err != nil {
			return fmt.Errorf("version: %q %v", s, err)
		}
		if u8 > 1 {
			return fmt.Errorf("version: %q %v", s, syscall.ERANGE)
		}
		c.tinfo = append(c.tinfo,
			rtnl.Attr{rtnl.IFLA_HSR_VERSION,
				rtnl.Uint8Attr(u8)})
	}
	return nil
}
