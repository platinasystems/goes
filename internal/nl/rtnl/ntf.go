// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

const (
	NTF_USE uint8 = iota
	NTF_SELF
	NTF_MASTER
	NTF_PROXY
	NTF_EXT_LEARNED
	NTF_ROUTER
)

var NtfName = map[uint8]string{
	NTF_USE:         "use",
	NTF_SELF:        "self",
	NTF_MASTER:      "master",
	NTF_PROXY:       "proxy",
	NTF_EXT_LEARNED: "learned",
	NTF_ROUTER:      "router",
}
