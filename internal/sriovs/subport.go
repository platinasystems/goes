// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package sriovs

const SubPortShift = 16

const (
	SubPort0 Vf = iota << SubPortShift
	SubPort1
	SubPort2
	SubPort3
)
