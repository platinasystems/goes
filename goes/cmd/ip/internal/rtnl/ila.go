// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

const ILA_GENL_NAME = "ila"
const ILA_GENL_VERSION uint8 = 0x1

const (
	ILA_ATTR_UNSPEC        uint16 = iota
	ILA_ATTR_LOCATOR              // u64
	ILA_ATTR_IDENTIFIER           // u64
	ILA_ATTR_LOCATOR_MATCH        // u64
	ILA_ATTR_IFINDEX              // s32
	ILA_ATTR_DIR                  // u32
	ILA_ATTR_PAD                  // ?
	ILA_ATTR_CSUM_MODE            // u8
	N_ILA_ATTR
)

const ILA_ATTR_MAX = N_ILA_ATTR - 1

const (
	ILA_CMD_UNSPEC uint16 = iota
	ILA_CMD_ADD
	ILA_CMD_DEL
	ILA_CMD_GET

	N_ILA_CMD
)

const ILA_CMD_MAX = N_ILA_CMD - 1

const (
	ILA_DIR_IN uint8 = 1 << iota
	ILA_DIR_OUT
)

const (
	ILA_CSUM_ADJUST_TRANSPORT uint8 = iota
	ILA_CSUM_NEUTRAL_MAP
	ILA_CSUM_NO_ACTION
)

var IlaCsumModeByName = map[string]uint8{
	"adj-transport":    ILA_CSUM_ADJUST_TRANSPORT,
	"adjust-transport": ILA_CSUM_ADJUST_TRANSPORT,
	"neutral-map":      ILA_CSUM_NEUTRAL_MAP,
	"no-action":        ILA_CSUM_NO_ACTION,
}
