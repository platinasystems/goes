// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package iflink

var IfOperByName = map[string]IfOper{
	"unknown":     IF_OPER_UNKNOWN,
	"not-present": IF_OPER_NOTPRESENT,
	"down":        IF_OPER_DOWN,
	"lower-down":  IF_OPER_LOWERLAYERDOWN,
	"testing":     IF_OPER_TESTING,
	"dormant":     IF_OPER_DORMANT,
	"up":          IF_OPER_UP,
}
