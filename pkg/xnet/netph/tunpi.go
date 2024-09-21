// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

type TunPI struct {
	Flags,
	Proto uint16
}

const TunPISize = 2 + 2
