// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

// Extended info filters for IFLA_EXT_MASK
const (
	RTEXT_FILTER_VF Uint32Attr = 1 << iota
	RTEXT_FILTER_BRVLAN
	RTEXT_FILTER_BRVLAN_COMPRESSED
	RTEXT_FILTER_SKIP_STATS
)
