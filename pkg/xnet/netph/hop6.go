// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import "unsafe"

type HOP6 struct {
	Type uint8
	Len  uint8
}

const SizeofHOP6 = int(unsafe.Sizeof(HOP6{}))
