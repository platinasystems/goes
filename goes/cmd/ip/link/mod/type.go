// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"strings"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

func (m *mod) parseType(name string) error {
	m.tinfo = m.tinfo[:0]
	kind := rtnl.Attr{rtnl.IFLA_INFO_KIND, rtnl.KstringAttr(name)}
	dt := rtnl.IFLA_INFO_DATA
	if strings.HasSuffix(name, "_slave") {
		dt = rtnl.IFLA_INFO_SLAVE_DATA
	}
	if parse, found := map[string]func() error{
		"vlan":      m.parseTypeVlan,
		"vxlan":     m.parseTypeVxlan,
		"gre":       m.parseTypeGre,
		"gretap":    m.parseTypeGre,
		"ip6gre":    m.parseTypeIp6Gre,
		"ip6gretap": m.parseTypeIp6Gre,
		"ipip":      m.parseTypeGre,
		"sit":       m.parseTypeGre,
		"geneve":    m.parseTypeGeneve,
		"ipoib":     m.parseTypeIpoib,
		"macvlan":   m.parseTypeMacVlan,
		"macvtap":   m.parseTypeMacVlan,
		"hsr":       m.parseTypeHsr,
		"bridge":    m.parseTypeBridge,
		"macsec":    m.parseTypeMacSec,
		"vrf":       m.parseTypeVrf,
	}[name]; found {
		if err := parse(); err != nil {
			return err
		}
	}
	if len(m.tinfo) == 0 {
		m.attrs = append(m.attrs, rtnl.Attr{rtnl.IFLA_LINKINFO, kind})
	} else {
		m.attrs = append(m.attrs,
			rtnl.Attr{rtnl.IFLA_LINKINFO,
				rtnl.Attrs{
					kind,
					rtnl.Attr{dt, m.tinfo},
				},
			},
		)
	}
	return nil
}
