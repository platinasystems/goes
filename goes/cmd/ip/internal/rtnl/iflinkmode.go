// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

const (
	IF_LINK_MODE_DEFAULT uint8 = iota
	IF_LINK_MODE_DORMANT
)

var IfLinkModeName = map[uint8]string{
	IF_LINK_MODE_DEFAULT: "default",
	IF_LINK_MODE_DORMANT: "dormant",
}
