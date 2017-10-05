// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

func (m *mod) parseTypeVrf() error {
	m.args = m.opt.Parms.More(m.args, "table")
	s := m.opt.Parms.ByName["table"]
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
	m.tinfo = append(m.tinfo,
		rtnl.Attr{rtnl.IFLA_VRF_TABLE, rtnl.Uint32Attr(tbl)})
	return nil
}
