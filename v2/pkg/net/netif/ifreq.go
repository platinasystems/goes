// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

import "unsafe"

func (ifr *Ifreq) Pointer() uintptr    { return uintptr(unsafe.Pointer(ifr)) }
func (ifr *In6Ifreq) Pointer() uintptr { return uintptr(unsafe.Pointer(ifr)) }
