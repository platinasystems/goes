// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xeth

import (
	"regexp"
)

type Xid uint32
type Xids []Xid

func (xids Xids) Cut(i int) Xids {
	copy(xids[i:], xids[i+1:])
	return xids[:len(xids)-1]
}

func (xids Xids) FilterName(re *regexp.Regexp) Xids {
	for i := 0; i < len(xids); {
		if re.MatchString(LinkOf(xids[i]).IfInfoName()) {
			i += 1
		} else {
			xids = xids.Cut(i)
		}
	}
	return xids
}

func (xids Xids) FilterNetNs(re *regexp.Regexp) Xids {
	for i := 0; i < len(xids); {
		ns := LinkOf(xids[i]).IfInfoNetNs()
		if re.MatchString(ns.String()) {
			i += 1
		} else {
			xids = xids.Cut(i)
		}
	}
	return xids
}
