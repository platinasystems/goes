// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

const (
	IF_OPER_UNKNOWN uint8 = iota
	IF_OPER_NOTPRESENT
	IF_OPER_DOWN
	IF_OPER_LOWERLAYERDOWN
	IF_OPER_TESTING
	IF_OPER_DORMANT
	IF_OPER_UP
)

var IfOperName = map[uint8]string{
	IF_OPER_UNKNOWN:        "unknown",
	IF_OPER_NOTPRESENT:     "not-present",
	IF_OPER_DOWN:           "down",
	IF_OPER_LOWERLAYERDOWN: "lower-down",
	IF_OPER_TESTING:        "testing",
	IF_OPER_DORMANT:        "dormant",
	IF_OPER_UP:             "up",
}
