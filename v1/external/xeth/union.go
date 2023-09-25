// Copyright © 2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xeth

type DevJoin struct{ Lower, Upper Xid }
type DevQuit struct{ Lower, Upper Xid }

func (lower Xid) join(upper Xid) *DevJoin {
	lowerl := LinkOf(lower)
	upperl := LinkOf(upper)
	if lowerl == nil || upperl == nil {
		return nil
	}
	lowerl.Uppers(upper.List(lowerl.Uppers()))
	upperl.Lowers(lower.List(upperl.Lowers()))
	return &DevJoin{lower, upper}
}

func (lower Xid) quit(upper Xid) *DevQuit {
	lowerl := LinkOf(lower)
	upperl := LinkOf(upper)
	if lowerl == nil || upperl == nil {
		return nil
	}
	lowerl.Uppers(upper.Delist(lowerl.Uppers()))
	upperl.Lowers(lower.Delist(upperl.Lowers()))
	return &DevQuit{lower, upper}
}

func (xid Xid) List(xids []Xid) []Xid {
	for _, entry := range xids {
		if entry == xid {
			return xids
		}
	}
	return append(xids, xid)
}

func (xid Xid) Delist(xids []Xid) []Xid {
	for i, entry := range xids {
		if entry == xid {
			n := len(xids) - 1
			copy(xids[i:], xids[i+1:])
			xids = xids[:n]
			break
		}
	}
	return xids
}
