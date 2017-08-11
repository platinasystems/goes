// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

func (c *Command) parseTypeVrf() error {
	c.args = c.opt.Parms.More(c.args, "table")
	s := c.opt.Parms.ByName["table"]
	if len(s) == 0 {
		return nil
	}
	tbl, found := rtnl.RtTableByName[s]
	if !found {
		_, err := fmt.Sscan(s, &tbl)
		if err != nil {
			return fmt.Errorf("invalid vrf table")
		}
	}
	c.tinfo = append(c.tinfo,
		rtnl.Attr{rtnl.IFLA_VRF_TABLE, rtnl.Uint32Attr(tbl)})
	return nil
}
