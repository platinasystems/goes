// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import "github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"

func (opt *Options) Afs() []uint8 {
	switch opt.Parms.ByName["-f"] {
	case "inet":
		return []uint8{rtnl.AF_INET}
	case "inet6":
		return []uint8{rtnl.AF_INET6}
	default:
		return []uint8{rtnl.AF_INET, rtnl.AF_INET6}
	}
}
