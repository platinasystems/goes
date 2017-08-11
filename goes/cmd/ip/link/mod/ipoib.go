// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

func (c *Command) parseTypeIpoib() error {
	c.args = c.opt.Parms.More(c.args,
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
		s := c.opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		if _, err := fmt.Sscan(s, &u16); err != nil {
			return fmt.Errorf("type ipoib %s: %q %v",
				x.name, s, err)
		}
		c.tinfo = append(c.tinfo,
			rtnl.Attr{x.t, rtnl.Uint16Attr(u16)})
	}
	if s := c.opt.Parms.ByName["mode"]; len(s) > 0 {
		if mode, found := map[string]uint16{
			"datagram":  rtnl.IPOIB_MODE_DATAGRAM,
			"connected": rtnl.IPOIB_MODE_CONNECTED,
		}[s]; !found {
			return fmt.Errorf("type ipoib mode: %q invalid", s)
		} else {
			c.tinfo = append(c.tinfo,
				rtnl.Attr{rtnl.IFLA_IPOIB_MODE,
					rtnl.Uint16Attr(mode)})
		}
	}
	return nil
}
