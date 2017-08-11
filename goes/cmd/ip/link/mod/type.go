// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"strings"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

func (c *Command) parseType(name string) error {
	c.tinfo = c.tinfo[:0]
	kind := rtnl.Attr{rtnl.IFLA_INFO_KIND, rtnl.KstringAttr(name)}
	dt := rtnl.IFLA_INFO_DATA
	if strings.HasSuffix(name, "_slave") {
		dt = rtnl.IFLA_INFO_SLAVE_DATA
	}
	if parse, found := map[string]func() error{
		"vlan":      c.parseTypeVlan,
		"vxlan":     c.parseTypeVxlan,
		"gre":       c.parseTypeGre,
		"gretap":    c.parseTypeGre,
		"ip6gre":    c.parseTypeIp6Gre,
		"ip6gretap": c.parseTypeIp6Gre,
		"ipip":      c.parseTypeGre,
		"sit":       c.parseTypeGre,
		"geneve":    c.parseTypeGeneve,
		"ipoib":     c.parseTypeIpoib,
		"macvlan":   c.parseTypeMacVlan,
		"macvtap":   c.parseTypeMacVlan,
		"hsr":       c.parseTypeHsr,
		"bridge":    c.parseTypeBridge,
		"macsec":    c.parseTypeMacSec,
		"vrf":       c.parseTypeVrf,
	}[name]; found {
		if err := parse(); err != nil {
			return err
		}
	}
	if len(c.tinfo) == 0 {
		c.attrs = append(c.attrs, rtnl.Attr{rtnl.IFLA_LINKINFO, kind})
	} else {
		c.attrs = append(c.attrs,
			rtnl.Attr{rtnl.IFLA_LINKINFO,
				rtnl.Attrs{
					kind,
					rtnl.Attr{dt, c.tinfo},
				},
			},
		)
	}
	return nil
}
